package main

import (
	"database/sql"
	"github.com/k0kubun/pp"
	soqlDriver "github.com/tzmfreedom/go-soql-driver"
	"os"
)

func main() {
	username := os.GetEnv("SFDC_USERNAME")
	password := os.GetEnv("SFDC_PASSWORD")

	dsn := soqlDriver.CreateDsn(username, password, "login.salesforce.com")
	db, err := sql.Open("soql", dsn)
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("SELECT Id, Name FROM Account")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		debug(id, name)
	}
}

func debug(args ...interface{}) {
	pp.Println(args...)
}
