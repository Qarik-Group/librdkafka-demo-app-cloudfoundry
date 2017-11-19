package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func webHandlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}

func main() {
	cfApp, err := cfenv.Current()
	kafkaHostname := "localhost:9092"
	kafkaTopicName := "opencv-kafka-demo-status"
	if err == nil {
		services, err := cfApp.Services.WithTag("kafka")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot find service with label 'kafka': %v", err)
			os.Exit(1)
		}
		kafkaService := services[0]
		kafkaHostname, _ = kafkaService.CredentialString("hostname")
		kafkaTopicName, _ = kafkaService.CredentialString("topicName")
	} else {
		fmt.Fprintf(os.Stderr, "Not running inside Cloud Foundry. Assume local Kafka on localhost:9092\n")
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    kafkaHostname,
		"group.id":             "librdkafka-demo-app-cloudfoundry",
		"session.timeout.ms":   6000,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"}})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Kafka consumer %v\n", consumer)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	err = consumer.SubscribeTopics([]string{kafkaTopicName}, nil)

	go func() {
		run := true

		for run == true {
			select {
			case sig := <-sigchan:
				fmt.Printf("Caught signal %v: terminating\n", sig)
				run = false
			default:
				ev := consumer.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					fmt.Printf("%% Message on %s:\n%s\n",
						e.TopicPartition, string(e.Value))
				case kafka.PartitionEOF:
					fmt.Printf("%% Reached %v\n", e)
				case kafka.Error:
					fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
					run = false
				default:
					fmt.Printf("Ignored %v\n", e)
				}
			}
		}

		fmt.Printf("Closing consumer\n")
		consumer.Close()
		os.Exit(1)
	}()

	http.HandleFunc("/", webHandlerRoot)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	http.ListenAndServe(":"+port, nil)
}
