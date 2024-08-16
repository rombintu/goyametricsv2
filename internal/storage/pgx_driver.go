package storage

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

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

type AnyMetrics map[string]string

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

	// Почему то такая схема не работает. TODO
	// var pgErr *pgconn.PgError
	// Try connect to database
	// if _, errConn = d.conn.Acquire(context.Background()); errConn != nil {
	// if errors.As(err, &pgErr) {
	// 	logger.Log.Debug("Error is a pgconn.PgError", zap.String("code", pgErr.Code))
	// 	if pgerrcode.IsConnectionException(pgErr.Code) {
	// 		logger.Log.Debug(pgerrcode.ConnectionFailure, zap.Int("attemp", 1))
	// 		return err
	// 	}
	// } else {
	// 	return err
	// }

	// }
	var errConn error
	var ok bool
	for i := 1; i <= 5; i += 2 {
		if errConn = d.Ping(); errConn == nil {
			ok = true
			break
		}
		logger.Log.Debug("Try reconnect to database", zap.Int("sleep seconds", i))
		time.Sleep(time.Duration(i) * time.Second)
	}
	if !ok {
		return errConn
	}

	err = d.createTables()
	if err != nil {
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
	var sqlScript string
	switch mtype {
	case CounterType:
		sqlScript = `
		INSERT INTO metrics (mtype, mname, mvalue) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (mname) DO 
		UPDATE SET mvalue = EXCLUDED.mvalue::int + metrics.mvalue::int
		`
	case GaugeType:
		sqlScript = `
		INSERT INTO metrics (mtype, mname, mvalue) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (mname) DO 
		UPDATE SET mvalue = EXCLUDED.mvalue
		`
	}
	_, err := d.exec(context.Background(), sqlScript, mtype, mname, mval)
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

// TODO: нужны тесты, не хватает времени
func (d *pgxDriver) GetAll() Data {
	var data Data
	rows, err := d.queryRows(context.Background(), `SELECT mtype, mname, mvalue FROM metrics`)
	if err != nil {
		logger.Log.Error(err.Error())
		return data
	}
	defer rows.Close()

	counters := make(map[string]int64)
	gauges := make(map[string]float64)
	for rows.Next() {
		var mtype, mname, mvalue string
		if err = rows.Scan(&mtype, &mname, &mvalue); err != nil {
			logger.Log.Error(err.Error())
			return data
		}
		switch mtype {
		case CounterType:
			var value int64
			if value, err = strconv.ParseInt(mvalue, 10, 64); err != nil {
				return data
			}
			counters[mname] = value
		case GaugeType:
			var value float64
			if value, err = strconv.ParseFloat(mvalue, 64); err != nil {
				return data
			}
			gauges[mname] = value
		}
	}
	err = rows.Err()
	if err != nil {
		return data
	}
	data.Counters = counters
	data.Gauges = gauges
	return data
}

func (d *pgxDriver) UpdateAll(data Data) error {
	ctx := context.Background()
	counters := counters2Any(data.Counters)
	gauges := gauges2Any(data.Gauges)
	if err := d.updateAny(ctx, counters, CounterType); err != nil {
		return err
	}
	if err := d.updateAny(ctx, gauges, GaugeType); err != nil {
		return err
	}
	return nil
}

func (d *pgxDriver) updateAny(ctx context.Context, m AnyMetrics, mtype string) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var sqlScript string
	switch mtype {
	case CounterType:
		sqlScript = `
		INSERT INTO metrics (mtype, mname, mvalue) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (mname) DO 
		UPDATE SET mvalue = EXCLUDED.mvalue::int + metrics.mvalue::int 
		`
	case GaugeType:
		sqlScript = `
		INSERT INTO metrics (mtype, mname, mvalue) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (mname) DO 
		UPDATE SET mvalue = EXCLUDED.mvalue
		`
	default:
		return errors.New("invalid metric type")
	}

	for mname, mvalue := range m {
		_, err := tx.Exec(ctx, sqlScript, mtype, mname, mvalue)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
