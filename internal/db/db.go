package db

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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

func (d *DB) TableData(ctx context.Context, schema, table string, page, pageSize int, sortCol, sortDir, filterCol, filterVal string) (*QueryResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT * FROM %s.%s`, quoteIdent(schema), quoteIdent(table))

	args := []any{}
	argIdx := 1

	if filterCol != "" && filterVal != "" {
		query += fmt.Sprintf(` WHERE %s::text ILIKE $%d`, quoteIdent(filterCol), argIdx)
		args = append(args, "%"+filterVal+"%")
		argIdx++
	}

	if sortCol != "" {
		dir := "ASC"
		if sortDir == "DESC" || sortDir == "desc" {
			dir = "DESC"
		}
		query += fmt.Sprintf(` ORDER BY %s %s`, quoteIdent(sortCol), dir)
	}

	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := d.pool.Query(ctx, query, args...)
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

func (d *DB) TableCount(ctx context.Context, schema, table, filterCol, filterVal string) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s`, quoteIdent(schema), quoteIdent(table))
	var args []any

	if filterCol != "" && filterVal != "" {
		query += fmt.Sprintf(` WHERE %s::text ILIKE $1`, quoteIdent(filterCol))
		args = append(args, "%"+filterVal+"%")
	}

	var count int64
	err := d.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (d *DB) InsertRow(ctx context.Context, schema, table string, data map[string]string) error {
	if d.readOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	if len(data) == 0 {
		query := fmt.Sprintf(`INSERT INTO %s.%s DEFAULT VALUES`, quoteIdent(schema), quoteIdent(table))
		_, err := d.pool.Exec(ctx, query)
		return err
	}

	var cols []string
	var placeholders []string
	var args []any

	i := 1
	for k, v := range data {
		if v == "" {
			continue // letting db handle defaults/null
		}
		cols = append(cols, quoteIdent(k))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, v)
		i++
	}

	if len(cols) == 0 {
		query := fmt.Sprintf(`INSERT INTO %s.%s DEFAULT VALUES`, quoteIdent(schema), quoteIdent(table))
		_, err := d.pool.Exec(ctx, query)
		return err
	}

	query := fmt.Sprintf(`INSERT INTO %s.%s (%s) VALUES (%s)`,
		quoteIdent(schema), quoteIdent(table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))

	_, err := d.pool.Exec(ctx, query, args...)
	return err
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

func (d *DB) TablePK(ctx context.Context, schema, table string) ([]string, error) {
	query := `
		SELECT a.attname
		FROM   pg_index i
		JOIN   pg_attribute a ON a.attrelid = i.indrelid
							 AND a.attnum = ANY(i.indkey)
		WHERE  i.indrelid = $1::regclass
		AND    i.indisprimary;
	`
	relName := quoteIdent(schema) + "." + quoteIdent(table)
	rows, err := d.pool.Query(ctx, query, relName)
	if err != nil {
		return nil, nil // Return empty, ignore errors for regclass lookup failures
	}
	defer rows.Close()
	var pks []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	return pks, nil
}

func (d *DB) UpdateRow(ctx context.Context, schema, table string, pkValues map[string]any, updates map[string]any) error {
	if d.readOnly {
		return fmt.Errorf("database is in read-only mode")
	}
	if len(pkValues) == 0 {
		return fmt.Errorf("no primary key provided for update")
	}
	if len(updates) == 0 {
		return nil // nothing to update
	}

	setClauses := []string{}
	args := []any{}
	argIdx := 1

	for k, v := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(k), argIdx))
		args = append(args, v)
		argIdx++
	}

	whereClauses := []string{}
	for k, v := range pkValues {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(k), argIdx))
		args = append(args, v)
		argIdx++
	}

	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE %s`, quoteIdent(schema), quoteIdent(table), strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))

	res, err := d.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no rows matched the primary key")
	}
	return nil
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
