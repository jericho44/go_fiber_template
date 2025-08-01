package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"
	"go-fiber-template/internal/seeders"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seed",
	Short: "Database seeder tool for Go Fiber Template",
	Long:  `A CLI tool for managing database seeders including run, rollback, and status operations.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Initialize seeder registry
	registry := seeders.NewRegistry(db, cfg)
	manager := registry.GetManager()

	// Add commands
	rootCmd.AddCommand(newRunCommand(manager, registry))
	rootCmd.AddCommand(newRollbackCommand(manager))
	rootCmd.AddCommand(newStatusCommand(manager))
	rootCmd.AddCommand(newListCommand(manager))
}

// runCommand handles seeder run operations
func newRunCommand(manager seeders.SeederManager, registry *seeders.Registry) *cobra.Command {
	var (
		all     bool
		names   []string
		envOnly bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run database seeders",
		Long:  `Run all seeders, specific seeders, or environment-appropriate seeders.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if all {
				fmt.Println("Running all seeders...")
				return manager.RunAll(ctx)
			}

			if envOnly {
				envSeeders := registry.GetSeedersByEnvironment()
				fmt.Printf("Running environment-specific seeders: %s\n", strings.Join(envSeeders, ", "))
				return manager.RunSeeders(ctx, envSeeders)
			}

			if len(names) > 0 {
				fmt.Printf("Running specific seeders: %s\n", strings.Join(names, ", "))
				return manager.RunSeeders(ctx, names)
			}

			if len(args) > 0 {
				fmt.Printf("Running seeders: %s\n", strings.Join(args, ", "))
				return manager.RunSeeders(ctx, args)
			}

			return fmt.Errorf("specify seeders to run using --all, --env, --seeders, or provide seeder names as arguments")
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Run all registered seeders")
	cmd.Flags().StringSliceVarP(&names, "seeders", "s", []string{}, "Comma-separated list of seeder names to run")
	cmd.Flags().BoolVarP(&envOnly, "env", "e", false, "Run only environment-appropriate seeders")

	return cmd
}

// rollbackCommand handles seeder rollback operations
func newRollbackCommand(manager seeders.SeederManager) *cobra.Command {
	var (
		all   bool
		names []string
	)

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback database seeders",
		Long:  `Rollback all seeders or specific seeders.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if all {
				fmt.Println("Rolling back all seeders...")
				return manager.RollbackAll(ctx)
			}

			if len(names) > 0 {
				fmt.Printf("Rolling back specific seeders: %s\n", strings.Join(names, ", "))
				for _, name := range names {
					if err := manager.RollbackSeeder(ctx, name); err != nil {
						return fmt.Errorf("failed to rollback seeder '%s': %w", name, err)
					}
				}
				return nil
			}

			if len(args) > 0 {
				fmt.Printf("Rolling back seeders: %s\n", strings.Join(args, ", "))
				for _, name := range args {
					if err := manager.RollbackSeeder(ctx, name); err != nil {
						return fmt.Errorf("failed to rollback seeder '%s': %w", name, err)
					}
				}
				return nil
			}

			return fmt.Errorf("specify seeders to rollback using --all, --seeders, or provide seeder names as arguments")
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Rollback all seeders")
	cmd.Flags().StringSliceVarP(&names, "seeders", "s", []string{}, "Comma-separated list of seeder names to rollback")

	return cmd
}

// statusCommand shows seeder status
func newStatusCommand(manager seeders.SeederManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show seeder execution status",
		Long:  `Show the execution status of all registered seeders.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			statuses, err := manager.GetSeederStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get seeder status: %w", err)
			}

			if len(statuses) == 0 {
				fmt.Println("No seeders registered.")
				return nil
			}

			fmt.Printf("%-25s %-10s %-20s %s\n", "SEEDER", "STATUS", "LAST RUN", "DESCRIPTION")
			fmt.Println(strings.Repeat("-", 80))

			for _, status := range statuses {
				statusStr := "Not Run"
				lastRun := "Never"

				if status.HasRun {
					statusStr = "Completed"
					if status.LastRunAt != nil {
						lastRun = time.Unix(*status.LastRunAt, 0).Format("2006-01-02 15:04:05")
					}
				}

				if status.Error != "" {
					statusStr = "Error"
				}

				fmt.Printf("%-25s %-10s %-20s %s\n",
					status.Name,
					statusStr,
					lastRun,
					status.Description)
			}

			return nil
		},
	}

	return cmd
}

// listCommand lists all available seeders
func newListCommand(manager seeders.SeederManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available seeders",
		Long:  `List all registered seeders with their descriptions and dependencies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			infos := manager.ListSeeders()

			if len(infos) == 0 {
				fmt.Println("No seeders registered.")
				return nil
			}

			fmt.Printf("%-25s %-15s %s\n", "SEEDER", "DEPENDENCIES", "DESCRIPTION")
			fmt.Println(strings.Repeat("-", 80))

			for _, info := range infos {
				deps := "None"
				if len(info.Dependencies) > 0 {
					deps = strings.Join(info.Dependencies, ", ")
				}

				fmt.Printf("%-25s %-15s %s\n",
					info.Name,
					deps,
					info.Description)
			}

			return nil
		},
	}

	return cmd
}
