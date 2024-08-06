package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const (
	pgxName = "postgres"
)

type pgxDriver struct {
	name  string
	dbURL string
	Conn  *pgx.Conn
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
	conn, err := pgx.Connect(context.Background(), d.dbURL)
	if err != nil {
		return err
	}
	d.Conn = conn
	return nil
}

func (d *pgxDriver) Close() error {
	return d.Conn.Close(context.Background())
}

func (d *pgxDriver) Ping() error {
	return d.Conn.Ping(context.Background())
}

func (d *pgxDriver) Save() error {
	return nil
}

func (d *pgxDriver) Restore() error {
	return nil
}

func (d *pgxDriver) Update(mtype, mname, mval string) error {
	return nil
}

func (d *pgxDriver) Get(mtype, mname string) (string, error) {
	return "", nil
}

func (d *pgxDriver) GetAll() Data {
	return Data{}
}
