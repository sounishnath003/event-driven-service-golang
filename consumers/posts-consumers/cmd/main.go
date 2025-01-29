package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sounishnath003/posts-consumers/cmd/kafkaa"
	"github.com/sounishnath003/posts-consumers/cmd/mongodatabase"
	"go.mongodb.org/mongo-driver/mongo"
)

// Post struct which will be read from kafka topics
type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

func main() {
	// init the kafka reader streamer
	kafkaReader := kafkaa.NewKafkaReaderClient(kafka.ReaderConfig{
		Topic:   "create-posts",
		Brokers: []string{"localhost:9092"},
		Dialer: &kafka.Dialer{
			Timeout:   5 * time.Second,
			DualStack: true,
		},
	})

	// init the mongoclient for database operations
	mongoClient, err := mongodatabase.NewMongoDatabaseClient()
	if err != nil {
		panic(err)
	}

	// 1. Read messages using go routines
	// 2. convert to Post structuure
	// 3. save to mongo db database name `posts-with-kafka`
	// 4. Handler errors gracefully.

	// Read messages using goroutines
	go consumeMessages(kafkaReader, mongoClient)

	// Keep the main function running
	select {}
}

// consumeMessages reads messages from Kafka and saves them to MongoDB in batches
func consumeMessages(reader *kafka.Reader, mongoClient *mongo.Client) {
	batchSize := 100
	batchInterval := 5 * time.Second
	var posts []Post

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(posts) > 0 {
				log.Printf("Batch interval reached. Saving %d posts to database.", len(posts))
				if err := SavePostsBulkToDatabse(mongoClient, posts); err != nil {
					log.Printf("Failed to save posts to database: %v", err)
				} else {
					log.Printf("Successfully saved %d posts to database.", len(posts))
				}
				posts = nil // Reset the batch
			} else {
				log.Println("Batch interval reached but no posts to process.")
			}
		default:
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Failed to read message: %v", err)
				continue
			}

			// Convert to Post structure
			var post Post
			if err := json.Unmarshal(m.Value, &post); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// add additional fields from backend
			post.ID = time.Now().UnixMilli()
			post.CreatedAt = time.Now()
			post.UpdatedAt = time.Now()

			posts = append(posts, post)
			if len(posts) >= batchSize {
				log.Printf("Batch size reached. Saving %d posts to database.", len(posts))
				if err := SavePostsBulkToDatabse(mongoClient, posts); err != nil {
					log.Printf("Failed to save posts to database: %v", err)
				} else {
					log.Printf("Successfully saved %d posts to database.", len(posts))
				}
				posts = nil // Reset the batch
			}
		}
	}
}

// SavePostsBulkToDatabse saves the posts to the MongoDB database in a batch
func SavePostsBulkToDatabse(mongoClient *mongo.Client, posts []Post) error {
	collection := mongoClient.Database("posts-with-kafka").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var documents []interface{}
	for _, post := range posts {
		documents = append(documents, post)
	}

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	return nil
}
