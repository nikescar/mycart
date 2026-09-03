package queries

import (
	"context"
	"errors"
	"strconv"

	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/pkg/database"
	"github.com/shurco/mycart/pkg/security"
)

// ErrAlreadyInstalled is returned by Install if the cart has already been initialized.
var ErrAlreadyInstalled = errors.New("cart already installed")

// IsInstalled reports whether the cart has completed first-time setup.
func (q *InstallQueries) IsInstalled(ctx context.Context) (bool, error) {
	var rawInstalled string
	if err := q.DB.QueryRow(ctx, `SELECT value FROM setting WHERE key = 'installed'`).Scan(&rawInstalled); err != nil {
		return false, err
	}
	installed, _ := strconv.ParseBool(rawInstalled)
	return installed, nil
}

// InstallQueries is a struct that uses database.Database to provide database functionality for installation operations.
type InstallQueries struct {
	DB database.Database
}

// Install performs the installation process for the cart system.
func (q *InstallQueries) Install(ctx context.Context, i *models.Install) error {
	installed, err := q.IsInstalled(ctx)
	if err != nil {
		return err
	}
	if installed {
		return ErrAlreadyInstalled
	}

	passwordHash := security.GeneratePassword(i.Password)
	jwt_secret, err := security.NewToken(passwordHash)
	if err != nil {
		return err
	}

	settings := map[string]string{
		"installed":  "true",
		"domain":     i.Domain,
		"email":      i.Email,
		"password":   passwordHash,
		"jwt_secret": jwt_secret,
	}

	// D1 doesn't support transactions, so execute updates directly
	// For SQLite, we could use transactions but for consistency we'll use direct updates
	for key, value := range settings {
		_, err := q.DB.Exec(ctx, `UPDATE setting SET value = ? WHERE key = ?`, value, key)
		if err != nil {
			return err
		}
	}

	return nil
}
