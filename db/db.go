package db

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq"
)


func NewDBConnection(host, user, password, dbname string, port int) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
    	host, port, user, password, dbname)


	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

type Email struct {
    ID      int   		`json:"id"`
    From    string		`json:"from"`
    To      string		`json:"to"`
    Date    time.Time	`json:"date"`
    Subject string		`json:"subject"`
    Body    string		`json:"body"`
	Sent 	bool		`json:"sent"`
	Status 	string		`json:"status"`
}

type DB struct{
	conn *sql.DB
} 

func (db *DB) WriteEmail(email Email) error {
	_, err := db.conn.Query(`
		INSERT INTO email("from", "to", date, subject, body, sent, status) VALUES ($1, $2, $3, $4, $5, $6, $7);`,
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
		LIMIT $1;`,
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

