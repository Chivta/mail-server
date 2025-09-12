package managers

import (
	"net/smtp"
	"mail-server/db"
	"time"
	"log"
	"fmt"
)

// Sends unsent mails from db once in a delay
// Parses logs from mail server and updates status in db
type MailManager struct{
	DB *db.DB
	Delay time.Duration
	MailsAtOnce int
}

func (es *MailManager) SendEmailsAndMarkSent(emails []db.Email) error{
	sent_emails_id, err := es.SendEmails(emails)
	if err!=nil{
		log.Println(err)
		return err
	}

	es.DB.MarkEmailsSent(sent_emails_id)

	log.Printf("%d emails was sent \n",len(sent_emails_id))

	return nil
}

func (es *MailManager) SendEmails(emails []db.Email) ([]int,error) {
	client, err := smtp.Dial("localhost:25")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var sent_emails_id []int
	for _, email := range emails {
		if err := client.Mail(email.From); err != nil {
			log.Println(err)
			continue
		}
		if err := client.Rcpt(email.To); err != nil {
			log.Println(err)
			continue
		}

		wc, err := client.Data()
		if err != nil {
			log.Println(err)
			continue
		}
		msg := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\nX-Tracking-ID: %d\r\n\r\n%s",
			email.From, email.To, email.Subject, email.ID, email.Body)
		_, err = fmt.Fprint(wc, msg)
		if err != nil {
			log.Println(err)
			continue
		}
		sent_emails_id = append(sent_emails_id, email.ID)
		wc.Close()
	}
	client.Quit()
	return sent_emails_id, nil
}

func (es *MailManager) StartSending() {
	for {
		time.Sleep(es.Delay)
		log.Println("Sending emails")
		emails, err:= es.DB.GetUnsentEmails(es.MailsAtOnce,0);

		if err!=nil{
			log.Println(err)
			continue
		}
		
		es.SendEmailsAndMarkSent(emails)
	}

}