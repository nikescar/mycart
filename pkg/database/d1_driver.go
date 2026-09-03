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
	return nil, fmt.Errorf("D1 does not support transactions - use batch queries or accept eventual consistency")
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
		// Convert interface{} to map
		firstRow, ok := firstResult.Results[0].(map[string]interface{})
		if ok {
			// Get column names from first row
			// Extract to slice first
			cols := make([]string, 0, len(firstRow))
			for col := range firstRow {
				cols = append(cols, col)
			}
			// Sort for stable ordering (map iteration order is undefined)
			sort.Strings(cols)
			columns = cols

			// Extract all rows
			for _, resultRow := range firstResult.Results {
				rowMap, ok := resultRow.(map[string]interface{})
				if !ok {
					continue
				}
				row := make([]interface{}, len(columns))
				for j, col := range columns {
					row[j] = rowMap[col]
				}
				rows = append(rows, row)
			}
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
