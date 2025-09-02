package db

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/lib/pq"
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
    ID      	int   		`json:"id"`
    From    	string		`json:"from"`
    To      	string		`json:"to"`
    Date    	time.Time	`json:"date"`
    Subject 	string		`json:"subject"`
	Reason		string		`json:"reason"`
    Body    	string		`json:"body"`
	RegistrarId	string 		`json:"registrar_id"`
	Sent 		bool		`json:"sent"`
	Status 		string		`json:"status"`
}

type DB struct{
	conn *sql.DB
} 

func (db *DB) MarkEmailsSent(ids []int) error {
    if len(ids) == 0 {
        return nil 
    }

    query := `UPDATE email SET sent = true WHERE id = ANY($1);`

    _, err := db.conn.Exec(query, pq.Array(ids))
    if err != nil {
        return err
    }
    return nil
}

func (db *DB) WriteEmail(email Email) error {
	_, err := db.conn.Query(`
		INSERT INTO email(
			"from", 
			"to", 
			date, 
			subject, 
			reason,
			body,
			redistrarid, 
			sent, 
			status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
			email.From, 
			email.To, 
			email.Date, 
			email.Subject, 
			email.Reason,
			email.Body, 
			email.RegistrarId,
			email.Sent, 
			email.Status,
		)
	if err!= nil{
		return err
	}
	return nil
}

func (db *DB) GetEmails(limit, offset int, search string) ([]Email,error){
	if search == "" {
		search = "%" // match all
	} else {
		search = "%" + search + "%"
	}

	rows, err := db.conn.Query(`
		SELECT *
		FROM email
		WHERE (
			"from" || ' ' || 
			"to" || ' ' || 
			subject || ' ' || 
			reason || ' ' || 
			body || ' ' || 
			redistrarid || ' ' || 
			status)
			ILIKE $1
		LIMIT $2
		OFFSET $3;
		`, search, limit, offset)
	if err!= nil{
		return nil, err
	}
	
	var emails []Email
	var email Email
	for rows.Next(){
		err = rows.Scan(&email.ID,&email.From,&email.To,&email.Date,&email.Subject,&email.Reason,&email.Body,&email.RegistrarId,&email.Sent,&email.Status)
		if err!=nil{
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (db *DB) GetUnsentEmails(limit, offset int) ([]Email,error){
	rows, err := db.conn.Query(`
		SELECT * FROM email
		WHERE sent = false
		LIMIT $1
		OFFSET $2;`,
		limit,offset)
	if err!= nil{
		return nil, err
	}
	
	var emails []Email
	var email Email
	for rows.Next(){
		err = rows.Scan(&email.ID,&email.From,&email.To,&email.Date,&email.Subject,&email.Reason,&email.Body,&email.RegistrarId,&email.Sent,&email.Status)
		if err!=nil{
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}

