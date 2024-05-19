package kafkaclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	connections map[string]*kafka.Writer
	mu          sync.RWMutex
}

func NewKafkaClient() *KafkaClient {
	return &KafkaClient{
		connections: make(map[string]*kafka.Writer),
	}
}

// Publish publishes a message to the specified Kafka topic.
// It takes the Kafka host, topic, payload, and key as parameters.
// If the connection to the Kafka host does not exist, it creates a new connection and caches it.
// Returns an error if there was a problem publishing the message.
func (kc *KafkaClient) Publish(kafkaHost, topic string, key []byte, value []byte) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	conn, err := kc.getConnection(kafkaHost)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	err = conn.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (kc *KafkaClient) getConnection(kafkaHost string) (*kafka.Writer, error) {
	kc.mu.RLock()

	conn, ok := kc.connections[kafkaHost]
	if ok {
		kc.mu.RUnlock()
		return conn, nil
	}

	// We dont have that connection. We will have to create a new one
	kc.mu.RUnlock()
	return kc.createConnection(kafkaHost)
}

func (kc *KafkaClient) createConnection(kafkaHost string) (*kafka.Writer, error) {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	// Check if the connection is already created by another goroutine
	conn, ok := kc.connections[kafkaHost]
	if ok {
		return conn, nil
	}

	conn = &kafka.Writer{
		Addr:     kafka.TCP(kafkaHost),
		Balancer: &kafka.LeastBytes{},
	}
	kc.connections[kafkaHost] = conn
	return conn, nil
}
