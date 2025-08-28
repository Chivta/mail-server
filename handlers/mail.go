package handlers

import (
	"encoding/json"
	"io"
	"log"
	"mail-server/db"
	"net/http"
)

type MailHandler struct{
	DB *db.DB
}

func (h *MailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %v \n", r.RemoteAddr)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotImplemented)
		log.Printf("Status Method Not Allowed\n")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	body,err := io.ReadAll(r.Body)
	if err!= nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	var email db.Email
	err = json.Unmarshal(body,&email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = h.DB.WriteEmail(email)
	if err!= nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}