package handlers

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"mail-server/db"
	"net/http"
	"strconv"
	"strings"
	"time"
    "regexp"
)

type TableHandler struct {
	DB       *db.DB
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
			case time.Time:
				text = v.Format("2006-01-02 15:04:05")
			default:
				text = fmt.Sprintf("%v", v)
			}
			if search == "" {
				return template.HTML(html.EscapeString(text))
			}
			escaped := html.EscapeString(text)

			pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(search))

			highlighted := pattern.ReplaceAllStringFunc(escaped, func(m string) string {
        		return fmt.Sprintf(`<span class="highlight">%s</span>`, html.EscapeString(m))
			})
			return template.HTML(highlighted)

		},
		"add":      func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
		"div":      func(a, b int) int { return a / b },
		"slice":    func(vals ...int) []int { return vals },
		"sliceStr": func(vals ...string) []string { return vals },
		"inSlice": func(val string, list []string) bool {
			for _, v := range list {
				if v == val {
					return true
				}
			}
			return false
		},
		"join": func(vals []string, sep string) string {
			if len(vals) == 0 {
				return ""
			}
			return strings.Join(vals, sep)
		},
	}

	tmpl := template.New("emails").Funcs(funcMap)
	var err error
	handler.Template, err = tmpl.ParseFiles("templates/db_table.html")
	return handler, err
}

type EmailPage struct {
	Emails          []db.Email
	Search          string
	Limit           int
	Offset          int
	SelectedColumns []string
}

func (h *TableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	query := r.URL.Query()
	search := query.Get("search")

	limitStr := query.Get("limit")
	if limitStr == "" {
		limitStr = "50"
	}
	offsetStr := query.Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}

	// Convert limit and offset to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Extract selected columns from query
	selectedColumns := query["columns"]
	// Optional: Validate column names against a whitelist to prevent injection
	allowed := map[string]bool{
		"id": true, "from": true, "to": true, "subject": true,"date":true, "reason": true,
		"body": true, "registrarid": true, "sent": true, "status": true,
	}
	var validColumns []string
	for _, col := range selectedColumns {
		if allowed[strings.ToLower(col)] {
			validColumns = append(validColumns, strings.ToLower(col))
		}
	}
	// Get emails using selected columns
	content, err := h.DB.GetEmails(limit, offset, search, validColumns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Wrap in EmailPage struct
	page := EmailPage{
		Emails:          content,
		Search:          search,
		Limit:           limit,
		Offset:          offset,
		SelectedColumns: selectedColumns,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = h.Template.ExecuteTemplate(w, "emails", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}
