package mysqlpersister

import (
	"database/sql"
	"fmt"
)

type UnsafeConnector struct {
	dbUser string
	dbPass string
	dbHost string
	dbPort int
	dbName string
}

func NewUnsafeConnector(dbUser, dbPass, dbHost string, dbPort int, dbName string) *UnsafeConnector {
	return &UnsafeConnector{
		dbUser: dbUser,
		dbPass: dbPass,
		dbHost: dbHost,
		dbPort: dbPort,
		dbName: dbName,
	}
}

func (c *UnsafeConnector) Connect() (*sql.DB, error) {
	return sql.Open("mysql",
		fmt.Sprintf("%s:%s@(%s:%v)/%s",
			c.dbUser,
			c.dbPass,
			c.dbHost,
			c.dbPort,
			c.dbName))
}
