package db

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRunMigrations_InvalidPath(t *testing.T) {
	err := RunMigrations("/nonexistent/path", "postgres://user:pass@localhost:5432/test?sslmode=disable")
	if err == nil {
		t.Fatal("expected error for invalid migrations path")
	}
}

func TestRunMigrations_FullyMigrated(t *testing.T) {
	databaseURL, schemaName := setupMigrateTestSchema(t)
	migrateURL := buildMigrateURL(databaseURL, schemaName)
	dir := createTempMigration(t, "000001", "test_table", false)

	if err := RunMigrations(dir, migrateURL); err != nil {
		t.Fatalf("first RunMigrations should apply pending migration: %v", err)
	}

	if err := RunMigrations(dir, migrateURL); err != nil {
		t.Fatalf("second RunMigrations (no-op) should succeed: %v", err)
	}
}

func TestRunMigrations_PendingMigration(t *testing.T) {
	databaseURL, schemaName := setupMigrateTestSchema(t)
	migrateURL := buildMigrateURL(databaseURL, schemaName)

	dir := t.TempDir()
	writeMigration(t, dir, "000001", "CREATE TABLE IF NOT EXISTS t1 (id INT PRIMARY KEY);", "DROP TABLE IF EXISTS t1;")

	if err := RunMigrations(dir, migrateURL); err != nil {
		t.Fatalf("first RunMigrations should apply v1: %v", err)
	}

	writeMigration(t, dir, "000002", "CREATE TABLE IF NOT EXISTS t2 (id INT PRIMARY KEY);", "DROP TABLE IF EXISTS t2;")

	if err := RunMigrations(dir, migrateURL); err != nil {
		t.Fatalf("second RunMigrations should apply v2: %v", err)
	}

	pool := connectPool(t, migrateURL)
	defer pool.Close()
	var tableExists bool
	err := pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = $1 AND table_name = 't2')",
		schemaName).Scan(&tableExists)
	if err != nil {
		t.Fatalf("check t2: %v", err)
	}
	if !tableExists {
		t.Fatal("expected t2 table to exist after migration")
	}
}

func TestRunMigrations_BrokenSQL(t *testing.T) {
	databaseURL, schemaName := setupMigrateTestSchema(t)
	migrateURL := buildMigrateURL(databaseURL, schemaName)
	dir := t.TempDir()

	writeMigration(t, dir, "000001", "THIS IS NOT VALID SQL;", "DROP TABLE IF EXISTS broken;")

	err := RunMigrations(dir, migrateURL)
	if err == nil {
		t.Fatal("expected error for broken SQL migration")
	}
}

func TestRunMigrations_EmptyDir(t *testing.T) {
	databaseURL, schemaName := setupMigrateTestSchema(t)
	migrateURL := buildMigrateURL(databaseURL, schemaName)

	dir := t.TempDir()

	err := RunMigrations(dir, migrateURL)
	if err == nil || !strings.Contains(err.Error(), "no change") {
		t.Fatalf("expected 'no change' error for empty dir, got: %v", err)
	}
}

func setupMigrateTestSchema(t testing.TB) (databaseURL, schemaName string) {
	t.Helper()

	databaseURL = os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration integration test")
	}

	schemaName = fmt.Sprintf("migrate_test_%d", rand.Intn(900000)+100000)

	pool := connectPool(t, databaseURL)
	defer pool.Close()

	_, err := pool.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA %s", pgx.Identifier{schemaName}.Sanitize()))
	if err != nil {
		t.Fatalf("create schema %s: %v", schemaName, err)
	}

	t.Cleanup(func() {
		cleanupPool := connectPool(t, databaseURL)
		defer cleanupPool.Close()
		_, _ = cleanupPool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pgx.Identifier{schemaName}.Sanitize()))
	})

	return databaseURL, schemaName
}

func buildMigrateURL(databaseURL, schemaName string) string {
	u, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		panic(fmt.Sprintf("parse database URL: %v", err))
	}
	q := u.ConnConfig.RuntimeParams
	if q == nil {
		q = map[string]string{}
	}
	return databaseURL + "&search_path=" + schemaName
}

func connectPool(t testing.TB, databaseURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func writeMigration(t testing.TB, dir, version, upSQL, downSQL string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, version+".up.sql"), []byte(upSQL), 0644); err != nil {
		t.Fatalf("write up migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, version+".down.sql"), []byte(downSQL), 0644); err != nil {
		t.Fatalf("write down migration: %v", err)
	}
}

func createTempMigration(t testing.TB, version, tableName string, fail bool) string {
	t.Helper()
	dir := t.TempDir()

	up := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT PRIMARY KEY);", tableName)
	down := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	if fail {
		up = "INVALID SQL;"
	}

	writeMigration(t, dir, version, up, down)
	return dir
}
