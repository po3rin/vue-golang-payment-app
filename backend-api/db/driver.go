package db

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Conn - sql connection handler
var Conn *sql.DB

// NewSQLHandler - init sql handler
func init() {
	user := os.Getenv("MYSQL_USER")
	host := os.Getenv("MYSQL_HOST")
	pass := os.Getenv("MYSQL_PASSWORD")
	name := os.Getenv("MYSQL_DATABASE")
	port := os.Getenv("MYSQL_PORT")

	dbconf := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name
	conn, err := sql.Open("mysql", dbconf)
	if err != nil {
		panic(err.Error)
	}
	Conn = conn
}
