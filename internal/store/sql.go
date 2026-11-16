package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

const (
	NameSQLite     = "sqlite"
	NamePostgres   = "postgres"
	EnvPostgresURL = "PROVEN_POSTGRES_URL"
)

// SQLStore owns a database connection. Query construction stays in querytpl.
// Register/Replace publish a table and its content identity in one transaction.
type SQLStore struct {
	db     *sql.DB
	engine string
}

func OpenSQLite(path string) (*SQLStore, error) {
	if err := checkFile(path); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	// Create securely before SQLite opens it. A supplied file keeps its permissions.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	f.Close()
	u := url.URL{Scheme: "file", Path: path}
	db, err := sql.Open("sqlite", u.String()+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &SQLStore{db: db, engine: NameSQLite}, nil
}

// PostgreSQL uses the same network policy as HTTP adapters, including local mode.
type policyDialer struct{ p *network.Policy }

func (d policyDialer) Dial(networkName, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), networkName, address)
}
func (d policyDialer) DialTimeout(networkName, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.DialContext(ctx, networkName, address)
}
func (d policyDialer) DialContext(ctx context.Context, networkName, address string) (net.Conn, error) {
	if networkName != "tcp" && networkName != "tcp4" && networkName != "tcp6" {
		return nil, fmt.Errorf("postgres: TCP required")
	}
	return d.p.DialContext(ctx, networkName, address)
}

func OpenPostgres(raw string, policy *network.Policy) (*SQLStore, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.Fragment != "" {
		return nil, refuse(EnvPostgresURL, "", "invalid database URL", "a postgres:// or postgresql:// URL")
	}
	if policy == nil {
		policy, _ = network.Load(nil)
	}
	address := "https://" + u.Host
	if err := policy.CheckURL(EnvPostgresURL, address); err != nil {
		return nil, refuse(EnvPostgresURL, "", "database outside the configured network", "an allowed PostgreSQL host")
	}
	connector, err := pq.NewConnector(raw)
	if err != nil {
		return nil, refuse(EnvPostgresURL, "", "invalid connection settings", "a PostgreSQL connection URL")
	}
	connector.Dialer(policyDialer{p: policy})
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(4)
	return &SQLStore{db: db, engine: NamePostgres}, nil
}

func (s *SQLStore) Close() error { return s.db.Close() }

func (s *SQLStore) Register(ctx context.Context, table Table) error {
	return s.writeTable(ctx, table, false)
}
func (s *SQLStore) Replace(ctx context.Context, table Table) error {
	return s.writeTable(ctx, table, true)
}

func (s *SQLStore) quote(name string) string {
	if s.engine == NameSQLite {
		return `"` + name + `"`
	}
	parts := strings.Split(name, ".")
	for i := range parts {
		parts[i] = `"` + parts[i] + `"`
	}
	return strings.Join(parts, ".")
}
func (s *SQLStore) param(i int) string {
	if s.engine == NamePostgres {
		return fmt.Sprintf("$%d", i)
	}
	return "?"
}

