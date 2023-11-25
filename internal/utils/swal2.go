package utils

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed swal2.html
var swal2Template string

func Swal2Response(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("main").Parse(swal2Template)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	err = tmpl.Execute(w, map[string]any{
		"Body": template.HTML(body),
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
