package db

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestConnection attempts to validate credentials for supported DBs
func TestConnection(dbType, host string, port int, user, password, dbName string) error {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch dbType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%s", user, password, host, port, dbName, timeout.String())
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.PingContext(ctx)
	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.PingContext(ctx)
	case "mongodb", "mongo":
		uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", user, password, host, port)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			return err
		}
		defer client.Disconnect(ctx)
		return client.Ping(ctx, nil)
	case "sqlite", "sqlite3":
		// for sqlite just check file exists
		if _, err := os.Stat(dbName); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported db type for connection test: %s", dbType)
	}
}

// RunFullBackup runs a full backup and writes to provided writer (placeholder)
func RunFullBackup(dbType, host string, port int, user, password, dbName string, outPath string) error {
	gz := strings.HasSuffix(outPath, ".gz")
	switch dbType {
	case "mysql":
		// mysqldump -- single SQL
		args := []string{"-h", host, "-P", fmt.Sprint(port), "-u", user, fmt.Sprintf("-p%s", password), dbName}
		cmd := exec.Command("mysqldump", args...)
		if gz {
			return streamCmdToGzip(cmd, outPath)
		}
		// open output file
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()
		// running mysqldump and writing to outPath
		cmd.Stdout = outFile
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("mysqldump failed: %w", err)
		}
		return nil
	case "postgres", "postgresql":
		// use pg_dump
		// pg_dump's password can be provided via PGPASSWORD env
		if gz {
			// stream pg_dump stdout into gzip
			args := []string{"-h", host, "-p", fmt.Sprint(port), "-U", user, dbName}
			cmd := exec.Command("pg_dump", args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
			return streamCmdToGzip(cmd, outPath)
		}
		args := []string{"-h", host, "-p", fmt.Sprint(port), "-U", user, "-F", "c", "-b", "-v", "-f", outPath, dbName}
		cmd := exec.Command("pg_dump", args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pg_dump failed: %w", err)
		}
		return nil
	case "mongodb", "mongo":
		// mongodump -- archive or directory
		if gz {
			args := []string{"--host", host, "--port", fmt.Sprint(port), "--db", dbName, "--archive=-"}
			cmd := exec.Command("mongodump", args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("MONGODB_URI=%s", host))
			return streamCmdToGzip(cmd, outPath)
		}
		args := []string{"--host", host, "--port", fmt.Sprint(port), "--db", dbName, "--archive=" + outPath}
		cmd := exec.Command("mongodump", args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("MONGODB_URI=%s", host))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("mongodump failed: %w", err)
		}
		return nil
	case "sqlite", "sqlite3":
		// sqlite is file-based; copy the file path provided as dbName to outPath
		in, err := os.Open(dbName)
		if err != nil {
			return err
		}
		defer in.Close()
		if gz {
			// compress into gzip
			out, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer out.Close()
			gw := gzip.NewWriter(out)
			defer gw.Close()
			if _, err := io.Copy(gw, in); err != nil {
				return err
			}
			return nil
		}
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported db type: %s", dbType)
	}
}

// streamCmdToGzip runs cmd, reads its stdout and writes gzipped output to outPath
func streamCmdToGzip(cmd *exec.Cmd, outPath string) error {
	rdr, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		// try to kill the process
		_ = cmd.Process.Kill()
		return err
	}
	defer outFile.Close()
	gw := gzip.NewWriter(outFile)
	defer gw.Close()
	if _, err := io.Copy(gw, rdr); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// RunRestore restores a backup file into the target DB using external clients
