package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gumper23/dbstuff/dbhelper"
)

// MySQLVariable variables have a name and a value
type MySQLVariable struct {
	Name  string `json:"variable_name"`
	Value string `json:"variable_value"`
}

// GetMySQLDSN returns a TCP DSN for MySQL
func GetMySQLDSN(username, password, host, port, dbname, timeout string) (dsn string) {
	if len(port) == 0 {
		port = "3306"
	}
	if len(dbname) == 0 {
		dbname = "information_schema"
	}
	if len(timeout) == 0 {
		timeout = "timeout=3s"
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", username, password, host, port, dbname, timeout)
	return
}

// GetMySQLDB returns an active mysql connection
func GetMySQLDB(username, password, host, port, dbname, timeout string) (db *sql.DB, err error) {
	dsn := GetMySQLDSN(username, password, host, port, dbname, timeout)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return
	}
	err = db.Ping()
	return
}

// GetMySQLQueries returns the status of the MySQL instance
func GetMySQLQueries(username, password, host, port, dbname, timeout string) (queries map[string]string, err error) {
	db, err := GetMySQLDB(username, password, host, port, dbname, timeout)
	if err != nil {
		return
	}
	defer db.Close()

	queries, _, err = dbhelper.QueryRow(db, "select now() as queries_ts, truncate(unix_timestamp(now(6)) * 1000000, 0) as queries_us, variable_value as queries from global_status where variable_name = 'queries'")
	if err != nil {
		return
	}
	return
}

// GetMySQLVariables returns a slice of MySQLVariable objects
func GetMySQLVariables(username, password, host, port, dbname, timeout, variableList string) (variables []MySQLVariable, err error) {
	db, err := GetMySQLDB(username, password, host, port, dbname, timeout)
	if err != nil {
		return
	}
	defer db.Close()

	// Transform: variables="var1,var2"
	// Into:      where variable_name in ('var1', 'var2')
	query := "select variable_name, variable_value from global_variables"
	if len(variableList) > 0 {
		query += " where variable_name in ("
		variableNames := strings.Split(variableList, ",")
		for i, variableName := range variableNames {
			query += fmt.Sprintf("'%s'", variableName)
			if i != len(variableNames)-1 {
				query += ", "
			} else {
				query += ")"
			}
		}
	}

	rows, _, err := dbhelper.QueryRows(db, query)
	if err != nil {
		return
	}
	variables = make([]MySQLVariable, 0, len(rows))
	for _, row := range rows {
		var variable MySQLVariable
		variable.Name = row["variable_name"]
		variable.Value = row["variable_value"]
		variables = append(variables, variable)
	}
	return
}

// SetMySQLVariables sets global MySQL variables and returns the new state
func SetMySQLVariables(username, password, host, port, dbname, timeout string, variableMap map[string]string) (variables []MySQLVariable, err error) {
	db, err := GetMySQLDB(username, password, host, port, dbname, timeout)
	if err != nil {
		return
	}
	defer db.Close()

	// Execute the set statements
	for varName, varValue := range variableMap {
		_, err = db.Exec(fmt.Sprintf("set global %s = %s", varName, varValue))
		if err != nil {
			return
		}
	}

	// Return the new state of the variables
	varNames := make([]string, 0, len(variableMap))
	for varName := range variableMap {
		varNames = append(varNames, varName)
	}
	varList := strings.Join(varNames, ",")
	variables, err = GetMySQLVariables(username, password, host, port, dbname, timeout, varList)
	return
}
