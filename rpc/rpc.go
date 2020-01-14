package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bolaxy/common/hexutil"
	"github.com/bolaxy/core/types"
)

func Transact(data string) (*RawTxRes, error) {
	svcUrl := fmt.Sprintf("%s%s", host, transferUrl)
	payload, err := post(svcUrl, "text/plain", strings.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "transfer[post]")
	}

	var res RawTxRes
	if err = json.Unmarshal(payload, &res); err != nil {
		return nil, errors.Wrap(err, "transfer[unmarshal]")
	}

	return &res, nil
}

func IsContract(address string) (bool, error) {
	acc, err := FetchAccount(address)
	if err != nil {
		return false, errors.Wrap(err, "isContract")
	}

	if len(acc.Code) > 0 {
		return true, nil
	}

	return false, nil
}

func FetchNonce(address string) (uint64, error) {
	acc, err := FetchAccount(address)
	if err != nil {
		return 0, errors.Wrap(err, "fetchNonce")
	}
	return acc.Nonce, nil
}

func FetchBalance(address string) (*big.Int, error) {
	acc, err := FetchAccount(address)
	if err != nil {
		return nil, errors.Wrap(err, "fetchBalance")
	}

	return acc.Balance, nil
}

func FetchAccount(address string) (*JsonAccount, error) {
	payload, err := get(host, accountUrl, address)
	if err != nil {
		return nil, errors.Wrap(err, "fetchAccount[get]")
	}

	var acc JsonAccount
	if err = json.Unmarshal(payload, &acc); err != nil {
		return nil, errors.Wrap(err, "fetchAccount[unmarshal]")
	}

	return &acc, nil
}

func FetchBlock(index int) (*types.Block, error) {
	payload, err := get(host, blkSvcUrl, strconv.Itoa(index))
	if err != nil {
		return nil, errors.Wrap(err, "fetchBlock[get]")
	}

	var blk types.Block
	if err = json.Unmarshal(payload, &blk); err != nil {
		return nil, errors.Wrap(err, "fetchBlock[unmarshal]")
	}

	return &blk, nil
}

func FetchReceipt(txhash string) (*JsonReceipt, error) {
	payload, err := get(host, receiptUrl, txhash)
	if err != nil {
		return nil, errors.Wrap(err, "fetchReceipt[get]")
	}

	var receipt JsonReceipt
	if err = json.Unmarshal(payload, &receipt); err != nil {
		return nil, errors.Wrap(err, "fetchReceipt[unmarshal]")
	}

	return &receipt, nil
}

func FetchChainInfo() (*ChainMeta, error) {
	payload, err := get(host, infoUrl)
	if err != nil {
		return nil, errors.Wrap(err, "fetchChainInfo[get]")
	}

	var meta ChainMeta
	if err = json.Unmarshal(payload, &meta); err != nil {
		return nil, errors.Wrap(err, "fetchChainInfo[unmarshal]")
	}

	return &meta, nil
}

/*
	{
		"from": "0x8F55dAa29339bB9685019D57ba70A638FE0040d9",
		"to": "0x3D5F11e6627422BFf4E5d8C8475d0c59C8521352",
		"gas": 100000,
		"gasPrice": 100,
		"value": 1000000,
		"data": "",
		"nonce": 1
	}
*/
func CallContract(msg *SendTxArgs) ([]byte, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Wrap(err, "callContract[marshal sendtxargs]")
	}

	callURL := fmt.Sprintf("%s%s", host, callSvcUrl)
	payload, err = post(callURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.Wrap(err, "callContract[post]")
	}

	var jsonCallRes JsonCallRes
	if err = json.Unmarshal(payload, &jsonCallRes); err != nil {
		return nil, errors.Wrap(err, "callContract[unmarshal call json]")
	}

	result, err := hexutil.Decode(jsonCallRes.Data)
	if err != nil {
		return nil, errors.Wrap(err, "callContract[decode hex]")
	}

	return result, nil
}

func get(getUrl ...string) ([]byte, error) {
	resp, err := http.Get(strings.Join(getUrl, ""))
	if err != nil {
		return nil, err
	}

	return readResp(resp.Body)
}

func post(postUrl, contentType string, body io.Reader) ([]byte, error) {
	resp, err := http.Post(postUrl, contentType, body)
	if err != nil {
		return nil, err
	}

	return readResp(resp.Body)
}

func readResp(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
