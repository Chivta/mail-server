package main

import (
	"fmt"
	"log"
	"mail-server/db"
	"mail-server/handlers"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
    Server struct {
		Address 		string	`yaml:"address"`
        Port    		int   	`yaml:"port"`
        LogFile 		string	`mapstructure:"log_file" yaml:"log_file"`
		SendMailDelay	int64	`mapstructure:"send_mail_delay" yaml:"send_mail_delay"`
		SendMailsAtOnce	int		`mapstructure:"send_mail_delay" yaml:"send_mails_at_once"`
    } `yaml:"server"`
    Database struct {
        Host 			string 	`yaml:"host"`
		Port 			int		`yaml:"port"`
        Name 			string 	`yaml:"name"`
        User 			string 	`yaml:"user"`
		Password 		string 
    } `yaml:"database"`
}

func ParseConfig(path string) (*Config){
	viper.SetConfigFile("config.yaml")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        log.Fatal(err)
    }
	cfg.Database.Password = os.Getenv("MAIL_SERVER_DB_PASS")

	return &cfg
}

type EmailSender struct{
	DB *db.DB
	delay time.Duration
	mailsAtOnce int
}

func (es *EmailSender) SendEmails(emails []db.Email) ([]int,error) {
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var sent_emails_id []int
	for _, email := range emails {
		if err := c.Mail(email.From); err != nil {
			log.Println(err)
			continue
		}
		if err := c.Rcpt(email.To); err != nil {
			log.Println(err)
			continue
		}

		wc, err := c.Data()
		if err != nil {
			log.Println(err)
			continue
		}
		msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
			email.From, email.To, email.Subject, email.Body)
		_, err = fmt.Fprint(wc, msg)
		if err != nil {
			log.Println(err)
			continue
		}
		sent_emails_id = append(sent_emails_id, email.ID)
		wc.Close()
	}
	c.Quit()
	return sent_emails_id, nil
}

func (es *EmailSender) Start() {
	for {
		time.Sleep(es.delay)
		log.Println("Sending emails")
		emails, err:= es.DB.GetUnsentEmails(es.mailsAtOnce,0);

		if err!=nil{
			log.Println(err)
			continue
		}
		
		sent_emails_id, err := es.SendEmails(emails)
		if err!=nil{
			log.Println(err)
			continue
		}

		es.DB.MarkEmailsSent(sent_emails_id)

		log.Printf("%d emails was sent \n",len(sent_emails_id))
	}

}



func main() {
	cfg := ParseConfig("config.yaml")

	switch cfg.Server.LogFile{
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		logfile, err := os.OpenFile(cfg.Server.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err!=nil{
			log.Fatal(err)
		}
		defer logfile.Close()

		log.SetOutput(logfile)
	}


	database, err := db.NewDBConnection(
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port)

	if err!=nil{
		log.Fatal(err)
	}

	es := EmailSender{DB: database, delay: time.Duration(cfg.Server.SendMailDelay) * time.Second, mailsAtOnce: cfg.Server.SendMailsAtOnce}
	go es.Start()

	http.Handle("/mail", &handlers.MailHandler{DB: database})

	table_handler,err := handlers.GetTableHandler(database)
	if err!= nil{
		log.Fatal(err)
	}
	http.Handle("/table", &table_handler)
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",cfg.Server.Port), nil))
}
