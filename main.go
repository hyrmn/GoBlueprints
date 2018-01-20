package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type options struct {
	port string
}

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

var opt options

func init() {
	flag.StringVar(&opt.port, "p", os.Getenv("PORT"), "The default port to listen on")
	flag.Parse()

	if opt.port == "" {
		opt.port = "8080"
	}
}

func main() {
	http.Handle("/", &templateHandler{filename: "chat.html"})

	log.Printf("listening on port %v", opt.port)

	if err := http.ListenAndServe(":"+opt.port, nil); err != nil {
		log.Fatalf("ListenAndServe:%s", err)
	}
}
