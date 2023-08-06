package routemodels

import (
	"bytes"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

func TestToBytes(t *testing.T) {
	var r Route
	encodeData, err := r.ToBytes()
	if err != nil {
		t.Errorf("Error encoding data: %v\n.", err)
	}
	t.Log("Encoded Data:", string(encodeData))

	msgPackMarshaledData, err := msgpack.Marshal(&r)
	if err != nil {
		t.Errorf("Error marshalling with msgPack: %v.\n", err)
	}
	t.Logf("Encode Data: %v\nMsgPackMarhsaled Data: %v.\n",
		string(encodeData),
		string(msgPackMarshaledData),
	)
	if !bytes.Equal(encodeData, msgPackMarshaledData) {
		t.Errorf("Error encoded data not match: got: %v, want: %v.\n", encodeData, msgPackMarshaledData)
	}
}

func TestGetRouteFromBytes(t *testing.T) {
	var n []byte
	var r Route
	err := GetRouteFromBytes(n, &r)
	if err != nil {
		t.Errorf("Error getting route from encoded data: %v.\n", err)
	}
	t.Log("Decoded Route", r)

	mpBytes, err := msgpack.Marshal(
		&Route{
			ID:         "something",
			Type:       RouteType("foo"),
			WebhookURL: "https://google.com",
		})
	if err != nil {
		t.Errorf("Error marshalling mpBytes with msgpack: %v.\n", err)
	}

	var decodedRoute Route
	err = GetRouteFromBytes(mpBytes, &decodedRoute)
	if err != nil {
		t.Errorf("Error getting route from encoded data: %v.\n", err)
	}

	if !(decodedRoute.ID == "something" &&
		decodedRoute.Type == RouteType("foo") &&
		decodedRoute.WebhookURL == "https://google.com") {
		t.Errorf("Error wrong basic decoding.\n")
	}

	rand.NewSource(time.Now().UnixNano())
	initialRoute := Route{
		ID:         strconv.Itoa(rand.Int()),
		WebhookURL: strconv.Itoa(rand.Int()),
		Type:       RouteType(strconv.Itoa(rand.Int())),
	}
	mpBytes, err = msgpack.Marshal(&initialRoute)
	if err != nil {
		t.Errorf("Error marshalling mpBytes with msgpack: %v.\n", err)
	}

	err = GetRouteFromBytes(mpBytes, &decodedRoute)
	if err != nil {
		t.Errorf("Error getting route from encoded data: %v.\n", err)
	}
	if !(decodedRoute.ID == initialRoute.ID &&
		decodedRoute.Type == initialRoute.Type &&
		decodedRoute.WebhookURL == initialRoute.WebhookURL) {
		t.Errorf("Error wrong random decoding.\n")
	}
}

func TestValid(t *testing.T) {
	var r Route
	err := r.Valid()
	if err != errInvalidRouteID {
		t.Errorf("Error empty ids shouldn't be allowed.\n")
	}

	r.ID = "a"
	err = r.Valid()
	if err == errInvalidRouteID {
		t.Errorf("Error non-empty ids should be allowed.\n")
	}

	if err != errInvalidRouteType {
		t.Errorf("Error invalid route type shouldn't be allowed.\n")
	}

	r.Type = REST
	err = r.Valid()
	if err == errInvalidRouteType {
		t.Errorf("Error REST type route should be allowed.\n")
	}

	if err != errInvalidWebhookURL {
		t.Errorf("Error invalid url shouldn't be allowed.\n")
	}

	r.WebhookURL = "a"
	err = r.Valid()
	if err == errInvalidWebhookURL {
		t.Errorf("Error valid webhook should be allowed.\n")
	}
}
