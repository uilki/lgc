package server

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const dbName = "letsgochat"
const tbName = "backlog"
const openDb = "rootUser:%s@tcp(db-letsgochat.c3fb56m8zegn.us-east-1.rds.amazonaws.com:3306)/"

// statements
const (
	createDb = "CREATE DATABASE IF NOT EXISTS " + dbName
	dropTb   = "DROP TABLE IF EXISTS " + tbName
	createTb = `
	CREATE TABLE IF NOT EXISTS %s
		(
			id int AUTO_INCREMENT,
			name varchar(255),
			timestamp varchar(255), 
			message varchar(255), 
			PRIMARY KEY (id)
		)
	`
	selectAll  = "SELECT * FROM %s"
	addMessage = `
	INSERT INTO %s (name, timestamp, message)
	VALUES ('%s', '%s', '%s')
	`
)

type sqlBacklog struct {
	db *sql.DB
}

func NewSqlBacklog(pass string) (*sqlBacklog, error) {
	var backlog sqlBacklog
	var err error
	
	backlog.db, err = sql.Open("mysql", fmt.Sprintf(openDb, pass))
	if err != nil {
		return nil, err
	}

	_, err = backlog.db.Exec(createDb)
	backlog.db.Close()

	if err != nil {
		return nil, err
	}

	backlog.db, err = sql.Open("mysql", fmt.Sprintf(openDb, pass)+dbName)
	if err != nil {
		return nil, err
	}

	_, err = backlog.db.Exec(dropTb)
	if err != nil {
		backlog.db.Close()
		return nil, err
	}

	_, err = backlog.db.Exec(fmt.Sprintf(createTb, tbName))
	if err != nil {
		backlog.db.Close()
		return nil, err
	}

	return &backlog, nil
}

func (b *sqlBacklog) Close() {
        b.db.Close()
}

func (b *sqlBacklog) Update(m Message) (err error) {
	_, err = b.db.Exec(fmt.Sprintf(addMessage, tbName, m.Name, m.TimeStamp.Format(time.RFC850), m.Message))
	return
}

func (b *sqlBacklog) GetHistory() (history []Message, err error) {
	var rows *sql.Rows
	rows, err = b.db.Query(fmt.Sprintf(selectAll, tbName))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		m := Message{}
		var timestamp string
		var id int
		if err = rows.Scan(&id, &m.Name, &timestamp, &m.Message); err != nil {
			return nil, err
		}

		if m.TimeStamp, err = time.Parse(time.RFC850, timestamp); err != nil {
			return nil, err
		}

		history = append(history, m)
	}

	return history, nil
}
