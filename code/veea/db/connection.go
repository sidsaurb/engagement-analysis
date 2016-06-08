package db

import (
	"fmt"
	"database/sql"

	"github.com/gpahal/veea/conf"
	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var connectionTries int = 0

func openDb() error {
	if db != nil {
		return nil
	}

	connectionTries += 1

	tmpDb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=True", conf.DbUser, conf.DbPassword, conf.DbName))
	if err != nil {
		log.WithFields(log.Fields{
			"tries": connectionTries,
			"error": err.Error(),
		}).Error("Error opening the database")

		return err
	}

	if err := tmpDb.Ping(); err != nil {
		log.WithFields(log.Fields{
			"tries": connectionTries,
			"error": err.Error(),
		}).Error("Error connecting to the database")

		return err
	}

	log.WithFields(log.Fields{
		"tries": connectionTries,
	}).Info("Database connection established")

	db = tmpDb
	return nil
}

func closeDb() error {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error closing the database")
		}

		return err
	}

	return nil
}

func exec(query string, args ...interface{}) (sql.Result, error) {
	if db == nil {
		if err := openDb(); err != nil {
			var res sql.Result
			return res, err
		}
	}

	res, err := db.Exec(query, args...)
	return res, err
}

func query(query string, args ...interface{}) (*sql.Rows, error) {
	if db == nil {
		if err := openDb(); err != nil {
			return nil, err
		}
	}

	rows, err := db.Query(query, args...)
	return rows, err
}

func transaction() (*sql.Tx, error) {
	if db == nil {
		if err := openDb(); err != nil {
			return nil, err
		}
	}

	tx, err := db.Begin()
	return tx, err
}

func init() {
	openDb()
}
