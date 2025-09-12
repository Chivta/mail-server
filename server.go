package main

import (
	"fmt"
	"log"
	"mail-server/db"
	"mail-server/handlers"
	"mail-server/managers"
	"net/http"
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

	es := managers.MailManager{DB: database, Delay: time.Duration(cfg.Server.SendMailDelay) * time.Second, MailsAtOnce: cfg.Server.SendMailsAtOnce}
	go es.StartSending()

	http.Handle("/mail", &handlers.MailHandler{DB: database})
	table_handler,err := handlers.GetTableHandler(database)
	if err!= nil{
		log.Fatal(err)
	}
	http.Handle("/table", &table_handler)
	
	http.Handle("/sendEmails", &handlers.SendEmailsHandler{DB: database, MM: &es})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",cfg.Server.Port), nil))
}
