package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

// CreatePostHandlerWithKafka handles the create posts requests publishes without kafka
func CreatePostHandlerWithoutKafka(w http.ResponseWriter, r *http.Request) {
	var createPostPayload Post
	err := json.NewDecoder(r.Body).Decode(&createPostPayload)

	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// // add additional fields from backend
	createPostPayload.ID = time.Now().UnixMilli()
	createPostPayload.CreatedAt = time.Now()
	createPostPayload.UpdatedAt = time.Now()

	// Create a Database entry of the post.
	mongoClient, err := GetMongoClientFromContext(r.Context())
	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Save post to database
	err = savePostToDatabase(mongoClient, createPostPayload)
	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ApiJsonResponse(w, http.StatusOK, map[string]any{
		"post":    createPostPayload,
		"message": "post created",
	})
}

// savePostToDatabase saves the post to the MongoDB database
func savePostToDatabase(mongoClient *mongo.Client, createPostPayload Post) error {
	collection := mongoClient.Database("posts-without-kafka").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, createPostPayload)
	if err != nil {
		return err
	}

	return nil
}

// CreatePostHandlerWithKafka handles the create posts requests publishes without kafka
func CreatePostHandlerWithKafka(w http.ResponseWriter, r *http.Request) {
	var createPostPayload Post
	err := json.NewDecoder(r.Body).Decode(&createPostPayload)

	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// postBytes
	postBytes, err := json.Marshal(createPostPayload)
	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// build a kafka msg
	kmsg := kafka.Message{
		Value: postBytes,
	}

	// define kafka
	kafkaWriter, err := GetKafkaWriterFromContext(r.Context())
	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Publish msg to kafka
	err = kafkaWriter.WriteMessages(r.Context(), kmsg)
	if err != nil {
		ApiJsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ApiJsonResponse(w, http.StatusOK, map[string]string{
		"message":   "post created",
	})
}

// ApiJsonResponse sends a JSON response with the given status code and data
func ApiJsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetKafkaWriterFromContext retrieves the Kafka writer from the request context
func GetKafkaWriterFromContext(ctx context.Context) (*kafka.Writer, error) {
	writer, ok := ctx.Value("kafkaWriter").(*kafka.Writer)
	if !ok {
		return nil, errors.New("kafka writer instance is not initialized in posts producers")
	}
	return writer, nil
}

// GetMongoClientFromContext retrieves the Kafka writer from the request context
func GetMongoClientFromContext(ctx context.Context) (*mongo.Client, error) {
	mongoClient, ok := ctx.Value("mongoClient").(*mongo.Client)
	if !ok {
		return nil, errors.New("mongo db client instance is not initialized in posts producers")
	}
	return mongoClient, nil
}
