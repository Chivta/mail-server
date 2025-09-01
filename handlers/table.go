package handlers

import (
	"html"
	"strings"
	"html/template"
	"fmt"
	"log"
	"mail-server/db"
	"net/http"
	"time"
	"strconv"
)

type TableHandler struct{
	DB *db.DB
	Template *template.Template
}

func GetTableHandler(db *db.DB) (TableHandler, error) {
	handler := TableHandler{DB: db}

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"highlight": func(value interface{}, search string) template.HTML {
    var text string
    switch v := value.(type) {
    case string:
        text = v
    case int:
        text = fmt.Sprintf("%d", v)
    case bool:
        if v {
            text = "Yes"
        } else {
            text = "No"
        }
    case time.Time:
        text = v.Format("2006-01-02 15:04:05")
    default:
        text = fmt.Sprintf("%v", v)
    }
    if search == "" {
        return template.HTML(text)
    }
    escaped := html.EscapeString(text)
    highlighted := strings.ReplaceAll(
        escaped,
        search,
        fmt.Sprintf(`<span class="highlight">%s</span>`, html.EscapeString(search)),
    )
    return template.HTML(highlighted)
},
		"add": func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
		"div": func(a, b int) int { return a / b },
		"slice": func(vals ...int) []int { return vals },
	}

	tmpl := template.New("emails").Funcs(funcMap)
	var err error
	handler.Template, err = tmpl.ParseFiles("templates/db_table.html")
	return handler, err
}


type EmailPage struct {
    Emails   []db.Email
    Search   string
    Limit    string
    LimitInt int
    Offset   string
    OffsetInt int
}

func (h *TableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusNotImplemented)
        log.Println("Status Method Not Allowed")
        return
    }

    query := r.URL.Query()
    search := query.Get("search")
    limit := query.Get("limit")
    if limit == "" {
        limit = "50"
    }
    offset := query.Get("offset")
    if offset == "" {
        offset = "0"
    }

    // convert limit and offset to int
    limitInt, err := strconv.Atoi(limit)
    if err != nil {
        limitInt = 50
    }
    offsetInt, err := strconv.Atoi(offset)
    if err != nil {
        offsetInt = 0
    }

    // get emails
    content, err := h.DB.GetEmails(limitInt, offsetInt, search)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
        return
    }

    // wrap in EmailPage struct
    page := EmailPage{
        Emails:    content,
        Search:    search,
        Limit:     limit,
        LimitInt:  limitInt,
        Offset:    offset,
        OffsetInt: offsetInt,
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    err = h.Template.ExecuteTemplate(w, "emails", page)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
        return
    }
}

