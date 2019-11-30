package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/tzmfreedom/go-soapforce"
	"regexp"
)

type SOQLDriver struct {
}

type Connection struct {
	lr *soapforce.LoginResult
	client *soapforce.Client
}

type Config struct {
	username string
	password string
	hostname string
}

type Stmt struct {
	client *soapforce.Client
	query string
}

type Result struct {

}

func (r *Result) LastInsertId() (int64, error) {
	panic("implement me")
}

func (r *Result) RowsAffected() (int64, error) {
	panic("implement me")
}

type Rows struct {}

func (r *Rows) Columns() []string {
	panic("implement me")
}

func (r *Rows) Close() error {
	panic("implement me")
}

func (r *Rows) Next(dest []driver.Value) error {
	panic("implement me")
}

func (s *Stmt) Close() error {
	return nil
}

func (s *Stmt) NumInput() int {
	return 0
}

func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return &Result{}, nil
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return &Rows{}, nil
}

func (d *SOQLDriver) Open(dsn string) (driver.Conn, error) {
	cfg, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return login(cfg)
}

func (c *Connection) Prepare(query string) (driver.Stmt, error) {
	return &Stmt{
		client: c.client,
		query: query,
	}, nil
}

func (c *Connection) Close() error {
	return nil
}

func (c *Connection) Begin() (driver.Tx, error) {
	return nil, nil
}

func (c *Connection) ConnBeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, nil
}

func CreateDsn(username, password, hostname string) string {
	return fmt.Sprintf("%s:%s@%s", username, password, hostname)
}

func parseDSN(dsn string) (*Config, error) {
	r := regexp.MustCompile(`^([^@]+@[^@]+):([^@]+)@([^@]+)$`)
	if r.MatchString(dsn) {
		m := r.FindStringSubmatch(dsn)
		username := m[1]
		password := m[2]
		hostname := m[3]
		return &Config{
			username: username,
			password: password,
			hostname: hostname,
		}, nil
	}
	return nil, errors.New("DSN Parse Error")
}

func login(cfg *Config) (driver.Conn, error) {
	client := soapforce.NewClient()
	client.SetLoginUrl(cfg.hostname)
	lr, err := client.Login(cfg.username, cfg.password)
	if err != nil {
		return nil, err
	}
	return &Connection{
		lr: lr,
		client: client,
	}, nil
}

func init() {
	sql.Register("soql", &SOQLDriver{})
}
