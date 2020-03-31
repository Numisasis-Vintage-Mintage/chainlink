package service

import (
	"fmt"
	"time"

	"chainlink/ingester/client"
	"chainlink/ingester/util"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

// FeedsTracker is an interface for subscribing to new Chainlink aggregator feeds
type FeedsTracker interface {
	util.TickerService
	Subscribe() <-chan client.Aggregator
}

type feedsTracker struct {
	util.Ticker

	feedsByAddress map[common.Address]client.Aggregator
	feedsChannel   chan client.Aggregator

	eth       client.ETH
	feeds   client.Feeds
	networkId int
}

// NewFeedsTracker returns an instantiated instance of a FeedsTracker implementation
func NewFeedsTracker(eth client.ETH, feeds client.Feeds, networkId int) FeedsTracker {
	ag := &feedsTracker{
		feedsByAddress: map[common.Address]client.Aggregator{},
		feedsChannel:   make(chan client.Aggregator),
		eth:            eth,
		feeds:          feeds,
		networkId:      networkId,
	}
	ag.Ticker = util.Ticker{
		Ticker: time.NewTicker(time.Minute),
		Name:   "feedsTracker",
		Impl:   ag,
		Done:   make(chan struct{}),
		Exited: make(chan struct{}),
	}
	return ag
}

// Tick is the implementation of the ServiceTicker Tick interface
// calling the Feeds UI for the aggregator feeds, checking to see if there's any
// new feeds that aren't yet subscribed to
func (a *feedsTracker) Tick() {
	feeds, err := a.feeds.Feeds()
	if err != nil {
		log.Errorf("Error while calling feeds ui: %v", err)
	}
	for _, feed := range feeds {
		if feed.NetworkID != a.networkId || feed.ContractVersion != 2 { // Current monitor only supports Aggregator v2
			continue
		} else if agg, err := client.NewAggregator(a.eth, a.feeds, feed.Name, feed.ContractAddress); err != nil {
			log.Errorf("Error while creating new aggregator: %v", err)
		} else if err := a.add(feed.ContractAddress, agg); err != nil {
			a.WithService().WithFields(log.Fields{
				"name":      feed.Name,
				"address":   feed.ContractAddress.String(),
				"networkId": feed.NetworkID,
			}).Warnf("Ignoring aggregator contract due to latestRound error: %v", err)
			continue
		}
	}
}

// Subscribe returns the feeds channel so that any subscriber can receive new feeds
func (a *feedsTracker) Subscribe() <-chan client.Aggregator {
	return a.feedsChannel
}

func (a *feedsTracker) add(address common.Address, agg client.Aggregator) error {
	if _, ok := a.feedsByAddress[address]; ok {
		return nil
	}
	round, err := agg.LatestRound()
	if err != nil {
		return fmt.Errorf("invalid aggregator address: %+v", err)
	}

	a.WithService().WithFields(log.Fields{
		"address": agg.Address().String(),
		"name":    agg.Name(),
		"round":   round,
	}).Infof("New feed found")
	a.feedsByAddress[address] = agg
	a.feedsChannel <- agg
	return nil
}
