package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Initialises fetches other server location from the seed node,
// prints information regarding their health and the current leader
func initialise(seedNode string) (leaderAddress string, err error) {
	serverLocation, err := getOtherServerLocations(seedNode)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	ServerLocationLatest = serverLocation
	health := getHealth(serverLocation.Servers)

	var healthTable [][]interface{}
	for _, node := range serverLocation.Servers {
		var healthy emoji = Red
		if health[node.Address] {
			healthy = Green
		}

		healthTable = append(healthTable,
			[]interface{}{
				node.ID,
				node.Address,
				healthy,
			})

	}

	printTable(
		[]interface{}{"NodeID", "Address", "Health"},
		healthTable,
	)

	return getHTTPAddressFromRaft(serverLocation.Leader), nil
}

// returns the health of all the servers. Only looks for http status 200
func getHealth(sr []ServerAddress) (health map[string]bool) {
	health = make(map[string]bool)
	for _, server := range sr {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Address))
		if err != nil {
			health[server.Address] = false
			continue
		}
		if resp.StatusCode != http.StatusOK {
			health[server.Address] = false
			continue
		}

		health[server.Address] = true
	}

	return health
}

// getOtherServerLocations returns the address of other timeMachine nodes.
// It also fetches the information regarding the current leader node
func getOtherServerLocations(seedNode string) (*ServerLocation, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/cluster/servers", seedNode))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrNotSuccess
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sr ServerLocation
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}

	for i := range sr.Servers {
		sr.Servers[i].Address = getHTTPAddressFromRaft(sr.Servers[i].Address) // Validate
	}

	return &sr, nil
}

// Configure command to initilise the number of shards and replicas
func configure(shards, replicas int) error {
	// Define the data structure to be converted to JSON
	data := struct {
		Shards   int `json:"shards"`
		Replicas int `json:"replicas"`
	}{
		Shards:   shards,
		Replicas: replicas,
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/cluster/configure", LeaderAddress)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return err
		}

		if errResp.Error != "" {
			return errors.New(errResp.Error)
		}
	}

	return nil
}
