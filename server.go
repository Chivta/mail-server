package main

import (
	"fmt"
	"log"
	"mail-server/db"
	"mail-server/handlers"
	"net/http"
	"os"
	"github.com/spf13/viper"
)

type Config struct {
    Server struct {
		Address string	`yaml:"address"`
        Port    int   	`yaml:"port"`
        LogFile string	`yaml:"log_file"`
    } `yaml:"server"`
    Database struct {
        Host string 	`yaml:"host"`
		Port int		`yaml:"port"`
        Name string 	`yaml:"name"`
        User string 	`yaml:"user"`
		Password string 
    } `yaml:"database"`
}


func main() {
	viper.SetConfigFile("config.yaml")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        log.Fatal(err)
    }
	cfg.Database.Password = os.Getenv("MAIL_SERVER_DB_PASS")

	database, err := db.NewDBConnection(
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port)

	if err!=nil{
		log.Fatal(err)
	}

	http.Handle("/mail", &handlers.MailHandler{DB: database})
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d",cfg.Server.Address,cfg.Server.Port), nil))
}
