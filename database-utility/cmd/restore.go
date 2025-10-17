package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"database-utility/internal/db"
	"database-utility/internal/logger"
	"github.com/spf13/cobra"
)

func init() {
	restoreCmd.Flags().StringVar(&outDir, "in", "./backups", "Input directory containing backup file")
	restoreCmd.Flags().StringVar(&dbType, "type", "mysql", "Database type (mysql|postgres|mongodb|sqlite)")
	restoreCmd.Flags().StringVar(&host, "host", "localhost", "Database host")
	restoreCmd.Flags().IntVar(&port, "port", 0, "Database port")
	restoreCmd.Flags().StringVar(&user, "user", "", "Database user")
	restoreCmd.Flags().StringVar(&password, "password", "", "Database password")
	restoreCmd.Flags().StringVar(&dbName, "dbname", "", "Database name")
	restoreCmd.Flags().StringVar(&tables, "tables", "", "Comma-separated list of tables to restore (for CSV restores)")
	restoreCmd.Flags().StringVar(&mode, "mode", "full", "Restore mode: full|selective")
	restoreCmd.Flags().Bool("driver-restore", false, "Use driver-based restore (no external psql) for Postgres CSVs")
	restoreCmd.Flags().String("upsert-keys", "", "Comma-separated column names to use as upsert keys when using driver restore")
	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore a database from a backup file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("backup file does not exist: %s\n", path)
			os.Exit(1)
		}
		logger.Info("starting restore")
		// if mode is selective and postgres and path is csv or csv.gz and tables specified, use CSV restore
		if mode == "selective" && (dbType == "postgres" || dbType == "postgresql") && tables != "" {
			useDriver, _ := cmd.Flags().GetBool("driver-restore")
			upsertKeys, _ := cmd.Flags().GetString("upsert-keys")
			if useDriver {
				// if path is a directory, iterate per-table csvs
				info, err := os.Stat(path)
				if err != nil {
					logger.Error("restore failed: " + err.Error())
					fmt.Printf("restore failed: %v\n", err)
					os.Exit(1)
				}
				if info.IsDir() {
					for _, t := range splitAndTrim(tables) {
						// find file for table
						candidate := filepath.Join(path, t+".csv")
						if _, err := os.Stat(candidate); os.IsNotExist(err) {
							candidate = filepath.Join(path, t+".csv.gz")
						}
						if _, err := os.Stat(candidate); os.IsNotExist(err) {
							logger.Error("restore failed: missing csv for table " + t)
							fmt.Printf("restore failed: missing csv for table %s\n", t)
							os.Exit(1)
						}
						if err := db.RestorePostgresCSVDriver(host, port, user, password, dbName, candidate, t, upsertKeys); err != nil {
							logger.Error("restore failed: " + err.Error())
							fmt.Printf("restore failed: %v\n", err)
							os.Exit(1)
						}
						fmt.Printf("restored table %s from %s\n", t, candidate)
					}
				} else {
					// single file: restore into each specified table
					for _, t := range splitAndTrim(tables) {
						if err := db.RestorePostgresCSVDriver(host, port, user, password, dbName, path, t, upsertKeys); err != nil {
							logger.Error("restore failed: " + err.Error())
							fmt.Printf("restore failed: %v\n", err)
							os.Exit(1)
						}
						fmt.Printf("restored table %s from %s\n", t, path)
					}
				}
			} else {
				if err := db.RestorePostgresCSV(host, port, user, password, dbName, path, tables); err != nil {
					logger.Error("restore failed: " + err.Error())
					fmt.Printf("restore failed: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			if err := db.RunRestore(dbType, host, port, user, password, dbName, path); err != nil {
				logger.Error("restore failed: " + err.Error())
				fmt.Printf("restore failed: %v\n", err)
				os.Exit(1)
			}
		}
		logger.Info("restore completed")
		fmt.Printf("restore completed from %s\n", path)
	},
}

// splitAndTrim splits a comma-separated list and trims whitespace
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
