package main

import (
	"example.com/pz4-todo/internal/task"
	myMW "example.com/pz4-todo/pkg/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	repo, err := task.NewRepo("tasks.json")
	if err != nil {
		log.Fatal("load tasks: ", err)
	}
	h := task.NewHandler(repo)
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(myMW.Logger)
	r.Use(myMW.SimpleCORS)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"route not found"}`))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	r.Route("/api", func(api chi.Router) {
		api.Mount("/tasks", h.Routes())
		api.Route("/v1", func(v1 chi.Router) { v1.Mount("/tasks", h.Routes()) })
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
