package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"github.com/k0kubun/pp"
	soqlDriver "github.com/tzmfreedom/go-soql-driver"
)

func main() {
	username := os.Getenv("SFDC_USERNAME")
	password := os.Getenv("SFDC_PASSWORD")
	dsn := soqlDriver.CreateDsn(url.QueryEscape(username), url.QueryEscape(password), "login.salesforce.com")
	db, err := sql.Open("soql", dsn)
	if err != nil {
		panic(err)
	}
	query(db)
	//insert(db)
	//update(db)
	//delete(db)
}

func query(db *sql.DB) {
	rows, err := db.Query("SELECT Id, Name FROM Account ORDER BY CreatedDate DESC LIMIT 10")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		debug(id, name)
		fmt.Println(id, name)
	}
}

func insert(db *sql.DB) {
	r, err := db.Exec("INSERT INTO Account(Name) VALUES ('Created By sql driver')")
	if err != nil {
		panic(err)
	}
	debug(r.RowsAffected())
	query(db)
}

func update(db *sql.DB) {
	r, err := db.Exec("UPDATE Account SET Name = 'Updated By sql driver' WHERE Name = 'Created By sql driver'")
	if err != nil {
		panic(err)
	}
	debug(r.RowsAffected())
	query(db)
}

func delete(db *sql.DB) {
	r, err := db.Exec("DELETE FROM Account WHERE Name = 'Updated By sql driver'")
	if err != nil {
		panic(err)
	}
	debug(r.RowsAffected())
	query(db)
}

func debug(args ...interface{}) {
	pp.Println(args...)
}