func (s *SQLStore) writeTable(parent context.Context, table Table, replace bool) error {
	if err := checkTable(table); err != nil {
		return err
	}
	if strings.HasPrefix(strings.ToLower(table.Name), "_proven_") {
		return fmt.Errorf("store: reserved table name")
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return s.failed()
	}
	defer tx.Rollback()
	if s.engine == NamePostgres && strings.Contains(table.Name, ".") {
		if _, err := tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+s.quote(strings.Split(table.Name, ".")[0])); err != nil {
			return s.failed()
		}
	}
	if _, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS _proven_snapshots (name TEXT PRIMARY KEY, snapshot TEXT NOT NULL)"); err != nil {
		return s.failed()
	}
	if replace {
		if _, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS "+s.quote(table.Name)); err != nil {
			return s.failed()
		}
	}
	cols, values := make([]string, len(table.Columns)), make([]string, len(table.Columns))
	for i, col := range table.Columns {
		cols[i], values[i] = s.quote(col)+" TEXT NOT NULL", s.param(i+1)
	}
	if _, err = tx.ExecContext(ctx, "CREATE TABLE "+s.quote(table.Name)+" ("+strings.Join(cols, ",")+")"); err != nil {
		return fmt.Errorf("store: could not register table %s (it may already exist)", table.Name)
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO "+s.quote(table.Name)+" VALUES ("+strings.Join(values, ",")+")")
	if err != nil {
		return s.failed()
	}
	defer stmt.Close()
	for _, row := range table.Rows {
		args := make([]any, len(row))
		for i := range row {
			args[i] = row[i]
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return s.failed()
		}
	}
	encoded, _ := json.Marshal(table)
	digest := sha256.Sum256(encoded)
	snapshot := "sha256:" + hex.EncodeToString(digest[:])
	if _, err = tx.ExecContext(ctx, "INSERT INTO _proven_snapshots (name,snapshot) VALUES ("+s.param(1)+","+s.param(2)+") ON CONFLICT(name) DO UPDATE SET snapshot=excluded.snapshot", table.Name, snapshot); err != nil {
		return s.failed()
	}
	if err := tx.Commit(); err != nil {
		return s.failed()
	}
	return nil
}

var sqlRelations = regexp.MustCompile(`(?i)\b(FROM|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)

func (s *SQLStore) Run(parent context.Context, req querytpl.Request, rel querytpl.Relations) (Evidence, error) {
	forms, err := querytpl.BuildForms(req, rel)
	if err != nil {
		return Evidence{}, err
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	// A transaction keeps the source identity and rows in the same read snapshot.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelSerializable})
	if err != nil {
		return Evidence{}, s.failed()
	}
	defer tx.Rollback()
	text := strings.TrimSpace(forms.SQL)
	if !strings.HasPrefix(strings.ToUpper(text), "SELECT ") || strings.Contains(strings.TrimSuffix(text, ";"), ";") {
		return Evidence{}, fmt.Errorf("store: only one registered SELECT is allowed")
	}
	var snapshots []string
	for _, match := range sqlRelations.FindAllStringSubmatch(text, -1) {
		var snapshot string
		if err := tx.QueryRowContext(ctx, "SELECT snapshot FROM _proven_snapshots WHERE name="+s.param(1), match[2]).Scan(&snapshot); err != nil {
			return Evidence{}, &service.Down{Service: s.engine, Reason: "no registered snapshot for " + match[2]}
		}
		snapshots = append(snapshots, match[2]+"="+snapshot)
	}
	statement := sqlRelations.ReplaceAllStringFunc(text, func(m string) string {
		parts := sqlRelations.FindStringSubmatch(m)
		return parts[1] + " " + s.quote(parts[2])
	})
	rows, err := tx.QueryContext(ctx, statement)
	if err != nil {
		return Evidence{}, s.failed()
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return Evidence{}, s.failed()
	}
	result := make([][]string, 0)
	for rows.Next() {
		if len(result) >= 100000 {
			return Evidence{}, &service.Down{Service: s.engine, Reason: "row limit exceeded"}
		}
		cells, ptrs := make([]sql.NullString, len(cols)), make([]any, len(cols))
		for i := range cells {
			ptrs[i] = &cells[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return Evidence{}, s.failed()
		}
		row := make([]string, len(cols))
		for i := range cells {
			row[i] = cells[i].String
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return Evidence{}, s.failed()
	}
	if err := ctx.Err(); err != nil {
		return Evidence{}, err
	}
	return Evidence{Form: querytpl.FormSQL, Query: statement, Engine: s.engine, Snapshot: strings.Join(snapshots, ";"), Columns: cols, Rows: result}, nil
}

func (s *SQLStore) failed() error {
	return &service.Down{Service: s.engine, Reason: "database query failed"}
}
