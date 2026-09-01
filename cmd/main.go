// @title           myCart API
// @version         1.0
// @description     Open source shopping-cart backend API - a single-binary e-commerce solution
// @termsOfService  https://github.com/shurco/mycart

// @contact.name   API Support
// @contact.url    https://github.com/shurco/mycart/issues
// @contact.email  support@mycart.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	app "github.com/shurco/mycart/internal"
	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/pkg/update"

	_ "github.com/shurco/mycart/docs/swagger"
)

var (
	version   = "v0.0.1"
	gitCommit = "00000000"
	buildDate = "14.07.2023"
)

var rootCmd = &cobra.Command{
	Use:                "mycart",
	Short:              "myCart CLI",
	Long:               "🛒 myCart - shopping-cart in 1 file",
	Version:            fmt.Sprintf("myCart %s (%s) from %s", version, gitCommit, buildDate),
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
}

func main() {
	update.SetVersion(&update.Version{
		CurrentVersion: version,
		GitCommit:      gitCommit,
		BuildDate:      buildDate,
	})

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	rootCmd.AddCommand(cmdInit())
	rootCmd.AddCommand(cmdInstall())
	rootCmd.AddCommand(cmdServe())
	rootCmd.AddCommand(cmdUpdate())
	rootCmd.AddCommand(cmdMigrate())
	rootCmd.AddCommand(cmdMaintenance())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// handleCommandError handles command execution errors uniformly.
func handleCommandError(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

// cmdServe creates and returns the serve command.
func cmdServe() *cobra.Command {
	var noSite, devMode bool
	var httpAddr, httpsAddr string

	cmd := &cobra.Command{
		Use:   "serve [flags]",
		Short: "Starts the web server (default to 0.0.0.0:8080)",
		Run: func(_ *cobra.Command, _ []string) {
			handleCommandError(app.NewApp(httpAddr, httpsAddr, noSite, devMode))
		},
	}

	cmd.PersistentFlags().StringVar(&httpAddr, "http", "0.0.0.0:8080", "server address")
	cmd.PersistentFlags().StringVar(&httpsAddr, "https", "", "https server address (auto TLS)")
	cmd.PersistentFlags().BoolVar(&noSite, "no-site", false, "disable create site")
	cmd.PersistentFlags().BoolVar(&devMode, "dev", false, "develop mode")

	if err := cmd.PersistentFlags().MarkHidden("dev"); err != nil {
		fmt.Println("warning: failed to hide dev flag:", err)
	}

	return cmd
}

// cmdInstall creates and returns the install command.
func cmdInstall() *cobra.Command {
	var email, password, domain string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Create the admin account (first-time setup)",
		Run: func(_ *cobra.Command, _ []string) {
			handleCommandError(app.InstallAdmin(context.Background(), &models.Install{
				Email:    email,
				Password: password,
				Domain:   domain,
			}))
			fmt.Println("Cart installed successfully")
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "admin email address")
	cmd.Flags().StringVar(&password, "password", "", "admin password (6-72 chars)")
	cmd.Flags().StringVar(&domain, "domain", "localhost", "public store domain")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// cmdInit creates and returns the init command.
func cmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Creating the basic structure",
		Run: func(_ *cobra.Command, _ []string) {
			handleCommandError(app.Init())
		},
	}
}

// cmdUpdate creates and returns the update command.
func cmdUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updating the application to the latest version",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &update.Config{
				Owner:             "shurco",
				Repo:              "mycart",
				CurrentVersion:    version,
				ArchiveExecutable: "mycart",
			}

			if err := update.Init(cfg); err != nil {
				handleCommandError(err)
				return
			}

			handleCommandError(app.Migrate())
		},
	}
}

// cmdMigrate creates and returns the migrate command.
func cmdMigrate() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate on the latest version of database schema",
		Run: func(_ *cobra.Command, _ []string) {
			handleCommandError(app.Migrate())
		},
	}
}

// cmdMaintenance creates and returns the maintenance command with subcommands.
func cmdMaintenance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Manage maintenance mode",
	}

	cmd.AddCommand(cmdMaintenanceEnable())
	cmd.AddCommand(cmdMaintenanceDisable())
	cmd.AddCommand(cmdMaintenanceStatus())

	return cmd
}

// cmdMaintenanceEnable creates and returns the enable subcommand.
func cmdMaintenanceEnable() *cobra.Command {
	var restart bool

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable maintenance mode",
		Run: func(_ *cobra.Command, _ []string) {
			// Check if already in maintenance mode
			if _, err := os.Stat(".maintenance"); err == nil {
				fmt.Println("⚠️  Maintenance mode is already enabled")
				return
			}

			// Create .maintenance file with timestamp
			f, err := os.Create(".maintenance")
			if err != nil {
				handleCommandError(fmt.Errorf("failed to enable maintenance mode: %w", err))
				return
			}
			defer f.Close()

			// Write timestamp for audit trail
			timestamp := time.Now().Format(time.RFC3339)
			if _, err := f.WriteString(timestamp); err != nil {
				fmt.Printf("⚠️  Warning: failed to write timestamp: %v\n", err)
			}

			fmt.Println("✓ Maintenance mode enabled")
			fmt.Println("  Only localhost can access the application")
			fmt.Println("  Access maintenance panel at: http://localhost:8080/_/maintenance")

			if restart {
				fmt.Println("\n⟳ Restarting server...")
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}
		},
	}

	cmd.Flags().BoolVar(&restart, "restart", false, "restart server after enabling")
	return cmd
}

// cmdMaintenanceDisable creates and returns the disable subcommand.
func cmdMaintenanceDisable() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable maintenance mode",
		Run: func(_ *cobra.Command, _ []string) {
			if err := os.Remove(".maintenance"); err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Maintenance mode is not enabled")
					return
				}
				handleCommandError(fmt.Errorf("failed to disable maintenance mode: %w", err))
				return
			}

			fmt.Println("✓ Maintenance mode disabled")
			fmt.Println("  Application is now accessible to all users")
		},
	}
}

// cmdMaintenanceStatus creates and returns the status subcommand.
func cmdMaintenanceStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check maintenance mode status",
		Run: func(_ *cobra.Command, _ []string) {
			_, err := os.Stat(".maintenance")
			if err == nil {
				fmt.Println("Status: MAINTENANCE MODE")
				fmt.Println("  Access restricted to allowed IPs only")
			} else {
				fmt.Println("Status: NORMAL")
				fmt.Println("  Application is publicly accessible")
			}
		},
	}
}
