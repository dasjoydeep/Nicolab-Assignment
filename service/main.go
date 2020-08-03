package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handlers := http.NewServeMux()
	handlers.HandleFunc("/", handleEvent)


	log.Printf("Listening on port: %s", port)

	err := http.ListenAndServe(":"+port,
		requestLogger(handlers),
	)
	if err != nil {
		log.Printf("Failed to serve: %v", err)
	}
}

type attributes struct {
	BucketId                string    `json:"bucketId"`
	EventTime               time.Time `json:"eventTime"`
	EventType               string    `json:"eventType"`
	NotificationConfig      string    `json:"notificationConfig"`
	ObjectGeneration        string    `json:"objectGeneration"`
	ObjectId                string    `json:"objectId"`
	PayloadFormat           string    `json:"payloadFormat"`
	OverwroteGeneration     string    `json:"overwroteGeneration"`
	OverwrittenByGeneration string    `json:"overwrittenByGeneration"`
}

type message struct {
	Attributes  *attributes `json:"attributes"`
	Data        string      `json:"data"`
	MessageID   string      `json:"messageId"`
	PublishTime time.Time   `json:"publishTime"`
}

type cloudStorageEvent struct {
	Kind                    string            `json:"kind"`
	ID                      string            `json:"id"`
	SelfLink                string            `json:"selfLink"`
	Name                    string            `json:"name"`
	Bucket                  string            `json:"bucket"`
	Generation              string            `json:"generation"`
	Metageneration          string            `json:"metageneration"`
	ContentType             string            `json:"contentType"`
	TimeCreated             time.Time         `json:"timeCreated"`
	Updated                 time.Time         `json:"updated"`
	StorageClass            string            `json:"storageClass"`
	TimeStorageClassUpdated time.Time         `json:"timeStorageClassUpdated"`
	Size                    string            `json:"size"`
	MD5Hash                 string            `json:"md5Hash"`
	MediaLink               string            `json:"mediaLink"`
	Metadata                map[string]string `json:"metadata"`
	Crc32c                  string            `json:"crc32c"`
	ETag                    string            `json:"etag"`
}

type BucketNotificationMessage struct {
	EventType               string            `json:"eventType"`
	Metadata                map[string]string `json:"metadata"`
	Size                    int               `json:"size"`
	MD5Hash                 string            `json:"md5Hash"`
	TimeCreated             time.Time         `json:"timeCreated"`
	Updated                 time.Time         `json:"updated"`
	Bucket                  string            `json:"bucket"`
	Generation              int64             `json:"generation"`
	Metageneration          int               `json:"metageneration"`
	Name                    string            `json:"name"`
	PublishTime             time.Time         `json:"publishTime"`
	MessageID               string            `json:"messageId"`
	OverwroteGeneration     string            `json:"overwroteGeneration"`
	OverwrittenByGeneration string            `json:"overwrittenByGeneration"`
}

type pubSubMessage struct {
	Message      *message `json:"message"`
	Subscription string   `json:"subscription"`
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		event, err := parseEvent(r.Body)
		if err != nil {
			log.Printf("Failed to parse event body: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Println(event)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func parseEvent(readCloser io.ReadCloser) (*BucketNotificationMessage, error) {
	decoder := json.NewDecoder(readCloser)
	var pubSubMessage pubSubMessage
	err := decoder.Decode(&pubSubMessage)
	if err != nil {
		return nil, err
	}

	jsonMessageData, err := base64.
		StdEncoding.DecodeString(pubSubMessage.Message.Data)
	if err != nil {
		return nil, err
	}

	var messageData cloudStorageEvent
	err = json.Unmarshal(jsonMessageData, &messageData)
	if err != nil {
		return nil, err
	}

	notification := &BucketNotificationMessage{
		EventType:               pubSubMessage.Message.Attributes.EventType,
		Metadata:                messageData.Metadata,
		MD5Hash:                 messageData.MD5Hash,
		TimeCreated:             messageData.TimeCreated,
		Updated:                 messageData.Updated,
		Bucket:                  messageData.Bucket,
		Name:                    messageData.Name,
		PublishTime:             pubSubMessage.Message.PublishTime,
		MessageID:               pubSubMessage.Message.MessageID,
		OverwroteGeneration:     pubSubMessage.Message.Attributes.OverwroteGeneration,
		OverwrittenByGeneration: pubSubMessage.Message.Attributes.OverwrittenByGeneration,
	}

	// It's OK if these integer parse events fail
	size, _ := strconv.Atoi(messageData.Size)
	generation, _ := strconv.ParseInt(messageData.Generation, 10, 64)
	metageneration, _ := strconv.Atoi(messageData.Metageneration)
	notification.Size = size
	notification.Generation = generation
	notification.Metageneration = metageneration

	return notification, nil
}

/* Log all incoming requests */
func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		log.Printf(
			"%s\t\t%s\t\t%v",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}