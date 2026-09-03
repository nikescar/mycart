package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/d1"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

func init() {
	sql.Register("d1", &d1Driver{})
}

// extractColumnsFromSQL attempts to parse column names from a SELECT statement
// This is a simple parser that handles common cases but may not cover all SQL syntax
func extractColumnsFromSQL(sql string) []string {
	sql = strings.TrimSpace(sql)

	// Find SELECT keyword (case-insensitive)
	selectIdx := strings.Index(strings.ToUpper(sql), "SELECT")
	if selectIdx == -1 {
		return nil
	}

	// Find FROM keyword
	fromIdx := strings.Index(strings.ToUpper(sql[selectIdx:]), " FROM ")
	if fromIdx == -1 {
		return nil
	}
	fromIdx += selectIdx

	// Extract column list between SELECT and FROM
	columnsPart := strings.TrimSpace(sql[selectIdx+6 : fromIdx])

	// Remove DISTINCT if present
	columnsPart = strings.TrimPrefix(strings.TrimSpace(strings.ToUpper(columnsPart)), "DISTINCT")
	columnsPart = strings.TrimSpace(columnsPart)
	// Restore original case
	if strings.HasPrefix(strings.ToUpper(sql[selectIdx+6:fromIdx]), "DISTINCT") {
		distinctEnd := selectIdx + 6 + strings.Index(strings.ToUpper(sql[selectIdx+6:fromIdx]), "DISTINCT") + 8
		columnsPart = strings.TrimSpace(sql[distinctEnd:fromIdx])
	}

	// Handle SELECT *
	if columnsPart == "*" {
		return nil // Can't determine columns from *
	}

	// Split by comma, handling nested function calls
	var columns []string
	var current strings.Builder
	parenDepth := 0

	for _, ch := range columnsPart {
		switch ch {
		case '(':
			parenDepth++
			current.WriteRune(ch)
		case ')':
			parenDepth--
			current.WriteRune(ch)
		case ',':
			if parenDepth == 0 {
				col := parseColumnName(current.String())
				if col != "" {
					columns = append(columns, col)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add last column
	if current.Len() > 0 {
		col := parseColumnName(current.String())
		if col != "" {
			columns = append(columns, col)
		}
	}

	return columns
}

// parseColumnName extracts the actual column name from a column expression
// Handles: "table.column", "column AS alias", "function(column) as alias", etc.
func parseColumnName(expr string) string {
	expr = strings.TrimSpace(expr)

	// Check for AS alias
	asIdx := -1
	upperExpr := strings.ToUpper(expr)
	if idx := strings.LastIndex(upperExpr, " AS "); idx != -1 {
		asIdx = idx
	} else if idx := strings.LastIndex(upperExpr, " "); idx != -1 {
		// Handle implicit alias (column alias without AS keyword)
		// But only if it's not a function call or table.column
		if !strings.Contains(expr[:idx], "(") && !strings.Contains(expr[:idx], ".") {
			asIdx = idx
		}
	}

	var columnExpr string
	if asIdx != -1 {
		// Use the alias as the column name
		alias := strings.TrimSpace(expr[asIdx:])
		alias = strings.TrimPrefix(strings.ToUpper(alias), "AS")
		alias = strings.TrimSpace(alias)
		// Use original case from expr
		aliasStart := asIdx
		if strings.HasPrefix(strings.ToUpper(expr[asIdx:]), " AS ") {
			aliasStart += 4
		} else {
			aliasStart++
		}
		return strings.TrimSpace(expr[aliasStart:])
	}

	columnExpr = expr

	// Remove table prefix (table.column -> column)
	if dotIdx := strings.LastIndex(columnExpr, "."); dotIdx != -1 {
		columnExpr = columnExpr[dotIdx+1:]
	}

	return strings.TrimSpace(columnExpr)
}

// marshalParam converts a driver.Value to a string for D1 API in a type-safe way
func marshalParam(arg driver.Value) (string, error) {
	switch v := arg.(type) {
	case nil:
		return "NULL", nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		if v {
			return "1", nil
		}
		return "0", nil
	case []byte:
		return base64.StdEncoding.EncodeToString(v), nil
	case string:
		return v, nil
	case time.Time:
		return v.Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("unsupported parameter type: %T", v)
	}
}

// d1Driver implements database/sql/driver.Driver
type d1Driver struct{}

// Open returns a new connection to the D1 database
// DSN format: "accountID/databaseID?api_token=xxx"
func (d *d1Driver) Open(dsn string) (driver.Conn, error) {
	parts := strings.Split(dsn, "?")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid DSN format, expected: accountID/databaseID?api_token=xxx")
	}

	ids := strings.Split(parts[0], "/")
	if len(ids) != 2 {
		return nil, fmt.Errorf("invalid DSN format, expected: accountID/databaseID")
	}

	// Parse query params
	params := strings.Split(parts[1], "&")
	var apiToken string
	for _, param := range params {
		if strings.HasPrefix(param, "api_token=") {
			apiToken = strings.TrimPrefix(param, "api_token=")
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("api_token required in DSN")
	}

	client := cloudflare.NewClient(option.WithAPIToken(apiToken))

	return &d1Conn{
		client:     client,
		accountID:  ids[0],
		databaseID: ids[1],
	}, nil
}

// d1Conn implements driver.Conn
type d1Conn struct {
	client     *cloudflare.Client
	accountID  string
	databaseID string
}

func (c *d1Conn) Prepare(query string) (driver.Stmt, error) {
	return &d1DriverStmt{
		conn: c,
		sql:  query,
	}, nil
}

func (c *d1Conn) Close() error {
	return nil
}

func (c *d1Conn) Begin() (driver.Tx, error) {
	// D1 doesn't support real transactions, but we return a no-op transaction
	// that just executes statements directly to maintain API compatibility
	return &d1DriverTx{conn: c}, nil
}

// d1DriverStmt implements driver.Stmt
type d1DriverStmt struct {
	conn  *d1Conn
	sql   string
}

func (s *d1DriverStmt) Close() error {
	return nil
}

func (s *d1DriverStmt) NumInput() int {
	return -1 // variable number of parameters
}

func (s *d1DriverStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.execQuery(context.Background(), args)
}

func (s *d1DriverStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.queryRows(context.Background(), args)
}

func (s *d1DriverStmt) execQuery(ctx context.Context, args []driver.Value) (driver.Result, error) {
	// Convert args to strings with type-safe marshaling
	stringArgs := make([]string, len(args))
	for i, arg := range args {
		marshaled, err := marshalParam(arg)
		if err != nil {
			return nil, fmt.Errorf("marshal parameter %d: %w", i, err)
		}
		stringArgs[i] = marshaled
	}

	queryBody := d1.DatabaseQueryParamsBodyD1SingleQuery{
		Sql:    cloudflare.F(s.sql),
		Params: cloudflare.F(stringArgs),
	}

	params := d1.DatabaseQueryParams{
		AccountID: cloudflare.F(s.conn.accountID),
		Body:      queryBody,
	}

	result, err := s.conn.client.D1.Database.Query(ctx, s.conn.databaseID, params)
	if err != nil {
		return nil, fmt.Errorf("D1 query failed [%s]: %w", s.sql, err)
	}

	rowsAffected := int64(0)
	for _, item := range result.Result {
		if meta := item.Meta; meta.ChangedDB || meta.RowsWritten > 0 {
			rowsAffected += int64(meta.RowsWritten)
		}
	}

	return &d1DriverResult{rowsAffected: rowsAffected}, nil
}

func (s *d1DriverStmt) queryRows(ctx context.Context, args []driver.Value) (driver.Rows, error) {
	// Convert args to strings with type-safe marshaling
	stringArgs := make([]string, len(args))
	for i, arg := range args {
		marshaled, err := marshalParam(arg)
		if err != nil {
			return nil, fmt.Errorf("marshal parameter %d: %w", i, err)
		}
		stringArgs[i] = marshaled
	}

	queryBody := d1.DatabaseQueryParamsBodyD1SingleQuery{
		Sql:    cloudflare.F(s.sql),
		Params: cloudflare.F(stringArgs),
	}

	params := d1.DatabaseQueryParams{
		AccountID: cloudflare.F(s.conn.accountID),
		Body:      queryBody,
	}

	result, err := s.conn.client.D1.Database.Query(ctx, s.conn.databaseID, params)
	if err != nil {
		return nil, fmt.Errorf("D1 query failed [%s]: %w", s.sql, err)
	}

	if len(result.Result) == 0 {
		return &d1DriverRows{columns: []string{}, rows: [][]interface{}{}}, nil
	}

	// Get first result (D1 API returns array of results per query)
	firstResult := result.Result[0]

	// Extract columns and rows from D1 response
	// D1 returns data as []map[string]interface{}
	columns := make([]string, 0)
	rows := make([][]interface{}, 0)

	if len(firstResult.Results) > 0 {
		firstRow, ok := firstResult.Results[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected D1 result format")
		}

		// D1 API returns JSON objects where key order is undefined
		// We MUST use alphabetical sorting for consistent column order
		// NOTE: All SELECT statements must list columns alphabetically to match Scan order
		columns = make([]string, 0, len(firstRow))
		for col := range firstRow {
			columns = append(columns, col)
		}
		sort.Strings(columns)

		// Extract all rows using the determined column order
		// Note: D1 may return column names in different case than the SELECT statement
		// so we need case-insensitive matching
		for _, resultRow := range firstResult.Results {
			rowMap, ok := resultRow.(map[string]interface{})
			if !ok {
				continue
			}

			// Create case-insensitive lookup map
			lowerCaseMap := make(map[string]interface{}, len(rowMap))
			for k, v := range rowMap {
				lowerCaseMap[strings.ToLower(k)] = v
			}

			row := make([]interface{}, len(columns))
			for j, col := range columns {
				// Try exact match first, then case-insensitive
				if val, ok := rowMap[col]; ok {
					row[j] = val
				} else {
					row[j] = lowerCaseMap[strings.ToLower(col)]
				}
			}
			rows = append(rows, row)
		}
	}

	return &d1DriverRows{
		columns: columns,
		rows:    rows,
		index:   0,
	}, nil
}

// d1DriverResult implements driver.Result
type d1DriverResult struct {
	rowsAffected int64
}

func (r *d1DriverResult) LastInsertId() (int64, error) {
	// D1 API does not expose last insert ID in response metadata
	// Workaround: use RETURNING clause in SQLite: INSERT ... RETURNING id
	return 0, fmt.Errorf("D1 does not support LastInsertId - use INSERT ... RETURNING id instead")
}

func (r *d1DriverResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// d1DriverRows implements driver.Rows
type d1DriverRows struct {
	columns []string
	rows    [][]interface{}
	index   int
}

func (r *d1DriverRows) Columns() []string {
	return r.columns
}

func (r *d1DriverRows) Close() error {
	return nil
}

func (r *d1DriverRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}

	row := r.rows[r.index]
	r.index++

	for i, val := range row {
		// Convert JSON numbers to appropriate types
		if num, ok := val.(json.Number); ok {
			// Try int64 first (works for both regular ints and scientific notation ints)
			if n, err := num.Int64(); err == nil {
				dest[i] = n
			} else {
				// Fall back to float64
				f, _ := num.Float64()
				dest[i] = f
			}
		} else if f, ok := val.(float64); ok {
			// D1 may return numbers directly as float64
			// Check if it's actually an integer value within int64 range
			const maxInt64Float = 9223372036854775807.0 // math.MaxInt64 as float64
			const minInt64Float = -9223372036854775808.0 // math.MinInt64 as float64
			if f >= minInt64Float && f <= maxInt64Float && f == float64(int64(f)) {
				dest[i] = int64(f)
			} else {
				dest[i] = f
			}
		} else if str, ok := val.(string); ok {
			// D1 returns booleans as strings "true"/"false" or "1"/"0"
			// Convert to int64 for database/sql compatibility
			switch str {
			case "true", "1":
				dest[i] = int64(1)
			case "false", "0":
				dest[i] = int64(0)
			default:
				// Keep as string - this column is not a boolean type
				// Returning error here would break string columns
				dest[i] = val
			}
		} else {
			dest[i] = val
		}
	}

	return nil
}

// d1DriverTx implements driver.Tx
type d1DriverTx struct {
	conn *d1Conn
}

func (tx *d1DriverTx) Commit() error {
	// D1 auto-commits
	return nil
}

func (tx *d1DriverTx) Rollback() error {
	// D1 doesn't support rollback
	return nil
}
