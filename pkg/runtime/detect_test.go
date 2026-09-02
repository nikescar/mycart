package runtime

import (
	"os"
	"testing"
)

func TestIsCloudflare(t *testing.T) {
	tests := []struct {
		name     string
		runtime  string
		cfAppID  string
		expected bool
	}{
		{
			name:     "cloudflare env detected",
			runtime:  "",
			cfAppID:  "app-123",
			expected: true,
		},
		{
			name:     "explicit runtime override",
			runtime:  "cloudflare",
			cfAppID:  "",
			expected: true,
		},
		{
			name:     "local environment",
			runtime:  "",
			cfAppID:  "",
			expected: false,
		},
		{
			name:     "explicit override wins",
			runtime:  "cloudflare",
			cfAppID:  "app-123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			if tt.runtime != "" {
				os.Setenv("RUNTIME", tt.runtime)
			}
			if tt.cfAppID != "" {
				os.Setenv("CLOUDFLARE_APPLICATION_ID", tt.cfAppID)
			}

			got := IsCloudflare()
			if got != tt.expected {
				t.Errorf("IsCloudflare() = %v, want %v", got, tt.expected)
			}
		})
	}
}
