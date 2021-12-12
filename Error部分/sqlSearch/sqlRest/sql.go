package sqlRest

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"os"
	"sqlSearch/dbConfig"
)

var sqldb *sql.DB
var sqlpath string

func CreateDB(conf dbConfig.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", conf.Config())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("open db in path info:%s", conf.DBPath+"/"+conf.DBName+".db"))
	}
	return db, nil
}

func OpenDb(conf dbConfig.DBConfig) (*sql.DB, error) {
	if err := pathExist(conf.DBPathInfo()); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", conf.Config())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("open db in path info:%s", conf.DBPath+"/"+conf.DBName+".db"))
	}
	return db, nil
}

func CloseDb(db *sql.DB) error {
	return db.Close()
}

func CreateTbWithName(db *sql.DB, conf dbConfig.DBConfig) error {
	tbString := conf.TbName + ` (
			id  INTEGER PRIMARY KEY AUTOINCREMENT,
			name   VARCHAR NOT NULL
		);`
	sql_table := `CREATE TABLE IF NOT EXISTS ` + tbString
	_, err := db.Exec(sql_table)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("create tb name:%s ,path:%s", conf.TbName, conf.DBPathInfo()))
	}
	return nil
}
func InsertData(db *sql.DB, conf dbConfig.DBConfig, name string) error {
	sql_ins := `insert into ` + conf.TbName + `(name) values(?)`
	_, err := db.Exec(sql_ins, name)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("insert data into db:%s data:%s", conf.DBPathInfo(), name))
	}
	return nil
}

func QueryData(db *sql.DB, conf dbConfig.DBConfig, name string) (string, error) {
	sql_quer := `select name from ` + conf.TbName + ` where name = ?`
	rows, err := db.Query(sql_quer, name)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("query data db:%s data:%s", conf.DBPath, name))
	}
	defer rows.Close()
	var value string
	for rows.Next() {
		err = rows.Scan(&value)
	}
	if err != nil {
		return value, err
	}
	return value, nil
}

func pathExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return errors.Wrap(err, fmt.Sprintf("path info : %s", path))
	} else {
		return err
	}
}
