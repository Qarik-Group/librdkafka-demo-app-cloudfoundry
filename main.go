package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func webHandlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	cfApp, err := cfenv.Current()
	statusHostname := "localhost:9092"
	statusTopicName := "opencv-kafka-demo-status"
	if err == nil {
		statusTopicService, err := cfApp.Services.WithName("status-topic")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot find service name 'status-topic': %v", err)
			os.Exit(1)
		}
		statusHostname, _ = statusTopicService.CredentialString("hostname")
		statusTopicName, _ = statusTopicService.CredentialString("topicName")
	} else {
		fmt.Fprintf(os.Stderr, "Not running inside Cloud Foundry. Assume local Kafka on localhost:9092")
	}

	statusP, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": statusHostname})
	if err != nil {
		fmt.Printf("Failed to create 'status-topic' producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created 'static-topic' producer %v\n", statusP)

	// Optional delivery channel, if not specified the Producer object's
	// .Events channel is used.
	deliveryChan := make(chan kafka.Event)

	value := "{\"status\":\"starting\", \"client\": \"posted_images_to_kafka\", \"language\": \"golang\"}"
	err = statusP.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &statusTopicName, Partition: kafka.PartitionAny}, Value: []byte(value)}, deliveryChan)
	if err != nil {
		fmt.Printf("Failed .Produce for'status-topic' message: %s\n", err)
		os.Exit(1)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)

	http.HandleFunc("/", webHandlerRoot)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	http.ListenAndServe(":"+port, nil)
}
