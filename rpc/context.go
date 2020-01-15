package rpc

import (
	"math/big"

	"github.com/bolaxy/common"
	ethTypes "github.com/bolaxy/eth/types"
)

var (
	defaultHost = "http://localhost:8080"
)

const (
	callSvcUrl  = "/call"
	blkSvcUrl   = "/block/"
	infoUrl     = "/info"
	receiptUrl  = "/tx/"
	transferUrl = "/rawtx"
	accountUrl  = "/account/"
)

/* ChainMeta 链当前的状态，可以查询当前块高
{
    "consensus_events": "4025",
    "consensus_transactions": "230",
    "events_per_second": "0.00",
    "id": "3615552456",
    "last_block_index": "210",
    "last_consensus_round": "463",
    "moniker": "node1",
    "num_peers": "3",
    "round_events": "0",
    "rounds_per_second": "0.00",
    "state": "Babbling",
    "sync_rate": "1.00",
    "transaction_pool": "0",
    "type": "babble",
    "undetermined_events": "20"
}
*/
type ChainMeta struct {
	EventsNum          string `mapstructure:"consensus_events"`
	EventRate          string `mapstructure:"events_per_second"`
	TransactionsNum    string `mapstructure:"consensus_transactions"`
	Id                 string `mapstructure:"id"`
	BlockHeight        string `mapstructure:"last_block_index"`
	ConsensusRound     string `mapstructure:"last_consensus_round"`
	NodeName           string `mapstructure:"moniker"`
	PeersNum           string `mapstructure:"num_peers"`
	RoundEvents        string `mapstructure:"round_events"`
	RoundRate          string `mapstructure:"rounds_per_second"`
	State              string `mapstructure:"state"`
	SyncRate           string `mapstructure:"sync_rate"`
	InPoolNum          string `mapstructure:"transaction_pool"`
	Type               string `mapstructure:"type"`
	UnDeterminedEvents string `mapstructure:"undetermined_events"`
}

type RawTxRes struct {
	TxHash       string `mapstructure:"txHash"`
	ContractAddr string `mapstructure:"contractAddr"`
}

type JsonAccount struct {
	Address string   `mapstructure:"address"`
	Balance *big.Int `mapstructure:"balance"`
	Nonce   uint64   `mapstructure:"nonce"`
	Code    string   `mapstructure:"bytecode"`
}

type JsonReceipt struct {
	Root              common.Hash     `mapstructure:"root"`
	TransactionHash   common.Hash     `mapstructure:"transactionHash"`
	From              common.Address  `mapstructure:"from"`
	To                *common.Address `mapstructure:"to"`
	GasUsed           uint64          `mapstructure:"gasUsed"`
	CumulativeGasUsed uint64          `mapstructure:"cumulativeGasUsed"`
	ContractAddress   common.Address  `mapstructure:"contractAddress"`
	Logs              []*ethTypes.Log `mapstructure:"logs"`
	LogsBloom         ethTypes.Bloom  `mapstructure:"logsBloom"`
	Status            uint64          `mapstructure:"status"`
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From         common.Address           `json:"from"`
	To           *common.Address          `json:"to"`
	Gas          uint64                   `json:"gas"`
	GasPrice     *big.Int                 `json:"gasPrice"`
	Value        *big.Int                 `json:"value"`
	Data         string                   `json:"data"`
	Nonce        *uint64                  `json:"nonce"`
	FromChainid  string                   `json:"fromchainid"`
	ToChainid    string                   `json:"tochainid"`
	TxType       ethTypes.TransactionType `json:"txtype"`
	FromTxhash   *common.Hash             `json:"fromtxhash"`
	ContractAddr common.Address           `json:"contractaddr"` //要转到的地址,to为智能合约地址
	Raw          []byte
}

type JsonCallRes struct {
	Data string `mapstructure:"data"`
}