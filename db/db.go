package db

import (
	"database/sql"
	_ "log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

func NewDBConnection() (*DB, error) {
	projectDir := filepath.Dir(os.Args[0])
	dataDir := filepath.Join(projectDir, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", "data/db.sqlite")

	if err != nil {
		return nil, err
	}
	// _, err = db.Query(`DROP TABLE IF EXISTS email;`)
	// if err!= nil{
	// 	return nil,err
	// }
	_, err = db.Query(`
	CREATE TABLE IF NOT EXISTS 
		email(
			id INTEGER PRIMARY KEY,
			"from" TEXT NOT NULL,
			"to" TEXT NOT NULL,
			date TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			sent BOOLEAN NOT NULL,
			status TEXT NOT NULL
		);
	`)
	if err!= nil{
		return nil,err
	}

	return &DB{db}, nil
}

type Email struct {
    ID      int   	`json:"id"`
    From    string	`json:"from"`
    To      string	`json:"to"`
    Date    string	`json:"date"`
    Subject string	`json:"subject"`
    Body    string	`json:"body"`
	Sent 	bool	`json:"sent"`
	Status 	string	`json:"status"`
}

type DB struct{
	conn *sql.DB
} 

func (db *DB) WriteEmail(email Email) error {
	_, err := db.conn.Query(`
		INSERT INTO email("from", "to", date, subject, body, sent, status) VALUES(?,?,?,?,?,?,?);`,
		email.From, email.To, email.Date, email.Subject, email.Body, email.Sent, email.Status)
	if err!= nil{
		return err
	}
	return nil
}

func (db *DB) GetUnsentEmails(limit int) ([]Email,error){
	rows, err := db.conn.Query(`
		SELECT * FROM email
		WHERE sent = false
		LIMIT ?;`,
		limit)
	if err!= nil{
		return nil, err
	}
	
	var emails []Email
	var email Email
	for rows.Next(){
		
		err = rows.Scan(&email.ID,&email.From,&email.To,&email.Date,&email.Subject,&email.Body,&email.Sent,&email.Status)
		if err!=nil{
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}

