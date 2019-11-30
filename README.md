# go-soql-driver

SOQL driver for Go using database/sql

## Install

```bash
go get github.com/tzmfreedom/go-soql-driver
```

## Usage

```go
package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

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
	rows, err := db.Query("SELECT Id, Name FROM Account")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		fmt.Println(id, name)
	}
}
```

DSN definition
```
{url_escaped_username}:{url_escaped_password}@{login_url}
```
