package handlers

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/internal/queries"
	"github.com/shurco/mycart/migrations"
	"github.com/shurco/mycart/pkg/database"
	"github.com/shurco/mycart/pkg/envfile"
	"github.com/shurco/mycart/pkg/errors"
	"github.com/shurco/mycart/pkg/logging"
	"github.com/shurco/mycart/pkg/webutil"
)

type installStatus struct {
	Installed bool `json:"installed"`
}

type deploymentType struct {
	Type           string `json:"type"` // "local" or "cloudflare"
	CFAccountID    string `json:"cf_account_id,omitempty"`
	CFAPIToken     string `json:"cf_api_token,omitempty"`
	CFD1DatabaseID string `json:"cf_d1_database_id,omitempty"`
	CFR2BucketName string `json:"cf_r2_bucket_name,omitempty"`
	Installed      bool   `json:"installed"` // whether database is already installed
}

// InstallStatus reports whether first-time setup has been completed.
//
// @Summary      Installation status
// @Description  Returns whether the cart has been installed
// @Tags         Install
// @Produce      json
// @Success      200 {object} webutil.HTTPResponse{result=installStatus}
// @Failure      500 {object} webutil.HTTPResponse "Internal server error"
// @Router       /api/install/status [get]
func InstallStatus(c fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()

	installed, err := db.IsInstalled(c.Context())
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Installation status", installStatus{Installed: installed})
}

