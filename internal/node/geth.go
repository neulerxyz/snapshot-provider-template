// internal/node/geth.go

package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type GethNode struct {
	RPCURL string
}

func NewGethNode(rpcURL string) *GethNode {
	return &GethNode{RPCURL: rpcURL}
}

func (gn *GethNode) IsSynced() (bool, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_syncing",
		"params":  []interface{}{},
		"id":      1,
	}
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(gn.RPCURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)

	if syncing, ok := result["result"]; ok {
		if syncing == false {
			return true, nil
		}
		return false, nil
	}

	return false, errors.New("unexpected response from Geth node")
}
