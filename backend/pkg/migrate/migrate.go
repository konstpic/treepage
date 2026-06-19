package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const legacyInitCutoff = 14 // migrations 001–014 were previously applied via docker-entrypoint-initdb.d

// Result summarizes a migration run.
type Result struct {
	Applied []string
	Skipped []string
}

// Run applies pending *_up.sql files from dir in lexical order.
// Applied versions are tracked in schema_migrations; already applied files are skipped.
func Run(ctx context.Context, db *gorm.DB, dir string) (Result, error) {
	var out Result
	if err := ensureSchemaTable(ctx, db); err != nil {
		return out, err
	}
	if err := baselineLegacy(ctx, db, dir); err != nil {
		return out, err
	}

	files, err := listUpFiles(dir)
	if err != nil {
		return out, err
	}

	for _, path := range files {
		version := versionFromPath(path)
		done, err := isRecorded(ctx, db, version)
		if err != nil {
			return out, err
		}
		if done {
			out.Skipped = append(out.Skipped, version)
			continue
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return out, fmt.Errorf("read migration %s: %w", version, err)
		}

		if err := applyMigration(ctx, db, version, string(body)); err != nil {
			return out, fmt.Errorf("apply migration %s: %w", version, err)
		}
		out.Applied = append(out.Applied, version)
	}
	return out, nil
}

func ensureSchemaTable(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`).Error
}

// baselineLegacy marks migrations 001–014 as applied when upgrading an existing DB
// that was initialized before the automatic migrator (docker-entrypoint-initdb.d).
func baselineLegacy(ctx context.Context, db *gorm.DB, dir string) error {
	var count int64
	if err := db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var usersExist bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)`).Scan(&usersExist).Error; err != nil {
		return err
	}
	if !usersExist {
		return nil
	}

	files, err := listUpFiles(dir)
	if err != nil {
		return err
	}
	for _, path := range files {
		version := versionFromPath(path)
		if migrationNumber(version) > legacyInitCutoff {
			continue
		}
		if err := recordMigration(ctx, db, version); err != nil {
			return fmt.Errorf("baseline %s: %w", version, err)
		}
	}
	return nil
}

func applyMigration(ctx context.Context, db *gorm.DB, version, sql string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := execSQL(tx, sql); err != nil {
			return err
		}
		return recordMigration(ctx, tx, version)
	})
}

func execSQL(db *gorm.DB, sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil
	}
	return db.Exec(sql).Error
}

func recordMigration(ctx context.Context, db *gorm.DB, version string) error {
	return db.WithContext(ctx).Exec(
		`INSERT INTO schema_migrations (version) VALUES (?) ON CONFLICT (version) DO NOTHING`,
		version,
	).Error
}

func isRecorded(ctx context.Context, db *gorm.DB, version string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
	).Scan(&count).Error
	return count > 0, err
}

func listUpFiles(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*_up.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func versionFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".up.sql")
}

func migrationNumber(version string) int {
	i := strings.IndexByte(version, '_')
	if i <= 0 {
		i = len(version)
	}
	n, _ := strconv.Atoi(version[:i])
	return n
}

// ResolveDir picks the migrations directory from env or common paths.
func ResolveDir() string {
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		return d
	}
	for _, candidate := range []string{
		"/app/migrations",
		"migrations",
		"../../migrations",
		"../migrations",
	} {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate
		}
	}
	return "migrations"
}
