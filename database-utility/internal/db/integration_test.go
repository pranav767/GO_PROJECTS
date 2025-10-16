package db

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// TestRestorePostgresCSVDriver runs an integration test using Postgres in Docker.
// It creates a table, inserts initial rows, exports a CSV for the table, modifies the CSV
// to change one row, then uses RestorePostgresCSVDriver with upsert keys to merge changes.
func TestRestorePostgresCSVDriver(t *testing.T) {
	// short-circuit if testing short
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker not available, skipping integration test: %v", err)
	}

	// pull postgres image and run container
	options := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		},
	}
	hostConfig := func(h *docker.HostConfig) {
		h.AutoRemove = true
		h.RestartPolicy = docker.RestartPolicy{Name: "no"}
	}

	resource, err := pool.RunWithOptions(options, hostConfig)
	if err != nil {
		t.Skipf("could not start docker resource, skipping integration test: %v", err)
	}
	// ensure cleanup
	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	// Exponential backoff to wait for container to be ready
	var db *sql.DB
	var dsn string
	var portStr string
	pool.MaxWait = 60 * time.Second
	if err := pool.Retry(func() error {
		portStr = resource.GetPort("5432/tcp")
		dsn = fmt.Sprintf("host=localhost port=%s user=test password=secret dbname=testdb sslmode=disable", portStr)
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		t.Fatalf("could not connect to postgres: %v", err)
	}
	defer db.Close()

	// create table and insert rows
	_, err = db.Exec(`CREATE TABLE users (id text PRIMARY KEY, name text, age integer);`)
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (id, name, age) VALUES ('u1','Alice',30),('u2','Bob',25);`)
	if err != nil {
		t.Fatalf("insert rows failed: %v", err)
	}

	// export CSV via DB SELECT and write to local file
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "users.csv")
	f, ferr := os.Create(csvPath)
	if ferr != nil {
		t.Fatalf("could not create csv file: %v", ferr)
	}
	w := bufio.NewWriter(f)
	_, _ = w.WriteString("id,name,age\n")
	rows2, rerr := db.QueryContext(context.Background(), "SELECT id, name, age FROM users ORDER BY id")
	if rerr != nil {
		f.Close()
		t.Fatalf("could not query rows: %v", rerr)
	}
	for rows2.Next() {
		var id, name string
		var age int
		if err := rows2.Scan(&id, &name, &age); err != nil {
			rows2.Close()
			f.Close()
			t.Fatalf("scan error: %v", err)
		}
		_, _ = w.WriteString(fmt.Sprintf("%s,%s,%d\n", id, name, age))
	}
	rows2.Close()
	w.Flush()
	f.Close()

	// modify CSV to change Bob to Robert and age to 26 (simulate incremental change)
	data, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	content := string(data)
	content = strings.Replace(content, "Bob,25", "Bob,26", 1)
	// and add a new row
	content = content + "u3,Carol,28\n"
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("write modified csv failed: %v", err)
	}

	// truncate table to simulate restore into existing target (so upsert can re-add)
	if _, err := db.Exec("DELETE FROM users WHERE id='u2'"); err != nil {
		t.Fatalf("delete row failed: %v", err)
	}

	// call RestorePostgresCSVDriver with upsert on id
	if err := RestorePostgresCSVDriver("localhost", mustAtoi(portStr), "test", "secret", "testdb", csvPath, "users", "id"); err != nil {
		t.Fatalf("driver restore failed: %v", err)
	}

	// verify rows
	rows, err := db.Query("SELECT id, name, age FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("select after restore failed: %v", err)
	}
	defer rows.Close()
	got := map[string]struct {
		name string
		age  int
	}{}
	for rows.Next() {
		var id, name string
		var age int
		if err := rows.Scan(&id, &name, &age); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		got[id] = struct {
			name string
			age  int
		}{name, age}
	}
	// expected: u1 Alice 30 (unchanged), u2 Bob 26 (restored/upserted), u3 Carol 28 (inserted)
	if v, ok := got["u1"]; !ok || v.name != "Alice" || v.age != 30 {
		t.Fatalf("unexpected u1: %+v", v)
	}
	if v, ok := got["u2"]; !ok || v.name != "Bob" || v.age != 26 {
		t.Fatalf("unexpected u2: %+v", v)
	}
	if v, ok := got["u3"]; !ok || v.name != "Carol" || v.age != 28 {
		t.Fatalf("unexpected u3: %+v", v)
	}
}

func mustAtoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
