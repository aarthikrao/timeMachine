package routemodels

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

type RouteType string

const (
	// Route via Http API webhook
	Http  RouteType = "http"
	Kafka RouteType = "Kafka"
)

type Route struct {
	ID   string    `json:"id,omitempty" bson:"id,omitempty" msgpack:",omitempty"`
	Type RouteType `json:"type,omitempty" bson:"type,omitempty" msgpack:",omitempty"`

	// Incase of Http Route
	WebhookURL string `json:"webhook_url,omitempty" bson:"webhook_url,omitempty" msgpack:",omitempty"`

	// Incase of Kafka Route
	Topic string `json:"topic,omitempty" bson:"topic,omitempty" msgpack:",omitempty"`
	Host  string `json:"host,omitempty" bson:"host,omitempty" msgpack:",omitempty"`
}

var (
	ErrInvalidRouteID      = errors.New("invalid route id")
	ErrInvalidRouteType    = errors.New("invalid route type. Only REST is allowed")
	ErrInvalidWebhookURL   = errors.New("invalid webhook url")
	ErrInvalidKafkaDetails = errors.New("invalid kafka details")
)

func (r Route) Valid() error {
	if len(r.ID) == 0 {
		return ErrInvalidRouteID
	}
	switch r.Type {
	case Http:
		if len(r.WebhookURL) == 0 {
			return ErrInvalidWebhookURL
		}
	case Kafka:
		if len(r.Topic) == 0 || len(r.Host) == 0 {
			return ErrInvalidWebhookURL
		}
	default:
		return ErrInvalidRouteType
	}

	return nil
}

func (r Route) ToBytes() ([]byte, error) {
	encodedData, err := msgpack.Marshal(&r)
	return encodedData, err
}

func GetRouteFromBytes(encodedData []byte, r *Route) error {
	if encodedData == nil {
		return nil
	}
	return msgpack.Unmarshal(encodedData, &r)

}
