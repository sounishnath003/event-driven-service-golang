package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerConf struct {
	Port        int
	KafkaWriter *kafka.Writer
	MongoClient *mongo.Client
}

type Server struct {
	Log  *slog.Logger
	conf ServerConf
	mux  *http.ServeMux
}

func NewServer(serverConf ServerConf) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/create-post-without-kafka", CreatePostHandlerWithoutKafka)
	mux.HandleFunc("POST /api/create-post-with-kafka", CreatePostHandlerWithKafka)

	return &Server{
		mux:  mux,
		conf: serverConf,
		Log:  slog.Default(),
	}
}

func (s *Server) Start() {
	s.Log.Info("server is up and running on", slog.Int("Port", s.conf.Port))
	panic(http.ListenAndServe(fmt.Sprintf(":%d", s.conf.Port), s.AddCustomContexts(s.LogMiddleware(s.mux))))
}

// AddCustomContexts adds the Kafka writer to the request context
func (s *Server) AddCustomContexts(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "kafkaWriter", s.conf.KafkaWriter)
		ctx = context.WithValue(ctx, "mongoClient", s.conf.MongoClient)
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
