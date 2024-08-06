package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type pgxDriver struct {
	dbURL string
	Conn  *pgx.Conn
}

func NewPgxDriver(dbURL string) *pgxDriver {
	return &pgxDriver{
		dbURL: dbURL,
	}
}

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
