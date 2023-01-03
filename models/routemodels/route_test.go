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
		t.Fail()
		return
	}
	t.Log("Encoded Data:", string(encodeData))
	msgPackMarshaledData, err := msgpack.Marshal(&r)
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	t.Logf("Encode Data: %v\nMsgPackMarhsaled Data: %v",
		string(encodeData),
		string(msgPackMarshaledData))
	if !bytes.Equal(encodeData, msgPackMarshaledData) {
		t.Fail()

	}

}

func TestGetRouteFromBytes(t *testing.T) {
	var n []byte
	var r Route
	err := GetRouteFromBytes(n, &r)
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	t.Log("Decoded Route", r)

	mpBytes, err := msgpack.Marshal(&Route{ID: "something", Type: RouteType("foo"), WebhookURL: "https://google.com"})
	if err != nil {
		t.Fail()
		t.Log(err)
		return
	}
	var decodedRoute Route
	err = GetRouteFromBytes(mpBytes, &decodedRoute)
	if err != nil {
		t.Fail()
		t.Log(err)
		return
	}
	if !(decodedRoute.ID == "something" && decodedRoute.Type == RouteType("foo") && decodedRoute.WebhookURL == "https://google.com") {
		t.Fail()
		t.Log("Wrong Basic Decoding")
	}
	rand.Seed(time.Now().UnixMilli())
	intialRoute := Route{ID: strconv.Itoa(rand.Int()),
		WebhookURL: strconv.Itoa(rand.Int()),
		Type:       RouteType(strconv.Itoa(rand.Int())),
	}
	mpBytes, err = msgpack.Marshal(&intialRoute)
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	err = GetRouteFromBytes(mpBytes, &decodedRoute)
	if err != nil {
		t.Fail()
		t.Log(err)
		return

	}
	if !(decodedRoute.ID == intialRoute.ID && decodedRoute.Type == intialRoute.Type && decodedRoute.WebhookURL == intialRoute.WebhookURL) {
		t.Fail()
		t.Log("Wrong Random Decoding")
	}
}

func TestValid(t *testing.T) {
	var r Route
	err := r.Valid()
	if err != errInvalidRouteID {
		t.Fail()
		t.Log("Empty ids shouldn't be allowed")
		return
	}
	r.ID = "a"
	err = r.Valid()
	if err == errInvalidRouteID {
		t.Fail()
		t.Log("Non-empty ids should be allowed")
		return
	}
	if err != errInvalidRouteType {
		t.Fail()
		t.Log("Invalid route type shouldn't allowed")
	}
	r.Type = REST
	err = r.Valid()
	if err == errInvalidRouteType {
		t.Fail()
		t.Log("REST type route should be allowed")
	}

	if err != errInvalidWebhookURL {
		t.Fail()
		t.Log("Invalid url shouldn't be allowed")
	}
	r.WebhookURL = "a"
	err = r.Valid()
	if err == errInvalidWebhookURL {
		t.Fail()
		t.Log("Valid webhook should be allowed")
	}
}
