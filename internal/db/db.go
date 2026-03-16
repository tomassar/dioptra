package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool     *pgxpool.Pool
	readOnly bool
}

type TableInfo struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	Rows   int64  `json:"rows"`
}

type QueryResult struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

func Connect(ctx context.Context, localPort int, user, password, dbname string, readOnly bool) (*DB, error) {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("127.0.0.1:%d", localPort),
		Path:     dbname,
		RawQuery: "sslmode=disable",
	}

	cfg, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	if readOnly {
		cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, "SET default_transaction_read_only = on")
			return err
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{pool: pool, readOnly: readOnly}, nil
}

func (d *DB) Close()         { d.pool.Close() }
func (d *DB) ReadOnly() bool { return d.readOnly }

func (d *DB) Schemas(ctx context.Context) ([]string, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT schema_name FROM information_schema.schemata
		 WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		 ORDER BY schema_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

func (d *DB) Tables(ctx context.Context, schema string) ([]TableInfo, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT schemaname, relname, COALESCE(n_live_tup, 0)
		 FROM pg_stat_user_tables
		 WHERE schemaname = $1
		 ORDER BY relname`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Schema, &t.Name, &t.Rows); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *DB) TableData(ctx context.Context, schema, table string, page, pageSize int) (*QueryResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(
		`SELECT * FROM %s.%s LIMIT $1 OFFSET $2`,
		quoteIdent(schema), quoteIdent(table),
	)

	rows, err := d.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]string, len(rows.FieldDescriptions()))
	for i, fd := range rows.FieldDescriptions() {
		cols[i] = fd.Name
	}

	var result [][]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		result = append(result, vals)
	}
	return &QueryResult{Columns: cols, Rows: result}, rows.Err()
}

func (d *DB) TableCount(ctx context.Context, schema, table string) (int64, error) {
	var count int64
	err := d.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s`, quoteIdent(schema), quoteIdent(table)),
	).Scan(&count)
	return count, err
}

func (d *DB) RunQuery(ctx context.Context, sql string) (*QueryResult, error) {
	rows, err := d.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]string, len(rows.FieldDescriptions()))
	for i, fd := range rows.FieldDescriptions() {
		cols[i] = fd.Name
	}

	var result [][]any
	for i := 0; rows.Next() && i < 1000; i++ {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		result = append(result, vals)
	}
	return &QueryResult{Columns: cols, Rows: result}, rows.Err()
}

// quoteIdent quotes a SQL identifier to prevent injection.
func quoteIdent(s string) string {
	escaped := ""
	for _, c := range s {
		if c == '"' {
			escaped += `""`
		} else {
			escaped += string(c)
		}
	}
	return `"` + escaped + `"`
}
