package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Migrator struct {
	pool *pgxpool.Pool
}

func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{
		pool: pool,
	}
}

func (m *Migrator) Run(ctx context.Context, migrationsPath string) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return err
	}

	files, err := filepath.Glob(
		filepath.Join(migrationsPath, "*.up.sql"),
	)
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		version := migrationVersion(filepath.Base(file))

		applied, err := m.isApplied(ctx, version)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf(
				"failed to read migration %s: %w",
				file,
				err,
			)
		}

		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf(
				"failed to begin transaction: %w",
				err,
			)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)

			return fmt.Errorf(
				"failed to execute migration %s: %w",
				filepath.Base(file),
				err,
			)
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO schema_migrations (version)
			 VALUES ($1)`,
			version,
		)

		if err != nil {
			_ = tx.Rollback(ctx)

			return fmt.Errorf(
				"failed to register migration %s: %w",
				version,
				err,
			)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf(
				"failed to commit migration %s: %w",
				version,
				err,
			)
		}
	}

	return nil
}

func (m *Migrator) createMigrationsTable(
	ctx context.Context,
) error {
	_, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		return fmt.Errorf(
			"failed to create schema_migrations: %w",
			err,
		)
	}

	return nil
}

func (m *Migrator) isApplied(
	ctx context.Context,
	version string,
) (bool, error) {
	var exists bool

	err := m.pool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		)`,
		version,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"failed to check migration %s: %w",
			version,
			err,
		)
	}

	return exists, nil
}

func migrationVersion(filename string) string {
	return strings.SplitN(filename, "_", 2)[0]
}
