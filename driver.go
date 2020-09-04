package soqlDriver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/tzmfreedom/go-soapforce"
	"github.com/tzmfreedom/soql-cli/parser"
)

var ExecMode = ExecModeDenyNoWhere

const (
	ExecModeDenyNoWhere = iota
	ExecModeAllowNoWhere
)

type SOQLDriver struct {
}

type Connection struct {
	lr     *soapforce.LoginResult
	client *soapforce.Client
}

type Config struct {
	username string
	password string
	hostname string
}

type Stmt struct {
	client *soapforce.Client
	query  string
}

type Result struct {
	rowsAffected int64
}

func (r *Result) LastInsertId() (int64, error) {
	panic("LastInsertId does not support")
}

func (r *Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type Rows struct {
	index        int
	selectFields []string
	records      []*soapforce.SObject
}

func (r *Rows) Columns() []string {
	return r.selectFields
}

func (r *Rows) Close() error {
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.index >= len(r.records) {
		return io.EOF
	}
	record := r.records[r.index]
	for i, field := range r.selectFields {
		parts := strings.Split(field, ".")
		dest[i] = r.getField(record, parts)
	}
	r.index++
	return nil
}

func (r *Rows) getField(record *soapforce.SObject, parts []string) interface{} {
	key := parts[0]
	if len(parts) == 1 {
		if strings.ToLower(key) == "id" {
			return record.Id
		}
		return record.Fields[key]
	}
	return r.getField(record.Fields[key].(*soapforce.SObject), parts[1:])
}

func (s *Stmt) Close() error {
	return nil
}

func (s *Stmt) NumInput() int {
	return 0
}

func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	stmt := parser.ParseString(s.query)
	switch stmt.Type {
	case "INSERT":
		res, err := s.insert(stmt)
		if err != nil {
			return nil, err
		}
		if res.Success {
			//fmt.Printf("%s Created: Id = %s\n", stmt.Sobject, res.Id)
		} else {
			//for _, error := range res.Errors {
			//	fmt.Fprintln(os.Stderr, error.Message)
			//}
		}
		return &Result{1}, nil
	case "UPDATE":
		results, err := s.update(stmt)
		if err != nil {
			return nil, err
		}
		var successCnt int64 = 0
		for _, res := range results {
			if res.Success {
				successCnt++
				//fmt.Printf("%s Updated: Id = %s\n", stmt.Sobject, res.Id)
			} else {
				//for _, error := range res.Errors {
				//	fmt.Fprintln(os.Stderr, error.Message)
				//}
			}
		}
		return &Result{successCnt}, nil
	case "DELETE":
		results, err := s.delete(stmt)
		if err != nil {
			return nil, err
		}
		var successCnt int64 = 0
		for _, res := range results {
			if res.Success {
				successCnt++
				//fmt.Printf("%s Deleted: Id = %s\n", stmt.Sobject, res.Id)
			} else {
				//for _, error := range res.Errors {
				//	fmt.Fprintln(os.Stderr, error.Message)
				//}
			}
		}
		return &Result{successCnt}, nil
	}
	return nil, nil
}

func (s *Stmt) insert(stmt *parser.Statement) (*soapforce.SaveResult, error) {
	obj := &soapforce.SObject{}
	obj.Type = stmt.Sobject
	obj.Fields = map[string]interface{}{}
	for field, value := range stmt.Values {
		obj.Fields[field] = value
	}
	results, err := s.client.Create([]*soapforce.SObject{obj})
	if err != nil {
		return nil, err
	}
	res := results[0]
	return res, nil
}

func (s *Stmt) update(stmt *parser.Statement) ([]*soapforce.SaveResult, error) {
	if ExecMode == ExecModeDenyNoWhere && stmt.Where == "" {
		return nil, errors.New("WHERE clause should not be blank")
	}
	q := fmt.Sprintf("SELECT id FROM %s %s", stmt.Sobject, stmt.Where)
	r, err := s.client.Query(q)
	if err != nil {
		return nil, err
	}
	updateObjects := make([]*soapforce.SObject, len(r.Records))
	for i, record := range r.Records {
		obj := &soapforce.SObject{}
		obj.Id = record.Id
		obj.Type = stmt.Sobject
		obj.Fields = map[string]interface{}{}
		for field, value := range stmt.Values {
			obj.Fields[field] = value
		}
		updateObjects[i] = obj
	}
	return s.client.Update(updateObjects)
}

func (s *Stmt) delete(stmt *parser.Statement) ([]*soapforce.DeleteResult, error) {
	if ExecMode == ExecModeDenyNoWhere && stmt.Where == "" {
		return nil, errors.New("WHERE clause should not be blank")
	}
	q := fmt.Sprintf("SELECT id FROM %s %s", stmt.Sobject, stmt.Where)
	r, err := s.client.Query(q)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(r.Records))
	for i, record := range r.Records {
		ids[i] = record.Id
	}
	return s.client.Delete(ids)
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	r := regexp.MustCompile(`^(?i)SELECT\s+(.+)\s+FROM\s+`)
	m := r.FindStringSubmatch(s.query)
	fields := strings.Split(m[1], ",")
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}
	q, err := s.client.Query(s.query)
	if err != nil {
		return nil, err
	}
	return &Rows{
		selectFields: fields,
		records:      q.Records,
	}, nil
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
		query:  query,
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
	r := regexp.MustCompile(`^([^@]+):([^@]+)@([^@]+)$`)
	if r.MatchString(dsn) {
		m := r.FindStringSubmatch(dsn)
		username, err := url.QueryUnescape(m[1])
		if err != nil {
			return nil, err
		}
		password, err := url.QueryUnescape(m[2])
		if err != nil {
			return nil, err
		}
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
		lr:     lr,
		client: client,
	}, nil
}

func init() {
	sql.Register("soql", &SOQLDriver{})
}

func debug(args ...interface{}) {
	pp.Println(args...)
}