func RunRestore(dbType, host string, port int, user, password, dbName string, backupPath string) error {
	switch dbType {
	case "sqlite", "sqlite3":
		// overwrite the sqlite file (dbName is path to sqlite file)
		in, err := os.Open(backupPath)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(dbName)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	case "mysql":
		// mysql client: mysql -h host -P port -u user -p password dbname < backup.sql
		cmd := exec.Command("sh", "-c", fmt.Sprintf("mysql -h %s -P %d -u %s -p%s %s < %s", host, port, user, password, dbName, backupPath))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "postgres", "postgresql":
		// pg_restore for custom format or psql for SQL
		// try pg_restore first
		cmd := exec.Command("pg_restore", "-h", host, "-p", fmt.Sprint(port), "-U", user, "-d", dbName, backupPath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		// fallback to psql for plain SQL
		cmd2 := exec.Command("sh", "-c", fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %d -U %s -d %s -f %s", password, host, port, user, dbName, backupPath))
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		return cmd2.Run()
	case "mongodb", "mongo":
		// mongorestore --archive=backupPath --db dbName
		cmd := exec.Command("mongorestore", "--archive="+backupPath, "--nsFrom", "*.*", "--nsTo", "*.*", "--db", dbName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported db type for restore: %s", dbType)
	}
}

// BuildDumpCmd builds an exec.Cmd that when run will write the dump to stdout.
// tables is an optional comma-separated list of tables/collections to include.
func BuildDumpCmd(dbType, host string, port int, user, password, dbName, tables, where string) (*exec.Cmd, error) {
	switch dbType {
	case "mysql":
		args := []string{"-h", host, "-P", fmt.Sprint(port), "-u", user, fmt.Sprintf("-p%s", password)}
		// mysqldump supports --where for filtering rows
		if where != "" {
			args = append(args, "--where="+where)
		}
		if tables != "" {
			// pass tables after dbName
			parts := append(args, dbName)
			for _, t := range strings.Split(tables, ",") {
				parts = append(parts, strings.TrimSpace(t))
			}
			return exec.Command("mysqldump", parts...), nil
		}
		args = append(args, dbName)
		return exec.Command("mysqldump", args...), nil
	case "postgres", "postgresql":
		// pg_dump writes to stdout by default
		args := []string{"-h", host, "-p", fmt.Sprint(port), "-U", user, dbName}
		if tables != "" {
			parts := []string{"-h", host, "-p", fmt.Sprint(port), "-U", user, dbName}
			for _, t := range strings.Split(tables, ",") {
				parts = append(parts, "-t")
				parts = append(parts, strings.TrimSpace(t))
			}
			return exec.Command("pg_dump", parts...), nil
		}
		return exec.Command("pg_dump", args...), nil
	case "mongodb", "mongo":
		// mongodump can output archive to stdout with --archive=-
		args := []string{"--host", host, "--port", fmt.Sprint(port), "--db", dbName, "--archive=-"}
		if tables != "" {
			// Note: mongodump supports --collection for single collection; multiple collections require multiple runs.
			// For simplicity, if tables provided, only use the first as collection.
			first := strings.Split(tables, ",")[0]
			args = []string{"--host", host, "--port", fmt.Sprint(port), "--db", dbName, "--collection", strings.TrimSpace(first), "--archive=-"}
		}
		return exec.Command("mongodump", args...), nil
	default:
		return nil, fmt.Errorf("unsupported db type for BuildDumpCmd: %s", dbType)
	}
}

// DumpPostgresTablesWithWhere exports CSVs per table using COPY (SELECT ...) TO STDOUT
func DumpPostgresTablesWithWhere(host string, port int, user, password, dbName, tables, whereClause, outDir, base string) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	dbconn, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer dbconn.Close()

	for _, t := range strings.Split(tables, ",") {
		t = strings.TrimSpace(t)
		// build COPY query
		query := fmt.Sprintf("COPY (SELECT * FROM %s WHERE %s) TO STDOUT WITH CSV", t, whereClause)
		// open file
		csvPath := filepath.Join(outDir, base+"-"+t+".csv")
		f, err := os.Create(csvPath)
		if err != nil {
			return err
		}
		// use pg_dump via psql -c "COPY (...) TO STDOUT" > file as fallback
		// attempt to use psql COPY via exec to stream directly
		cmdStr := fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %d -U %s -d %s -c \"%s\" > %s", password, host, port, user, dbName, query, csvPath)
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// restoreCSVIntoTable runs psql \copy to import CSV (reader) into the given table
func restoreCSVIntoTable(host string, port int, user, password, dbName string, rdr io.Reader, table string) error {
	copyCmd := fmt.Sprintf("\\copy %s FROM STDIN WITH CSV", table)
	cmd := exec.Command("psql", "-h", host, "-p", fmt.Sprint(port), "-U", user, "-d", dbName, "-c", copyCmd)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	cmd.Stdin = rdr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RestorePostgresCSV restores CSV files into Postgres tables. csvPath can be a file or a directory.
// tables may be a comma-separated list. If csvPath is a directory, the function will search for per-table files.
func RestorePostgresCSV(host string, port int, user, password, dbName, csvPath, tables string) error {
	// prepare table list
	tbls := []string{}
	for _, t := range strings.Split(tables, ",") {
		if tt := strings.TrimSpace(t); tt != "" {
			tbls = append(tbls, tt)
		}
	}
	if len(tbls) == 0 {
		return fmt.Errorf("no tables specified for CSV restore")
	}

	info, err := os.Stat(csvPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// look for per-table CSVs
		for _, table := range tbls {
			var found string
			// try simpler patterns (no space in glob)
			patterns := []string{
				filepath.Join(csvPath, table+".csv"),
				filepath.Join(csvPath, "*-") + table + ".csv",
				filepath.Join(csvPath, table+".csv.gz"),
				filepath.Join(csvPath, "*-") + table + ".csv.gz",
			}
			for _, pat := range patterns {
				matches, _ := filepath.Glob(pat)
				if len(matches) > 0 {
					found = matches[0]
					break
				}
			}
			if found == "" {
				return fmt.Errorf("could not find CSV file for table %s in %s", table, csvPath)
			}
			// open and restore
			f, err := os.Open(found)
			if err != nil {
				return err
			}
			var rdr io.Reader = f
			if strings.HasSuffix(found, ".gz") {
				gr, err := gzip.NewReader(f)
				if err != nil {
					f.Close()
					return err
				}
				defer gr.Close()
				rdr = gr
			}
			if err := restoreCSVIntoTable(host, port, user, password, dbName, rdr, table); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
		return nil
	}

	// csvPath is a file. If multiple tables specified, attempt to restore the file into each table (user responsibility).
	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, table := range tbls {
		// for each table we need a fresh reader; reopen file
		rf, err := os.Open(csvPath)
		if err != nil {
			return err
		}
		var rdr io.Reader = rf
		if strings.HasSuffix(csvPath, ".gz") {
			gr, err := gzip.NewReader(rf)
			if err != nil {
				rf.Close()
				return err
			}
			rdr = gr
			if err := restoreCSVIntoTable(host, port, user, password, dbName, rdr, table); err != nil {
				gr.Close()
				rf.Close()
				return err
			}
			gr.Close()
		} else {
			if err := restoreCSVIntoTable(host, port, user, password, dbName, rdr, table); err != nil {
				rf.Close()
				return err
			}
		}
		rf.Close()
	}
	return nil
}

// RestorePostgresCSVDriver restores CSV into a Postgres table using the driver (no psql).
// If upsertKeys is provided (comma-separated), the function will upsert using ON CONFLICT on those keys.
func RestorePostgresCSVDriver(host string, port int, user, password, dbName, csvPath, table, upsertKeys string) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	dbconn, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer dbconn.Close()

	// Open CSV file
	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(csvPath, ".gz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gr.Close()
		reader = gr
	}

	r := csv.NewReader(reader)
	// read header
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("reading csv header: %w", err)
	}
	cols := make([]string, len(header))
	for i, c := range header {
		cols[i] = strings.TrimSpace(c)
	}

	tx, err := dbconn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	tempTable := fmt.Sprintf("tmp_%s_%d", table, time.Now().Unix())
	// create temp table with all text columns
	colDefs := make([]string, len(cols))
	for i, c := range cols {
		colDefs[i] = pq.QuoteIdentifier(c) + " text"
	}
	createSQL := fmt.Sprintf("CREATE TEMP TABLE %s (%s) ON COMMIT DROP;", pq.QuoteIdentifier(tempTable), strings.Join(colDefs, ","))
	if _, err := tx.Exec(createSQL); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	// prepare COPY IN
	copyStmt := pq.CopyIn(tempTable, cols...)
	stmt, err := tx.Prepare(copyStmt)
	if err != nil {
		return fmt.Errorf("prepare copy in: %w", err)
	}

	// read rows and execute copy
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			stmt.Close()
			return fmt.Errorf("read csv record: %w", err)
		}
		vals := make([]interface{}, len(rec))
		for i := range rec {
			vals[i] = rec[i]
		}
		if _, err := stmt.Exec(vals...); err != nil {
			stmt.Close()
			return fmt.Errorf("copy exec: %w", err)
		}
	}
	if _, err := stmt.Exec(); err != nil {
		stmt.Close()
		return fmt.Errorf("finalize copy: %w", err)
	}
	if err := stmt.Close(); err != nil {
		return fmt.Errorf("close copy stmt: %w", err)
	}

	// Build insert from temp into target, with optional upsert
	targetCols := make([]string, len(cols))
	for i, c := range cols {
		targetCols[i] = pq.QuoteIdentifier(c)
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", pq.QuoteIdentifier(table), strings.Join(targetCols, ","), strings.Join(targetCols, ","), pq.QuoteIdentifier(tempTable))
	if upsertKeys != "" {
		keys := []string{}
		for _, k := range strings.Split(upsertKeys, ",") {
			keys = append(keys, pq.QuoteIdentifier(strings.TrimSpace(k)))
		}
		// build update set for non-key columns
		nonKeys := []string{}
		keySet := map[string]struct{}{}
		for _, k := range keys {
			keySet[k] = struct{}{}
		}
		for _, c := range cols {
			qc := pq.QuoteIdentifier(c)
			if _, ok := keySet[qc]; ok {
				continue
			}
			nonKeys = append(nonKeys, fmt.Sprintf("%s = EXCLUDED.%s", qc, qc))
		}
		insertSQL = fmt.Sprintf("%s ON CONFLICT (%s) DO UPDATE SET %s", insertSQL, strings.Join(keys, ","), strings.Join(nonKeys, ","))
	}

	if _, err := tx.Exec(insertSQL); err != nil {
		return fmt.Errorf("insert from temp: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
