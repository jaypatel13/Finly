package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	http   *http.Server
}

func NewServer(port string) *Server {
	r := mux.NewRouter()
	return &Server{
		router: r,
		http: &http.Server{
			Addr:    ":" + port,
			Handler: r,
		},
	}
}

func (s *Server) Start() <- chan error{
	errChan := make(chan error, 1)

	go func() {
		log.Printf("starting server on %s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	return errChan

}

func (s *Server) AddRoute(path string, handler http.HandlerFunc) {
	s.router.HandleFunc(path, handler)
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("shutting down server...")
	return s.http.Shutdown(ctx)
}
