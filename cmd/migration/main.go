//go:build !test

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
)

var migrateCmd = &cobra.Command{Use: "migrate"}

func main() {

	if err := migrateCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

const defaultMigrationStep = -1

// upCmd represents the migrate command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run database migrations",
	Long:  `Run database migrations to update the database schema as per defined migration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		configProvider := &config.Config{}
		conf := configProvider.GetConfig()
		log := logger.NewLogger("migration", conf.Log)

		err := postgres.RunMigrations(log)
		if err != nil {
			log.Fatal(logger.Database, logger.MigrationUp, fmt.Sprintf("Migration failed: %v", err), nil)
			return
		}

		log.Info(logger.Database, logger.MigrationUp, "Migration applied successfully", nil)
	},
}

// downCmd represents the migrate command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Revert the last database migration",
	Long:  `Revert the last database migration.`,
	Run: func(cmd *cobra.Command, args []string) {
		configProvider := &config.Config{}
		conf := configProvider.GetConfig()
		log := logger.NewLogger("migration", conf.Log)

		err := postgres.RunDownMigration(log, step)
		if err != nil {
			log.Fatal(logger.Database, logger.MigrationDown, fmt.Sprintf("Migration failed: %v", err), nil)
			return
		}

		log.Info(logger.Database, logger.MigrationDown, "Migration reverted successfully", nil)
	},
}

var step int

func init() {
	downCmd.Flags().IntVarP(&step, "step", "s", defaultMigrationStep, "Number of migrations to revert")
	migrateCmd.AddCommand(upCmd, downCmd)
}
