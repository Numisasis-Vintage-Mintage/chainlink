package client

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	AggregatorV2Filename = "aggregation.v2.json"

	LatestAnswerFnName        = "latestAnswer"
	LatestRoundFnName         = "latestRound"
	OraclesInstanceVar        = "oracles"
	ResponseReceivedEventName = "ResponseReceived"
	NewRoundEventName         = "NewRound"
	MaxOracleCount            = 45

	UnmarshalEmptyStringError = "abi: attempting to unmarshall an empty string while arguments are expected"
)

type OracleMapping map[common.Address]string

type AggregatorOracle struct {
	Name    string
	Address common.Address
}

func (om OracleMapping) AggregatorOracle(address common.Address) *AggregatorOracle {
	name, ok := om[address]
	if !ok {
		name = "Unknown"
	}
	return &AggregatorOracle{
		Name:    name,
		Address: address,
	}
}

type Aggregator interface {
	Name() string
	Address() common.Address
	LatestAnswer() (*big.Int, error)
	LatestRound() (*big.Int, error)
	Oracles() (OracleMapping, error)
	SubscribeToNewRound(chan<- types.Log) (Subscription, error)
	UnmarshalNewRoundEvent(types.Log) (*NewRoundEvent, error)
	SubscribeToOracleAnswer(*big.Int, common.Address, chan<- types.Log) (Subscription, error)
	LogToTransaction(types.Log) (*types.Transaction, error)
}

type aggregator struct {
	name    string
	client  ETH
	feeds   Feeds
	abi     *abi.ABI
	address common.Address
}

func NewAggregator(client ETH, feeds Feeds, name string, address common.Address) (Aggregator, error) {
	aabi, err := client.ABI(AggregatorV2Filename)
	if err != nil {
		return nil, err
	}
	return &aggregator{
		name:    name,
		client:  client,
		feeds:  feeds,
		abi:     &aabi,
		address: address,
	}, nil
}

func (a *aggregator) Name() string {
	return a.name
}

func (a *aggregator) Address() common.Address {
	return a.address
}

func (a *aggregator) LatestAnswer() (*big.Int, error) {
	var answer *big.Int
	return answer, a.client.Call(a.address, a.abi, LatestAnswerFnName, &answer)
}

func (a *aggregator) LatestRound() (*big.Int, error) {
	var round *big.Int
	return round, a.client.Call(a.address, a.abi, LatestRoundFnName, &round)
}

func (a *aggregator) Oracles() (OracleMapping, error) {
	oracles := OracleMapping{}

	for i := int64(0); i < MaxOracleCount; i++ {
		var address common.Address
		if err := a.client.Call(a.address, a.abi, OraclesInstanceVar, &address, big.NewInt(i)); err != nil {
			if err.Error() == UnmarshalEmptyStringError {
				break
			}
			return oracles, err
		} else {
			oracles[address] = "Unknown"
		}
	}

	nodes, err := a.feeds.Nodes()
	if err != nil {
		return oracles, fmt.Errorf("error while calling feeds UI: %v", err)
	}
	for _, node := range nodes {
		if _, ok := oracles[node.Address]; ok {
			oracles[node.Address] = node.Name
		}
	}

	return oracles, nil
}

func (a *aggregator) SubscribeToNewRound(logChan chan<- types.Log) (Subscription, error) {
	e := a.abi.Events[NewRoundEventName]
	q := ethereum.FilterQuery{
		Addresses: []common.Address{a.address},
		Topics:    [][]common.Hash{{e.ID()}},
	}
	sub, err := a.client.SubscribeToLogs(logChan, q)
	if err != nil {
		return nil, err
	}
	return sub, err
}

func (a *aggregator) SubscribeToOracleAnswer(
	roundId *big.Int,
	oracle common.Address,
	logChan chan<- types.Log,
) (Subscription, error) {
	e := a.abi.Events[ResponseReceivedEventName]
	q := ethereum.FilterQuery{
		Addresses: []common.Address{a.address},
		Topics: [][]common.Hash{
			{e.ID()},
			{},
			{common.BigToHash(roundId)},
			{oracle.Hash()},
		},
	}
	sub, err := a.client.SubscribeToLogs(logChan, q)
	if err != nil {
		return nil, err
	}
	return sub, err
}

func (a *aggregator) UnmarshalNewRoundEvent(log types.Log) (*NewRoundEvent, error) {
	nr := &NewRoundEvent{}
	if len(log.Topics) == 3 {
		nr.RoundID = log.Topics[1].Big()
		nr.StartedBy = common.BytesToAddress(log.Topics[2].Bytes())
	} else {
		return nr, errors.New("invalid log type while un-marshaling new round, expected 3 topics")
	}
	return nr, nil
}

func (a *aggregator) LogToTransaction(log types.Log) (*types.Transaction, error) {
	return a.client.TransactionByHash(log.TxHash)
}
