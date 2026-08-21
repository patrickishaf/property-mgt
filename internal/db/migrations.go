package db

import (
	"fmt"
	"log"
	"strings"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(connectionString, migrationsDir string) error {
	pgxConnString := strings.Replace(connectionString, "postgres://", "pgx5://", 1)

	migrator, err := migrate.New(fmt.Sprintf("file://%s", migrationsDir), pgxConnString)
	if err != nil {
		return err
	}
	defer migrator.Close()

	// Try steps for better debugging
	err = migrator.Up()
	if err == migrate.ErrNoChange {
		log.Println("No changes to apply")
		return nil
	}
	if err != nil {
		log.Printf("Migration error: %v", err)
		return err
	}

	// Verify by checking if schema_migrations table exists
	log.Println("Migrations applied successfully")
	return nil
}

func RevertMigration(connectionString, migrationsDir string) error {
	pgxConnectionString := strings.Replace(connectionString, "postgres://", "pgx5://", 1)

	migrator, err := migrate.New(fmt.Sprintf("file://%s", migrationsDir), pgxConnectionString)
	if err != nil {
		return err
	}
	defer migrator.Close()

	version, _, _ := migrator.Version()

	fmt.Printf("database version: %d\n", version)

	if err := migrator.Migrate(version - 1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("migration reverted")
	return nil
}

func ResetAllMigrations(connectionString, migrationsDir string) error {
	pgxConnectionString := strings.Replace(connectionString, "postgres://", "pgx5://", 1)

	migrator, err := migrate.New(fmt.Sprintf("file://%s", migrationsDir), pgxConnectionString)
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("migration reverted")
	return nil
}
