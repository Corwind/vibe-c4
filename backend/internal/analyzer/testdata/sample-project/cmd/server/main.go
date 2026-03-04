package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/example/sample-project/internal/handler"
	"github.com/example/sample-project/internal/service"
)

func main() {
	svc := service.New()
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Get("/health", h.HealthCheck)
	r.Post("/work", h.DoWork)

	http.ListenAndServe(":8080", r)
}
