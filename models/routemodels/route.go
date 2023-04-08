package routemodels

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

type RouteType string

const (
	// Route via REST API webhook
	REST RouteType = "REST"
)

type Route struct {
	ID         string    `json:"id,omitempty" bson:"id,omitempty" msgpack:",omitempty"`
	Type       RouteType `json:"type,omitempty" bson:"type,omitempty" msgpack:",omitempty"`
	WebhookURL string    `json:"webhook_url,omitempty" bson:"webhook_url,omitempty" msgpack:",omitempty"`
}

var (
	errInvalidRouteID    = errors.New("invalid route id")
	errInvalidRouteType  = errors.New("invalid route type. Only REST is allowed")
	errInvalidWebhookURL = errors.New("invalid webhook url")
)

func (r Route) Valid() error {
	if len(r.ID) == 0 {
		return errInvalidRouteID
	}
	if r.Type != REST {
		return errInvalidRouteType
	}
	if len(r.WebhookURL) == 0 {
		return errInvalidWebhookURL
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
