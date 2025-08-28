package main

import (
	"log"
	"mail-server/db"
	"mail-server/handlers"
	"net/http"
	"time"
)

// func readExistingEmails(DB *db.DB){
// 	time.Sleep(10 * time.Second)

// 	res,err:= DB.GetUnsentEmails(10)

// 	if err!=nil{
// 		log.Println(err)
// 		return
// 	}

// 	log.Println(res)
// }

func main() {
	database, err := db.NewDBConnection()
	if err!=nil{
		log.Fatal(err)
	}

	http.Handle("/mail", &handlers.MailHandler{DB: database})
	// go readExistingEmails(database)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
