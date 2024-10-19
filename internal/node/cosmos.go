// internal/node/cosmos.go

package node

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type CosmosNode struct {
	RPCURL string
}

func NewCosmosNode(rpcURL string) *CosmosNode {
	return &CosmosNode{RPCURL: rpcURL}
}

func (cn *CosmosNode) IsSynced() (bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(cn.RPCURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)

	if res, ok := result["result"]; ok {
		if syncInfo, ok := res.(map[string]interface{})["sync_info"]; ok {
			if catchingUp, ok := syncInfo.(map[string]interface{})["catching_up"].(bool); ok {
				return !catchingUp, nil
			}
		}
	}

	return false, errors.New("unexpected response from Cosmos node")
}
