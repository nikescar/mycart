package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Install is ...
type Install struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Domain   string `json:"domain"`

	// Cloudflare-specific configuration (optional)
	CFAccountID    string `json:"cf_account_id,omitempty"`
	CFAPIToken     string `json:"cf_api_token,omitempty"`
	CFD1DatabaseID string `json:"cf_d1_database_id,omitempty"`
	CFR2BucketName string `json:"cf_r2_bucket_name,omitempty"`
}

// Validate is ...
func (v Install) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Email, validation.Required, is.Email),
		validation.Field(&v.Password, validation.Required, validation.Length(6, 72)),
	)
}
