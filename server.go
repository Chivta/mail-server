package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type MailHandler struct{}

func (h *MailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %v \n", r.RemoteAddr)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Status Method Not Allowed\n")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	decoder := json.NewDecoder(r.Body)

	var message map[string]any
	if err := decoder.Decode(&message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	log.Println(message)

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.Handle("/mail", &MailHandler{})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
