package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

const (
	pgxName = "postgres"
)

type pgxDriver struct {
	name  string
	dbURL string
	conn  *pgxpool.Pool
}

// Декораторы, чтобы логировать SQL
func (d *pgxDriver) exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	logger.Log.Debug(sql, zap.Any("args", args))
	return d.conn.Exec(ctx, sql, args...)
}

func (d *pgxDriver) queryRows(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	logger.Log.Debug(sql, zap.Any("args", args))
	return d.conn.Query(ctx, sql, args...)
}

func (d *pgxDriver) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	logger.Log.Debug(sql, zap.Any("args", args))
	return d.conn.QueryRow(ctx, sql, args...)
}

func NewPgxDriver(dbURL string) *pgxDriver {
	return &pgxDriver{
		name:  pgxName,
		dbURL: dbURL,
	}
}

// На будущее, возможно придется переделывать на что то такое

// func Initialize(dbURL string) (*sql.DB, error) {
// 	db, err := sql.Open(pgxName, dbURL)
// 	if err != nil {
// 		log.Fatalf("Error opening database: %v", err)
// 	}
// 	// Set the maximum number of open connections
// 	db.SetMaxOpenConns(25)
// 	// Set the maximum number of idle connections
// 	db.SetMaxIdleConns(25)
// 	// Set the maximum lifetime of a connection
// 	db.SetConnMaxLifetime(5 * time.Minute)
// 	return db, nil
// }

func (d *pgxDriver) Open() error {
	pool, err := pgxpool.New(context.Background(), d.dbURL)
	if err != nil {
		return err
	}
	d.conn = pool

	if err := d.createTables(); err != nil {
		return err
	}
	return nil
}

func (d *pgxDriver) Close() error {
	d.conn.Close()
	return nil
}

func (d *pgxDriver) Ping() error {
	return d.conn.Ping(context.Background())
}

func (d *pgxDriver) Save() error {
	return nil
}

func (d *pgxDriver) Restore() error {
	return nil
}

func (d *pgxDriver) Update(mtype, mname, mval string) error {
	_, err := d.exec(context.Background(), `
	INSERT INTO metrics (mtype, mname, mvalue) 
	VALUES ($1, $2, $3) 
	ON CONFLICT (mname) DO 
	UPDATE SET mvalue = EXCLUDED.mvalue
	`, mtype, mname, mval)
	return err
}

func (d *pgxDriver) Get(mtype, mname string) (string, error) {
	if mtype == "" || mname == "" {
		return "", errors.New("invalid metric type")
	}
	row := d.queryRow(context.Background(), `
	SELECT mvalue FROM metrics WHERE mtype=$1 AND mname=$2
	`, mtype, mname)
	var mval sql.NullString
	err := row.Scan(&mval)
	if err != nil {
		return "", err
	}
	if mval.Valid {
		return mval.String, nil
	}
	return "", errors.New("not found")
}

func (d *pgxDriver) GetAll() Data {
	return Data{}
}

func (d *pgxDriver) createTables() error {
	_, err := d.exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS metrics (
    	id SERIAL PRIMARY KEY,
    	mtype TEXT NOT NULL,
    	mname TEXT UNIQUE NOT NULL,
    	mvalue TEXT NOT NULL
	)
	`)
	return err
}
