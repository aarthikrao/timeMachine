package routemodels

import (
	"encoding/json"
	"fmt"
)

type RouteType string

const (
	// Route via REST API webhook
	REST RouteType = "REST"
)

type Route struct {
	ID         string    `json:"id,omitempty" bson:"id,omitempty"`
	Type       RouteType `json:"type,omitempty" bson:"type,omitempty"`
	WebhookURL string    `json:"webhook_url,omitempty" bson:"webhook_url,omitempty"`
}

func (r *Route) Valid() error {
	if r.ID == "" {
		return fmt.Errorf("invalid route id")
	}
	if r.Type != REST {
		return fmt.Errorf("invalid route type. Only REST is allowed")
	}
	if r.WebhookURL == "" {
		return fmt.Errorf("invalid webhook url")
	}

	return nil
}

// TODO: Change to msgpack later
func (r *Route) ToBytes() ([]byte, error) {
	return json.Marshal(&r)
}

// GetRouteFromBytes returns the route struct from byte array
// TODO: Change to msgpack later
func GetRouteFromBytes(by []byte) (*Route, error) {
	var r Route
	err := json.Unmarshal(by, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
