package store

import (
	"context"

	"github.com/shurco/mycart/internal/goosemigration/queries"
	"github.com/shurco/mycart/internal/models"
)

// IsInstalled checks if the application has been installed.
func IsInstalled(ctx context.Context) (bool, error) {
	return queries.DB().InstallQueries.IsInstalled(ctx)
}

// Install performs the initial application installation.
func Install(ctx context.Context, request *models.Install) error {
	return queries.DB().InstallQueries.Install(ctx, request)
}
