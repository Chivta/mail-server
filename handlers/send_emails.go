package handlers

import (
	"fmt"
	"log"
	"mail-server/db"
	"mail-server/managers"
	"net/http"
	"strconv"
)

type SendEmailsHandler struct{
	DB *db.DB
	MM *managers.MailManager
}

func (h *SendEmailsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Post from %v \n", r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		writeJSONError(w,501,fmt.Errorf("method is not allowed"))
		return
	}

	r.ParseForm()
	selected := r.Form["email	IDs"]
	log.Printf("%v\n",selected)

	ids := make([]int64, len(selected))
	for i,s := range selected{
		v,err := strconv.ParseInt(s,10,64)
		if err!=nil{
			continue
		}
		ids[i]=v
	}

	emails,err := h.DB.GetEmailsByIds(ids)
	
	if err != nil{
		log.Println(err)
		writeJSONError(w,500,err)
		return
	}

	err = h.MM.SendEmailsAndMarkSent(emails)
	if err != nil{
		log.Println(err)
		writeJSONError(w,500,err)
		return
	}

	redirectURL := r.Form["redirect_to"][0]
	if redirectURL == "" {
		redirectURL = "/table"
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
