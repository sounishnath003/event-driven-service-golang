package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
)

type Server struct {
	Port        int
	Log         *slog.Logger
	KafkaWriter *kafka.Writer

	mux *http.ServeMux
}

func NewServer(port int, kafkaWriter *kafka.Writer) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/create-post", CreatePostHandler)

	return &Server{
		mux:         mux,
		KafkaWriter: kafkaWriter,
		Port:        port,
		Log:         slog.Default(),
	}
}

func (s *Server) Start() {
	s.Log.Info("server is up and running on", slog.Int("Port", s.Port))
	panic(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.AddKafkaWriterContext(s.LogMiddleware(s.mux))))
}

// AddKafkaWriterContext adds the Kafka writer to the request context
func (s *Server) AddKafkaWriterContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "kafkaWriter", s.KafkaWriter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.Log.Info("Request", slog.String("Method", r.Method), slog.String("RequestUri", r.RequestURI), slog.String("RemoteAddr", r.RemoteAddr), slog.Float64("timeElapsed", time.Since(start).Seconds()))
	})
}
