package handlers

import (
	"context"
	"fmt"

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
				return webutil.StatusBadRequest(c, err.Error())
			}
			log.ErrorStack(err)
			return webutil.StatusInternalServerError(c)
		}
	}

	// Save Cloudflare credentials to .env if provided
	if request.CFAccountID != "" {
		envVars := map[string]string{
			// D1 configuration (used by config.LoadDatabaseConfig)
			"D1_ACCOUNT_ID":  request.CFAccountID,
			"D1_API_TOKEN":   request.CFAPIToken,
			"D1_DATABASE_ID": request.CFD1DatabaseID,
			// R2 configuration (used by config.LoadStorageConfig)
			"R2_ACCOUNT_ID":  request.CFAccountID,
			"R2_BUCKET_NAME": request.CFR2BucketName,
			// Keep CF_ prefixed versions for display/reference
			"CF_ACCOUNT_ID":     request.CFAccountID,
			"CF_API_TOKEN":      request.CFAPIToken,
			"CF_D1_DATABASE_ID": request.CFD1DatabaseID,
			"CF_R2_BUCKET_NAME": request.CFR2BucketName,
		}
		if err := envfile.WriteEnv(envVars); err != nil {
			log.Warn().Err(err).Msg("Failed to save CF credentials to .env")
			// Continue anyway - user can set manually
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Cart installed", nil)
}

// initializeCloudflareD1 sets up D1 database with migrations and admin user
func initializeCloudflareD1(ctx context.Context, install *models.Install, log *logging.Log) (err error) {
	fmt.Println("DEBUG: initializeCloudflareD1 called")

	// Add panic recovery to catch and log the exact panic location
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("DEBUG: Panic recovered: %v\n", r)
			if log != nil {
				log.Error().Msgf("Panic in initializeCloudflareD1: %v", r)
			}
			err = fmt.Errorf("panic during D1 initialization: %v", r)
		}
	}()

	fmt.Println("DEBUG: After defer setup")

	if install == nil {
		fmt.Println("DEBUG: install is nil!")
		if log != nil {
			log.Error().Msg("install parameter is nil")
		}
		return fmt.Errorf("install parameter is nil")
	}

	fmt.Println("DEBUG: install is not nil")

	if log == nil {
		fmt.Println("DEBUG: log is nil!")
		return fmt.Errorf("log parameter is nil")
	}

	fmt.Println("DEBUG: log is not nil")
	fmt.Printf("DEBUG: install.CFAccountID = %s\n", install.CFAccountID)

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

	log.Info().Msg("Running migrations...")
	// Initialize queries with D1 database and run migrations
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
				fmt.Printf("DEBUG: Panic in Install method: %v\n", r)
				err = fmt.Errorf("panic in Install method: %v", r)
			}
		}()
		fmt.Println("DEBUG: About to call dbInstance.Install")
		err = dbInstance.Install(ctx, install)
		fmt.Println("DEBUG: dbInstance.Install completed without panic")
	}()

	if err != nil {
		if errors.Is(err, queries.ErrAlreadyInstalled) {
			return err
		}
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Info().Msg("Cloudflare D1 initialized successfully")
	return nil
}
