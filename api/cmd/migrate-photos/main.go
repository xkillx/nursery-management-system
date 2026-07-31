package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/storage"
)

const (
	localPrefix = "uploads/child-photos/"
	s3Prefix    = "children/"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Print actions without uploading or updating")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.S3.AccessKey == "" {
		logger.Error("S3_ACCESS_KEY is required for migration")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s3svc, err := storage.NewS3Service(storage.S3Config{
		Endpoint:   cfg.S3.Endpoint,
		AccessKey:  cfg.S3.AccessKey,
		SecretKey:  cfg.S3.SecretKey,
		BucketName: cfg.S3.BucketName,
		Region:     cfg.S3.Region,
		UseSSL:     cfg.S3.UseSSL,
	})
	if err != nil {
		logger.Error("failed to create S3 service", "error", err)
		os.Exit(1)
	}

	rows, err := pool.Query(ctx, `
		SELECT id, tenant_id, branch_id, profile_photo_path
		FROM children
		WHERE profile_photo_path IS NOT NULL
		  AND profile_photo_path LIKE $1
	`, localPrefix+"%")
	if err != nil {
		logger.Error("failed to query children", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var scanned, migrated, skipped, failed int

	for rows.Next() {
		scanned++

		var id, tenantID, branchID string
		var photoPath string
		if err := rows.Scan(&id, &tenantID, &branchID, &photoPath); err != nil {
			logger.Error("failed to scan row", "error", err, "row", scanned)
			failed++
			continue
		}

		relPath := photoPath
		if !strings.HasPrefix(relPath, localPrefix) {
			logger.Info("skipping non-local path", "child_id", id, "path", photoPath)
			skipped++
			continue
		}

		remainder := strings.TrimPrefix(relPath, localPrefix)
		parts := strings.SplitN(remainder, "/", 2)
		if len(parts) < 2 {
			logger.Warn("unexpected path format, skipping", "child_id", id, "path", photoPath)
			skipped++
			continue
		}

		childDir := parts[0]
		filename := parts[1]
		newKey := s3Prefix + tenantID + "/" + branchID + "/" + childDir + "/" + filename

		if *dryRun {
			logger.Info("[dry-run] would migrate",
				"child_id", id,
				"from", relPath,
				"to", newKey,
			)
			migrated++
			continue
		}

		fileData, err := os.ReadFile(filepath.Join(".", relPath))
		if err != nil {
			logger.Warn("failed to read local file, skipping",
				"child_id", id,
				"path", relPath,
				"error", err,
			)
			failed++
			continue
		}

		ext := strings.TrimPrefix(filepath.Ext(filename), ".")
		contentType := "application/octet-stream"
		switch ext {
		case "jpg", "jpeg":
			contentType = "image/jpeg"
		case "png":
			contentType = "image/png"
		}

		if err := s3svc.Upload(ctx, newKey, fileData, contentType); err != nil {
			logger.Error("failed to upload to S3",
				"child_id", id,
				"key", newKey,
				"error", err,
			)
			failed++
			continue
		}

		_, err = pool.Exec(ctx, `
			UPDATE children
			SET profile_photo_path = $1, updated_at = NOW()
			WHERE id = $2
		`, newKey, id)
		if err != nil {
			logger.Error("failed to update database",
				"child_id", id,
				"new_key", newKey,
				"error", err,
			)
			failed++
			continue
		}

		migrated++
		logger.Info("migrated", "child_id", id, "from", relPath, "to", newKey)
	}

	if err := rows.Err(); err != nil {
		logger.Error("row iteration error", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Migration complete: scanned=%d migrated=%d skipped=%d failed=%d\n", scanned, migrated, skipped, failed)

	if *dryRun {
		fmt.Println("(dry-run mode — no changes were made)")
	}
}
