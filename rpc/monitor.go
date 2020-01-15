package rpc

import (
	"log"
	"strconv"
	"time"

	"github.com/bolaxytools/tool-sdk"
)

var (
	defaultPeriod = 3 * time.Second
)

// Monitor bolaxy block scanner
type Monitor interface {
	// Start start monitor
	Start()
	// Stop stop monitor
	Stop()
}

// MonitorOpt bolaxy block scanner settings
type MonitorOpt func(monitor *blkMonitor)

// WithPeriod set the scan cycle interval
func WithPeriod(period time.Duration) MonitorOpt {
	return func(monitor *blkMonitor) {
		monitor.period = period
	}
}

// WithStartIndex set start scan index
func WithStartIndex(startIndex uint64) MonitorOpt {
	return func(monitor *blkMonitor) {
		monitor.startIndex = startIndex
	}
}

// NewBlkMonitor new block scan monitoring program
// The block scanner will use the Emitter to notify
// the transaction hash in the block and the Log details in the Receipt.
// Transaction`s event type is hex of txhash sdk.GenHashType(receipt.TransactionHash)
// Event`s event type is hexutil.Encode(crypto.Keccak256(contractAddr.Bytes(), eventSig.Bytes()))
// if event result returned and result.Success == true then has been officially written into the block
func NewBlkMonitor(e *sdk.Emitter, client *Client, opts ...MonitorOpt) Monitor {
	monitor := &blkMonitor{emitter: e, http: client, quit: make(chan struct{}, 1)}
	for _, opt := range opts {
		opt(monitor)
	}

	if monitor.period <= 0 {
		monitor.period = defaultPeriod
	}

	return monitor
}

type blkMonitor struct {
	http       *Client
	emitter    *sdk.Emitter
	period     time.Duration
	quit       chan struct{}
	startIndex uint64
}

// Start start monitor
func (m *blkMonitor) Start() {
	ticker := time.NewTicker(m.period)

	// 检查
	go func() {
		firstStarting := true
		var next uint64
		for {
			select {
			case <-ticker.C:
				if firstStarting && m.startIndex > 0 {
					log.Printf("blkMonitor first starting, and start index: %d\n", m.startIndex)
					next = m.startIndex
				} else {
					log.Printf("blkMonitor loop:\n")
					info, err := m.http.FetchChainInfo()
					if err != nil {
						log.Printf("blkMonitor fetch chain info failed. %v\n", err)
						return
					}
					x, _ := strconv.ParseInt(info.BlockHeight, 10, 64)

					next += 1
					if uint64(x) < next {
						next = uint64(x)
						log.Printf("blkMonitor current blk height: %d skip\n", next)
						continue
					}
				}

				firstStarting = false
				blk, err := m.http.FetchBlock(int(next))
				if err != nil {
					log.Printf("blkMonitor fetch blk failed. %v\n", err)
					return
				}
				txs, err := sdk.GetTransactionsFromBlk(blk)
				if err != nil {
					log.Printf("blkMonitor get txs failed. %v\n", err)
					return
				}

				for _, tx := range txs {
					receipt, err := m.http.FetchReceipt(tx.Hash)
					if err != nil {
						log.Printf("blkMonitor fetch receipt failed. %v\n", err)
						return
					}

					success := true
					if receipt.Status == 0 {
						success = false
					}

					var evt *sdk.Event
					res := &sdk.Result{
						Success:         success,
						ContractAddress: nil,
						IsLog:           false,
						Data:            nil,
						Topics:          nil,
					}

					evtTyp := sdk.GenHashType(receipt.TransactionHash)
					if receipt.To == nil {
						log.Printf("blkMonitor fire contract creation event, event type: %s, %s (%v)\n", evtTyp, receipt.ContractAddress.String(), success)
						res.ContractAddress = &receipt.ContractAddress
						evt = sdk.NewEvent(evtTyp, res)
					} else {
						log.Printf("blkMonitor fire tx event, %s (%v)\n", tx.Hash, success)
						log.Printf("blkMonitor ->%s, %s \n", tx.Hash, receipt.TransactionHash.String())
						evt = sdk.NewEvent(evtTyp, res)
					}
					m.emitter.Emit(evt)

					if len(receipt.Logs) > 0 && receipt.To != nil {
						for _, lg := range receipt.Logs {
							k := sdk.GenLogType(*receipt.To, lg.Topics[0])
							logRes := &sdk.Result{
								Success:         success,
								ContractAddress: nil,
								IsLog:           true,
								Data:            lg.Data,
								Topics:          lg.Topics,
							}

							log.Printf("blkMonitor fire log event, %s\n", k)
							e := sdk.NewEvent(k, logRes)
							m.emitter.Emit(e)
						}
					}
				}

			case <-m.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stop monitor
func (m *blkMonitor) Stop() {
	close(m.quit)
}
