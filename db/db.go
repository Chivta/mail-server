package db

import (
	"database/sql"
	"fmt"
	"strings"
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
    ID      	int   		
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

func (db *DB) WriteEmailStatus(id, status string) error {
	query := `UPDATE email WHERE id = $1 SET status = $2;`

    _, err := db.conn.Exec(query, id, status)
	if err != nil {
        return err
    }
    return nil

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
			registrarid, 
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

func (db *DB) GetEmailsByIds(IDs []int64) ([]Email, error) {
	if len(IDs)==0{
		return []Email{},nil
	}

	rows, err := db.conn.Query(`
		SELECT * FROM email
		WHERE id = ANY($1)
		`, pq.Array(IDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails, err := fillEmailSlice(rows)
	if err!= nil{
		return nil, err
	}
	return emails,nil
}

func (db *DB) GetEmails(limit, offset int, search string, selectedColumns []string) ([]Email, error) {
	if search == "" {
		search = "%" // match all
	} else {
		search = "%" + search + "%"
	}

	// Default to all searchable columns if none selected
	allColumns := []string{"id", "from", "to", "subject", "reason", "body", "registrarid", "sent", "status"}
	if len(selectedColumns) == 0 {
		selectedColumns = allColumns
	}

	// Build WHERE conditions dynamically
	var conditions []string
	// var args []interface{}
	for _, col := range selectedColumns {
		conditions = append(conditions, fmt.Sprintf(`CAST("%s" AS TEXT) ILIKE $1`, col))
	}

	searchQuery := strings.Join(conditions, " OR ")

	// Query
	query := fmt.Sprintf(`
		SELECT *
		FROM email
		WHERE %s
		LIMIT $2 OFFSET $3;
	`, searchQuery)

	rows, err := db.conn.Query(query, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fillEmailSlice(rows)
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
	defer rows.Close()
	
	return fillEmailSlice(rows)
}

func fillEmailSlice(rows *sql.Rows) ([]Email, error){
	var emails []Email
	var email Email
	for rows.Next(){
		err := rows.Scan(&email.ID,&email.From,&email.To,&email.Date,&email.Subject,&email.Reason,&email.Body,&email.RegistrarId,&email.Sent,&email.Status)
		if err!=nil{
			return nil, err
		}
		emails = append(emails, email)
	}
	defer rows.Close()

	return emails, nil
}