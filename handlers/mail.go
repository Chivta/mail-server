package handlers

import (
	"encoding/json"
	"fmt"
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	if r.Method != http.MethodPost {
		writeJSONError(w,501,fmt.Errorf("method is not allowed"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	body,err := io.ReadAll(r.Body)
	if err!= nil{
		writeJSONError(w,500,err)
		return
	}
	
	var email db.Email
	err = json.Unmarshal(body,&email)
	if err != nil {
		writeJSONError(w,500,err)
		return
	}

	err = h.DB.WriteEmail(email)
	if err!= nil{
		writeJSONError(w,500,err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}