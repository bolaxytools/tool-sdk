package rpc

import (
	"log"
	"strconv"
	"time"

	"github.com/bolaxytools/tool-sdk"
)

func NewBlkMonitor(period time.Duration, startIndex uint64, e *sdk.Emitter) *BlkMonitor {
	return &BlkMonitor{
		emitter:    e,
		period:     period,
		quit:       make(chan struct{}, 1),
		startIndex: startIndex,
	}
}

type BlkMonitor struct {
	emitter    *sdk.Emitter
	period     time.Duration
	quit       chan struct{}
	startIndex uint64
}

func (m *BlkMonitor) Start() {
	ticker := time.NewTicker(m.period)

	// 检查
	go func() {
		firstStarting := true
		var next uint64
		for {
			select {
			case <-ticker.C:
				if firstStarting && m.startIndex > 0 {
					log.Printf("BlkMonitor first starting, and start index: %d\n", m.startIndex)
					next = m.startIndex
				} else {
					log.Printf("BlkMonitor loop:\n")
					info, err := FetchChainInfo()
					if err != nil {
						log.Printf("BlkMonitor fetch chain info failed. %v\n", err)
						return
					}
					x, _ := strconv.ParseInt(info.BlockHeight, 10, 64)

					next += 1
					if uint64(x) < next {
						next = uint64(x)
						log.Printf("BlkMonitor current blk height: %d skip\n", next)
						continue
					}
				}

				firstStarting = false
				blk, err := FetchBlock(int(next))
				if err != nil {
					log.Printf("BlkMonitor fetch blk failed. %v\n", err)
					return
				}
				txs, err := sdk.GetTransactionsFromBlk(blk)
				if err != nil {
					log.Printf("BlkMonitor get txs failed. %v\n", err)
					return
				}

				for _, tx := range txs {
					receipt, err := FetchReceipt(tx.Hash)
					if err != nil {
						log.Printf("BlkMonitor fetch receipt failed. %v\n", err)
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
						log.Printf("BlkMonitor fire contract creation event, event type: %s, %s (%v)\n", evtTyp, receipt.ContractAddress.String(), success)
						res.ContractAddress = &receipt.ContractAddress
						evt = sdk.NewEvent(evtTyp, res)
					} else {
						log.Printf("BlkMonitor fire tx event, %s (%v)\n", tx.Hash, success)
						log.Printf("BlkMonitor ->%s, %s \n", tx.Hash, receipt.TransactionHash.String())
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

							log.Printf("BlkMonitor fire log event, %s\n", k)
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

func (m *BlkMonitor) Stop() {
	close(m.quit)
}
