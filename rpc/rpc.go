package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/bolaxy/common"
	"github.com/bolaxy/common/hexutil"
	"github.com/bolaxy/core/types"
	ethTypes "github.com/bolaxy/eth/types"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func Transact(data string) (*RawTxRes, error) {
	svcUrl := fmt.Sprintf("%s%s", host, transferUrl)
	payload, err := post(svcUrl, "text/plain", strings.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "transfer[post]")
	}

	var res RawTxRes
	if err = decodeResult(payload, &res); err != nil {
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
	if err = decodeResult(payload, &acc); err != nil {
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
	if err = decodeResult(payload, &blk); err != nil {
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
	if err = decodeResult(payload, &receipt); err != nil {
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
	if err = decodeResult(payload, &meta); err != nil {
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
	if err = decodeResult(payload, &jsonCallRes); err != nil {
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

var typeBig = big.NewInt(0)

func float64ToBigInt() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Float64 || t != reflect.TypeOf(typeBig) {
			return data, nil
		}

		var z big.Int
		z.SetString(decimal.NewFromFloat(data.(float64)).String(), 10)
		return &z, nil
	}
}

func float64ToUint64() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Float64 || t.Kind() != reflect.Uint64 {
			return data, nil
		}

		return uint64(data.(float64)), nil
	}
}

func hexToHash() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(common.Hash{}) {
			return data, nil
		}

		if data.(string) == "null" {
			return common.Hash{}, nil
		}

		return common.HexToHash(data.(string)), nil
	}
}

func hexToAddress() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String ||
			(t != reflect.TypeOf(common.Address{}) &&
				t != reflect.TypeOf(new(common.Address))) {
			return data, nil
		}

		if data.(string) == "null" {
			return nil, nil
		}

		addr := common.HexToAddress(data.(string))
		if t == reflect.TypeOf(common.Address{}) {
			return addr, nil
		} else {
			return &addr, nil
		}

	}
}

func base64ToSlice() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(make([]byte, 0)) {
			return data, nil
		}

		if data == nil {
			return nil, nil
		}

		ret, err := base64.StdEncoding.DecodeString(data.(string))
		if err != nil {
			// try to convert hex to slice
			return hexutil.Decode(data.(string))
		}

		return ret, nil
	}
}

func base64ToArray() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Slice || t != reflect.TypeOf(make([][]byte, 0)) {
			return data, nil
		}

		if data == nil {
			return nil, nil
		}

		transactions := data.([]interface{})
		ret := make([][]byte, len(transactions))
		for i, ts := range transactions {
			v, err := base64.StdEncoding.DecodeString(ts.(string))
			if err != nil {
				return nil, err
			}
			ret[i] = v
		}
		return ret, nil
	}
}

func hexToBloom() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		var bloom ethTypes.Bloom
		if f.Kind() != reflect.String || t != reflect.TypeOf(bloom) {
			return data, nil
		}

		ret, err := hexutil.Decode(data.(string))
		if err != nil {
			return nil, err
		}

		copy(bloom[:], ret)
		return bloom, nil
	}
}

func hexToUint64OrUint() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || (t.Kind() != reflect.Uint64 && t.Kind() != reflect.Int) {
			return data, nil
		}

		value, err := hexutil.DecodeUint64(data.(string))
		if err != nil {
			if t.Kind() == reflect.Uint64 {
				return uint64(0), err
			} else {
				return uint(0), err
			}
		}

		if t.Kind() == reflect.Uint64 {
			return value, nil
		} else {
			return uint(value), nil
		}
	}
}

func decodeResult(result []byte, value interface{}) error {
	var ret map[string]interface{}
	if err := json.Unmarshal(result, &ret); err != nil {
		return errors.Wrap(err, "json unmarshal")
	}

	if ret["Err"] != "" {
		return errors.New(fmt.Sprintf("response err -> %s", ret["Err"].(string)))
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			float64ToBigInt(),
			float64ToUint64(),
			hexToHash(),
			hexToAddress(),
			base64ToSlice(),
			base64ToArray(),
			hexToBloom(),
			hexToUint64OrUint(),
		),
		Result: value,
	})
	if err != nil {
		return errors.Wrap(err, "mapstructure new")
	}

	// fmt.Printf("[Data] %s\n", ret["Data"])
	if err = decoder.Decode(ret["Data"]); err != nil {
		return errors.Wrap(err, "mapstructure decode")
	}

	return nil
}