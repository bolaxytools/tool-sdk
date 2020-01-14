package rpc

import (
	"math/big"

	"github.com/bolaxy/common"
	ethTypes "github.com/bolaxy/eth/types"
)

var (
	host = "http://localhost:8080"
)

const (
	callSvcUrl  = "/call"
	blkSvcUrl   = "/block/"
	infoUrl     = "/info"
	receiptUrl  = "/tx/"
	transferUrl = "/rawtx"
	accountUrl  = "/account/"
)

func SetHost(h string) {
	host = h
}

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
	EventsNum          string `json:"consensus_events"`
	EventRate          string `json:"events_per_second"`
	TransactionsNum    string `json:"consensus_transactions"`
	Id                 string `json:"id"`
	BlockHeight        string `json:"last_block_index"`
	ConsensusRound     string `json:"last_consensus_round"`
	NodeName           string `json:"moniker"`
	PeersNum           string `json:"num_peers"`
	RoundEvents        string `json:"round_events"`
	RoundRate          string `json:"rounds_per_second"`
	State              string `json:"state"`
	SyncRate           string `json:"sync_rate"`
	InPoolNum          string `json:"transaction_pool"`
	Type               string `json:"type"`
	UnDeterminedEvents string `json:"undetermined_events"`
}

type RawTxRes struct {
	TxHash       string `json:"txHash"`
	ContractAddr string `json:"contractAddr"`
}

type JsonAccount struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
	Nonce   uint64   `json:"nonce"`
	Code    string   `json:"bytecode"`
}

type JsonReceipt struct {
	Root              common.Hash     `json:"root"`
	TransactionHash   common.Hash     `json:"transactionHash"`
	From              common.Address  `json:"from"`
	To                *common.Address `json:"to"`
	GasUsed           uint64          `json:"gasUsed"`
	CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
	ContractAddress   common.Address  `json:"contractAddress"`
	Logs              []*ethTypes.Log `json:"logs"`
	LogsBloom         ethTypes.Bloom  `json:"logsBloom"`
	Status            uint64          `json:"status"`
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
	Data string `json:"data"`
}