// DetectDeployment reports the current deployment configuration from .env
//
// @Summary      Detect deployment type
// @Description  Returns deployment type and credentials only if not yet installed
// @Tags         Install
// @Produce      json
// @Success      200 {object} webutil.HTTPResponse{result=deploymentType}
// @Router       /api/install/detect [get]
func DetectDeployment(c fiber.Ctx) error {
	log := logging.New()
	detection := deploymentType{
		Type: "local",
	}

	// Check for Cloudflare credentials in environment
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	d1DatabaseID := os.Getenv("CLOUDFLARE_D1_DATABASE_ID")

	if accountID != "" && apiToken != "" {
		detection.Type = "cloudflare"

		// If D1 database ID is configured, check if it's already installed
		var isInstalled bool
		if d1DatabaseID != "" {
			var err error
			isInstalled, err = checkD1Installation(accountID, apiToken, d1DatabaseID)
			if err != nil {
				// Log error but don't block - user can still install
				log.Warn().Err(err).Msg("Failed to check D1 installation status")
			}
			detection.Installed = isInstalled
		}

		// Only expose credentials if NOT yet installed (for pre-filling install form)
		// If already installed, backend middleware will redirect before this is called
		if !isInstalled {
			detection.CFAccountID = accountID
			detection.CFAPIToken = apiToken
			detection.CFD1DatabaseID = d1DatabaseID
			detection.CFR2BucketName = os.Getenv("CLOUDFLARE_R2_BUCKET_NAME")
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Deployment configuration", detection)
}

// checkD1Installation attempts to connect to D1 database and check if installed
func checkD1Installation(accountID, apiToken, databaseID string) (bool, error) {
	dbConfig := database.Config{
		Type:       "d1",
		AccountID:  accountID,
		DatabaseID: databaseID,
		APIToken:   apiToken,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return false, fmt.Errorf("connect to D1: %w", err)
	}

	// Check if setting table exists and has installed flag
	var rawInstalled string
	err = db.QueryRow(context.Background(), `SELECT value FROM setting WHERE key = 'installed'`).Scan(&rawInstalled)
	if err != nil {
		// Query failed - check if it's because table doesn't exist or row doesn't exist
		// Try to check if any tables exist in the database
		var tableCount int
		checkErr := db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`).Scan(&tableCount)

		if checkErr == nil && tableCount > 0 {
			// Tables exist but installed flag is missing or query failed
			// This means a previous installation was incomplete - treat as installed to prevent conflicts
			return true, nil
		}

		// No tables exist - not installed
		return false, nil
	}

	installed, _ := strconv.ParseBool(rawInstalled)
	return installed, nil
}

// Install performs the initial installation of the application.
//
// @Summary      Install application
// @Description  Perform initial setup with admin credentials and domain
// @Tags         Install
// @Accept       json
// @Produce      json
// @Param        request body models.Install true "Installation data"
// @Success      200 {object} webutil.HTTPResponse "Cart installed"
// @Failure      400 {object} webutil.HTTPResponse "Validation error"
// @Failure      500 {object} webutil.HTTPResponse "Internal server error"
// @Router       /api/install [post]
func Install(c fiber.Ctx) error {
	log := logging.New()
	request := new(models.Install)

	if err := c.Bind().Body(request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := request.Validate(); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	// Validate Cloudflare configuration consistency
	if (request.CFD1DatabaseID != "" || request.CFR2BucketName != "") && (request.CFAccountID == "" || request.CFAPIToken == "") {
		return webutil.StatusBadRequest(c, "Cloudflare Account ID and API Token are required when using D1 or R2")
	}

	// Initialize Cloudflare D1 if credentials provided
	if request.CFD1DatabaseID != "" && request.CFAccountID != "" {
		if err := initializeCloudflareD1(c.Context(), request, log); err != nil {
			log.ErrorStack(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Cloudflare D1 initialization failed: %v", err),
			})
		}
	} else {
		// Use existing local database
		db := queries.DB()
		if db == nil {
			log.Error().Msg("Database not initialized")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database not initialized. Please restart the server or check database configuration.",
			})
		}
		if err := db.Install(c.Context(), request); err != nil {
			if errors.Is(err, queries.ErrAlreadyInstalled) {
				// Already installed is fine - just continue
				log.Info().Msg("Database already initialized, skipping")
			} else {
				log.ErrorStack(err)
				return webutil.StatusInternalServerError(c)
			}
		}
	}

	// Save Cloudflare credentials to .env if provided
	if request.CFAccountID != "" {
		envVars := map[string]string{
			"CLOUDFLARE_ACCOUNT_ID":     request.CFAccountID,
			"CLOUDFLARE_API_TOKEN":      request.CFAPIToken,
			"CLOUDFLARE_D1_DATABASE_ID": request.CFD1DatabaseID,
			"CLOUDFLARE_R2_BUCKET_NAME": request.CFR2BucketName,
		}

		// Set database and storage types for Cloudflare deployment
		if request.CFD1DatabaseID != "" {
			envVars["DB_TYPE"] = "d1"
		}
		if request.CFR2BucketName != "" {
			envVars["STORAGE_TYPE"] = "r2"
			// R2 credentials are already in envVars (access key and secret key should be provided)
		}

		if err := envfile.WriteEnv(envVars); err != nil {
			log.Warn().Err(err).Msg("Failed to save Cloudflare credentials to .env")
			// Continue anyway - user can set manually
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Cart installed", nil)
}

// initializeCloudflareD1 sets up D1 database with migrations and admin user
func initializeCloudflareD1(ctx context.Context, install *models.Install, log *logging.Log) (err error) {
	// Add panic recovery to catch and log the exact panic location
	defer func() {
		if r := recover(); r != nil {
			if log != nil {
				log.Error().Msgf("Panic in initializeCloudflareD1: %v", r)
			}
			err = fmt.Errorf("panic during D1 initialization: %v", r)
		}
	}()

	if install == nil {
		if log != nil {
			log.Error().Msg("install parameter is nil")
		}
		return fmt.Errorf("install parameter is nil")
	}

	if log == nil {
		return fmt.Errorf("log parameter is nil")
	}

	log.Info().Msg("Starting D1 initialization...")
	log.Info().Msgf("D1 Config - AccountID: %s, DatabaseID: %s", install.CFAccountID, install.CFD1DatabaseID)

	// Create D1 database connection
	dbConfig := database.Config{
		Type:       "d1",
		AccountID:  install.CFAccountID,
		DatabaseID: install.CFD1DatabaseID,
		APIToken:   install.CFAPIToken,
	}

	log.Info().Msg("Creating D1 database connection...")
	db, err := database.New(dbConfig)
	if err != nil {
		return fmt.Errorf("create D1 connection: %w", err)
	}
	if db == nil {
		return fmt.Errorf("database.New returned nil database")
	}

	log.Info().Msg("Running D1 migrations (transaction-free)...")
	// Initialize queries with D1 database
	// The New function will detect D1 and use transaction-free migrations
	if err := queries.New(db, migrations.Embed()); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Info().Msg("Creating admin user...")
	// Create admin user in D1
	dbInstance := queries.DB()
	if dbInstance == nil {
		return fmt.Errorf("database instance is nil after initialization")
	}

	// Add extra protection around Install call
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic in Install method: %v", r)
			}
		}()
		err = dbInstance.Install(ctx, install)
	}()

	if err != nil {
		if errors.Is(err, queries.ErrAlreadyInstalled) {
			// Already installed is fine - just continue
			log.Info().Msg("D1 database already initialized, skipping")
			return nil
		}
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Info().Msg("Cloudflare D1 initialized successfully")
	return nil
}
