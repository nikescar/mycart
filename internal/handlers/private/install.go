package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/shurco/mycart/internal/handlers/cloudflare"
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
			"CF_ACCOUNT_ID":    request.CFAccountID,
			"CF_API_TOKEN":     request.CFAPIToken,
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
func initializeCloudflareD1(ctx context.Context, install *models.Install, log *logging.Log) error {
	// Create D1 database connection
	dbConfig := database.Config{
		Type:       "d1",
		AccountID:  install.CFAccountID,
		DatabaseID: install.CFD1DatabaseID,
		APIToken:   install.CFAPIToken,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return fmt.Errorf("create D1 connection: %w", err)
	}

	// Initialize queries with D1 database and run migrations
	if err := queries.New(db, migrations.Embed()); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// Create admin user in D1
	if err := queries.DB().Install(ctx, install); err != nil {
		if errors.Is(err, queries.ErrAlreadyInstalled) {
			return err
		}
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Info().Msg("Cloudflare D1 initialized successfully")
	return nil
}
