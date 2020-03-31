package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

type Feed struct {
	ContractAddress common.Address `json:"contractAddress"`
	Name            string         `json:"name"`
	Pair            []string       `json:"pair"`
	Counter         int            `json:"counter"`
	ContractVersion int            `json:"contractVersion"`
	NetworkID       int            `json:"networkId"`
	History         bool           `json:"history"`
	Bollinger       bool           `json:"bollinger"`
	DecimalPlaces   int            `json:"decimalPlaces"`
	Multiply        string         `json:"multiply"`
}

type Node struct {
	Address   common.Address `json:"address"`
	Name      string         `json:"name"`
	NetworkId int            `json:"networkId"`
}

type Feeds interface {
	Feeds() ([]*Feed, error)
  Nodes() ([]*Node, error)
}

type feeds struct {
	client  *http.Client
	baseURL string
}

func NewFeeds(baseURL string) Feeds {
	return &feeds{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (f *feeds) Nodes() ([]*Node, error) {
	var nodes []*Node
	return nodes, f.do("/nodes.json", &nodes)
}

func (f *feeds) Feeds() ([]*Feed, error) {
	var feeds []*Feed
	return feeds, f.do("/feeds.json", &feeds)
}

func (f *feeds) do(endpoint string, obj interface{}) error {
	if resp, err := f.client.Get(fmt.Sprintf("%s%s", f.baseURL, endpoint)); err != nil {
		return err
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else if err := json.Unmarshal(b, obj); err != nil {
		return err
	}
	return nil
}
