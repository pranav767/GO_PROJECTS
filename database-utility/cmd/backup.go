package cmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinx/database-utility/internal/compress"
	"github.com/jinx/database-utility/internal/db"
	"github.com/jinx/database-utility/internal/notify"
	"github.com/jinx/database-utility/internal/storage"
	"github.com/spf13/cobra"
)

var (
	dbType       string
	host         string
	port         int
	user         string
	password     string
	dbName       string
	outDir       string
	compressFlag bool
	s3Bucket     string
	s3Key        string
	streamFlag   bool
	tables       string
	slackWebhook string
	mode         string
	whereClause  string
)

func init() {
	backupCmd.Flags().StringVar(&dbType, "type", "mysql", "Database type (mysql|postgres|mongodb|sqlite)")
	backupCmd.Flags().StringVar(&host, "host", "localhost", "Database host")
	backupCmd.Flags().IntVar(&port, "port", 0, "Database port")
	backupCmd.Flags().StringVar(&user, "user", "", "Database user")
	backupCmd.Flags().StringVar(&password, "password", "", "Database password")
	backupCmd.Flags().StringVar(&dbName, "dbname", "", "Database name")
	backupCmd.Flags().StringVar(&outDir, "out", "./backups", "Output directory for backups")
	backupCmd.Flags().BoolVar(&compressFlag, "gzip", true, "Compress backup with gzip")
	backupCmd.Flags().StringVar(&s3Bucket, "s3-bucket", "", "S3 bucket to upload to (optional)")
	backupCmd.Flags().StringVar(&s3Key, "s3-key", "", "S3 key/prefix to use when uploading")
	backupCmd.Flags().BoolVar(&streamFlag, "stream", false, "Stream dump output (use with --s3-bucket to upload directly)")
	backupCmd.Flags().StringVar(&tables, "tables", "", "Comma-separated list of tables/collections to include (if supported)")
	backupCmd.Flags().StringVar(&slackWebhook, "slack-webhook", "", "Slack webhook URL to post notifications to (optional)")
	backupCmd.Flags().StringVar(&mode, "mode", "full", "Backup mode: full|incremental|differential")
	backupCmd.Flags().StringVar(&whereClause, "where", "", "Optional WHERE clause to filter rows for incremental backups (DB-specific)")

	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of a database",
	Run: func(cmd *cobra.Command, args []string) {
		// create output directory
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Printf("failed to create out dir: %v\n", err)
			os.Exit(1)
		}

		// validate connection
		if err := db.TestConnection(dbType, host, port, user, password, dbName); err != nil {
			fmt.Printf("connection test failed: %v\n", err)
			if slackWebhook != "" {
				_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup failed: connection test failed for %s: %v", dbName, err))
			}
			os.Exit(1)
		}

		timestamp := time.Now().Format("20060102T150405")
		base := fmt.Sprintf("%s-%s", dbName, timestamp)

		// streaming path: if stream and s3-bucket provided, stream directly
		if streamFlag && s3Bucket != "" {
			dumpCmd, err := db.BuildDumpCmd(dbType, host, port, user, password, dbName, tables, whereClause)
			if err != nil {
				fmt.Printf("failed to build dump command: %v\n", err)
				os.Exit(1)
			}
			// set any required env (e.g., PGPASSWORD)
			if dbType == "postgres" || dbType == "postgresql" {
				dumpCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
			}

			// incremental/differential support (Postgres backup via per-table COPY with WHERE)
			if mode != "full" && dbType == "postgres" || dbType == "postgresql" {
				if whereClause == "" || tables == "" {
					fmt.Printf("incremental mode for postgres requires --tables and --where to be set\n")
					os.Exit(1)
				}
				// create CSV dumps per table
				if err := db.DumpPostgresTablesWithWhere(host, port, user, password, dbName, tables, whereClause, outDir, base); err != nil {
					fmt.Printf("incremental dump failed: %v\n", err)
					if slackWebhook != "" {
						_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Incremental backup failed: %v", err))
					}
					os.Exit(1)
				}
				// compress and optionally upload each file
				for _, t := range strings.Split(tables, ",") {
					t = strings.TrimSpace(t)
					csvPath := filepath.Join(outDir, base+"-"+t+".csv")
					finalName := base + "-" + t + ".csv"
					if compressFlag {
						finalName = finalName + ".gz"
						finalPath := filepath.Join(outDir, finalName)
						if err := compress.GzipFile(csvPath, finalPath); err != nil {
							fmt.Printf("failed to compress %s: %v\n", csvPath, err)
							continue
						}
						_ = os.Remove(csvPath)
						csvPath = finalPath
					}
					fmt.Printf("table backup written to %s\n", csvPath)
					if s3Bucket != "" {
						key := s3Key
						if key == "" {
							key = finalName
						}
						if err := storage.UploadToS3(s3Bucket, key, csvPath); err != nil {
							fmt.Printf("s3 upload failed for %s: %v\n", csvPath, err)
						} else {
							fmt.Printf("uploaded to s3://%s/%s\n", s3Bucket, key)
						}
					}
				}
				if slackWebhook != "" {
					_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Incremental backup completed: %s", base))
				}
				os.Exit(0)
			}

			pr, pw := io.Pipe()
			// gzip writer that writes into pipe writer
			gw := gzip.NewWriter(pw)

			// start dumpCmd with stdout piped to gzip
			outRdr, err := dumpCmd.StdoutPipe()
			if err != nil {
				fmt.Printf("failed to get stdout pipe: %v\n", err)
				os.Exit(1)
			}
			dumpCmd.Stderr = os.Stderr

			if err := dumpCmd.Start(); err != nil {
				fmt.Printf("failed to start dump command: %v\n", err)
				os.Exit(1)
			}

			// goroutine: read from dump stdout -> gzip writer -> pipe writer
			go func() {
				defer pw.Close()
				defer gw.Close()
				if _, err := io.Copy(gw, outRdr); err != nil {
					fmt.Printf("error during gzip stream: %v\n", err)
				}
			}()

			// upload stream to S3 (reads from pipe reader)
			key := s3Key
			if key == "" {
				key = base + ".sql.gz"
			}
			if err := storage.UploadStreamToS3(s3Bucket, key, pr); err != nil {
				fmt.Printf("stream upload to s3 failed: %v\n", err)
				if slackWebhook != "" {
					_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup failed: stream upload to s3 failed: %v", err))
				}
				os.Exit(1)
			}
			// wait for dumpCmd to finish
			if err := dumpCmd.Wait(); err != nil {
				fmt.Printf("dump command failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("uploaded stream to s3://%s/%s\n", s3Bucket, key)
			if slackWebhook != "" {
				_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup succeeded: s3://%s/%s", s3Bucket, key))
			}
			os.Exit(0)
		}

		// produce initial dump to a temp file
		tmpPath := filepath.Join(outDir, base+".dump")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Printf("failed to create out dir: %v\n", err)
			os.Exit(1)
		}

		if err := db.RunFullBackup(dbType, host, port, user, password, dbName, tmpPath); err != nil {
			fmt.Printf("backup failed: %v\n", err)
			os.Exit(1)
		}

		finalName := base + ".sql"
		finalPath := filepath.Join(outDir, finalName)
		if compressFlag {
			finalName = finalName + ".gz"
			finalPath = filepath.Join(outDir, finalName)
			if err := compress.GzipFile(tmpPath, finalPath); err != nil {
				fmt.Printf("failed to compress: %v\n", err)
			}
			// remove tmp
			_ = os.Remove(tmpPath)
		} else {
			_ = os.Rename(tmpPath, finalPath)
		}

		fmt.Printf("backup written to %s\n", finalPath)

		if s3Bucket != "" {
			key := s3Key
			if key == "" {
				key = finalName
			}
			if err := storage.UploadToS3(s3Bucket, key, finalPath); err != nil {
				fmt.Printf("s3 upload failed: %v\n", err)
				if slackWebhook != "" {
					_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup failed: upload to s3 failed: %v", err))
				}
			} else {
				fmt.Printf("uploaded to s3://%s/%s\n", s3Bucket, key)
				if slackWebhook != "" {
					_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup succeeded: s3://%s/%s", s3Bucket, key))
				}
			}
		} else {
			if slackWebhook != "" {
				_ = notify.SlackWebhook(slackWebhook, fmt.Sprintf("Backup succeeded: %s", finalPath))
			}
		}
	},
}
