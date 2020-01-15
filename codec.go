package sdk

import (
	"encoding/json"
	"errors"

	"github.com/bolaxy/common"
	"github.com/bolaxy/core/types"
	"github.com/bolaxy/crypto"
	ethType "github.com/bolaxy/eth/types"
	"github.com/bolaxy/rlp"
)

// Transaction the display transaction
type Transaction struct {
	Hash     string // Hash hex hash string e.g 0xe349b239e5b2fbb8ebe96556c3caa4c2b419f9a51af5e497bba0735c88a48b6d
	From     string // From hex address string e.g. 0x599D7ABDB0A289F85aACA706b55D1b96cc07f348
	To       string // To hex address string e.g. 0x599D7ABDB0A289F85aACA706b55D1b96cc07f348
	Value    string // Value hex string for value
	Gas      uint64 // Gas it must greater than 21001 when send to server
	GasPrice string // GasPrice hex string for price
	Data     []byte // Data raw input data or some other informations
	Nonce    uint64 // Nonce nonce value
}

// GetTransactions 解析区块中的交易列表
// 参数
// 	serialized string 格式如下：
// {
//    "Body": {
//        "Index": 1,
//        "RoundReceived": 7,
//        "StateHash": "SH3WGKb4GhTerV7HN+TZELcBr22iqqtZVrt6f5U3xoI=",
//        "FrameHash": "xCMMmler6e5im6MYwJBSMIYKO2k75e0joYk04aS6t4w=",
//        "PeersHash": "eUmEw4HSpUdibDRG/4Ap686Lv0ui+URMCvezBEPdsuE=",
//        "Transactions": [
//            "+QKJAQqDD0JAgIC5Ahg2MDgwNjA0MDUyMzQ4MDE1NjEwMDEwNTc2MDAwODBmZDViNTA2MDAwODA1NDYwMDE2MDAxNjBhMDFiMDMxOTE2MzMxNzkwNTU2MGRiODA2MTAwMzE2MDAwMzk2MDAwZjNmZTYwODA2MDQwNTIzNDgwMTU2MDBmNTc2MDAwODBmZDViNTA2MDA0MzYxMDYwMjg1NzYwMDAzNTYwZTAxYzgwNjM4ZGE1Y2I1YjE0NjA3NTU3NWI2MDQwODA1MTYyNDYxYmNkNjBlNTFiODE1MjYwMjA2MDA0ODIwMTUyNjAxYTYwMjQ4MjAxNTI3ZjUzNjU2ZTY0MjA2MjYxNjM2YjIwNjU3NDY4NjU3MjIwNzM2NTZlNzQyMDc0NmYyMDZkNjUwMDAwMDAwMDAwMDA2MDQ0ODIwMTUyOTA1MTkwODE5MDAzNjA2NDAxOTBmZDViNjA3YjYwOTc1NjViNjA0MDgwNTE2MDAxNjAwMTYwYTAxYjAzOTA5MjE2ODI1MjUxOTA4MTkwMDM2MDIwMDE5MGYzNWI2MDAwNTQ2MDAxNjAwMTYwYTAxYjAzMTY4MTU2ZmVhMjY1NjI3YTdhNzIzMTU4MjBkNDg5ZGM0MDc4Zjk1MDQ2NDlkMGM4ZjA4M2QzZDE1Yjg4YjNjODBhOTI3YWJhYzM3NDNmMGQzMTI0NmExMzUyNjQ3MzZmNmM2MzQzMDAwNTBjMDAzMoCAoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJaC3AoVcyTlpPGUI5LPQZw8O6fsOJiH+IVhWyRKbm/4oGqBhPHh00wpBZpEp2YhDvekbvbyPxRB1LzjV+0faun42LQ=="
//        ],
//        "InternalTransactions": [],
//        "InternalTransactionReceipts": []
//    },
//    "Signatures": {
//        "0X02C42176372A9C30F29F04B99E2599F329F68B5F079D32834DAD91203E41529694": "0x3bd960d3280a69cda35d7a1ac75ab5b976424bdbcb69834813bb5c31af1f4f9b1b0cd6b36ceade9e5767b684c759b848043eb972b10574b81dccd212685a256500"
//    }
// }
//
// 	从节点API返回的区块数据。JSON格式。
// 返回结果
// 	[]*Transaction 本SDK下的交易结构。将信息内容解析成可读格式。
func GetTransactions(serialized string) ([]*Transaction, map[string]string, error) {
	if len(serialized) == 0 {
		return nil, nil, errors.New("wrong input param")
	}

	var block types.Block
	if err := json.Unmarshal([]byte(serialized), &block); err != nil {
		return nil, nil, err
	}

	trans := block.Body.Transactions
	transactions := make([]*Transaction, 0, len(trans))
	for _, tran := range trans {
		var t ethType.Transaction
		if err := rlp.DecodeBytes(tran, &t); err != nil {
			return nil, nil, err
		}

		from, err := ethType.Sender(ethType.NewEIP155Signer(common.ChainID), &t)
		if err != nil {
			return nil, nil, err
		}

		to := ""
		if t.To() != nil {
			to = t.To().String()
		}

		data := &Transaction{
			From:     from.String(),
			To:       to,
			Value:    t.Value().String(),
			Gas:      t.Gas(),
			GasPrice: t.GasPrice().String(),
			Data:     t.Data(),
			Hash:     t.Hash().String(),
		}

		transactions = append(transactions, data)
	}

	sigs := make(map[string]string, len(block.GetSignatures()))
	for _, item := range block.GetSignatures() {
		pk, _ := crypto.DecompressPubkey(item.Validator)
		sigs[crypto.PubkeyToAddress(*pk).String()] = item.Signature
	}

	return transactions, sigs, nil
}

func GetTransactionsFromBlk(blk *types.Block) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0, len(blk.Transactions()))
	for _, tran := range blk.Transactions() {
		var t ethType.Transaction
		if err := rlp.DecodeBytes(tran, &t); err != nil {
			return nil, err
		}

		from, err := ethType.Sender(ethType.NewEIP155Signer(common.ChainID), &t)
		if err != nil {
			return nil, err
		}

		to := ""
		if t.To() != nil {
			to = t.To().String()
		}

		data := &Transaction{
			From:     from.String(),
			To:       to,
			Value:    t.Value().String(),
			Gas:      t.Gas(),
			GasPrice: t.GasPrice().String(),
			Data:     t.Data(),
			Hash:     t.Hash().String(),
		}

		transactions = append(transactions, data)
	}
	return transactions, nil
}