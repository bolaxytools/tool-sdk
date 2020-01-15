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
	"sync"

	"github.com/bolaxy/common"
	"github.com/bolaxy/common/hexutil"
	"github.com/bolaxy/core/types"
	ethTypes "github.com/bolaxy/eth/types"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Client bolaxy client
type Client struct {
	host        string
	decoderPool sync.Pool
}

// Dial http client api
func Dial(host string) *Client {
	if host == "" {
		host = defaultHost
	}
	return &Client{host: host, decoderPool: sync.Pool{New: func() interface{} {
		return NewDecoder(
			WithHook(float64ToBigInt),
			WithHook(float64ToUint64),
			WithHook(hexToHash),
			WithHook(hexToAddress),
			WithHook(base64ToSlice),
			WithHook(base64ToArray),
			WithHook(hexToBloom),
			WithHook(hexToUint64OrUint),
		)
	}}}
}

// Decoder json response decoder, can be reuse in next time
type Decoder interface {
	// Reset reset decoder for reuse next time
	Reset()
	// Decode decode json response and mapping into the struct
	Decode(result []byte, output interface{}) error
}

// DecoderOpt options of Decoder
type DecoderOpt func(decoder *responseDecoder)

// WithHook add hooks for the decoder
func WithHook(fn mapstructure.DecodeHookFunc) DecoderOpt {
	return func(decoder *responseDecoder) {
		if fn == nil {
			return
		}

		if decoder.hooks == nil {
			decoder.hooks = make([]mapstructure.DecodeHookFunc, 0, 10)
		}

		decoder.hooks = append(decoder.hooks, fn)
	}
}

// NewDecoder create new decoder
func NewDecoder(opts ...DecoderOpt) Decoder {
	decoder := &responseDecoder{}
	for _, opt := range opts {
		opt(decoder)
	}

	return decoder
}

// Transact bolaxy tansfer api
// the parameter data is hex encoded string that been rlp serialized data.
func (c *Client) Transact(data string) (*RawTxRes, error) {
	svcUrl := c.host + transferUrl
	payload, err := post(svcUrl, "text/plain", strings.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "transfer[post]")
	}

	var res RawTxRes
	if err = c.decode(payload, &res); err != nil {
		return nil, errors.Wrap(err, "transfer[unmarshal]")
	}

	return &res, nil
}

// IsContract bolaxy check  is contract api
// the parameter address is hex formated string, e.g. 0x599d7abdb0a289f85aaca706b55d1b96cc07f348
func (c *Client) IsContract(address string) (bool, error) {
	acc, err := c.FetchAccount(address)
	if err != nil {
		return false, errors.Wrap(err, "isContract")
	}

	if len(acc.Code) > 0 {
		return true, nil
	}

	return false, nil
}

// FetchNonce bolaxy fetch nonce api
// the parameter address is hex formated string, e.g. 0x599d7abdb0a289f85aaca706b55d1b96cc07f348
func (c *Client) FetchNonce(address string) (uint64, error) {
	acc, err := c.FetchAccount(address)
	if err != nil {
		return 0, errors.Wrap(err, "fetchNonce")
	}
	return acc.Nonce, nil
}

// FetchBalance bolaxy fetch balance value api
// the parameter address is hex formated string, e.g. 0x599d7abdb0a289f85aaca706b55d1b96cc07f348
func (c *Client) FetchBalance(address string) (*big.Int, error) {
	acc, err := c.FetchAccount(address)
	if err != nil {
		return nil, errors.Wrap(err, "fetchBalance")
	}

	return acc.Balance, nil
}

// FetchAccount bolaxy fetch account info api
// the parameter address is hex formated string, e.g. 0x599d7abdb0a289f85aaca706b55d1b96cc07f348
func (c *Client) FetchAccount(address string) (*JsonAccount, error) {
	payload, err := get(c.host, accountUrl, address)
	if err != nil {
		return nil, errors.Wrap(err, "fetchAccount[get]")
	}

	var acc JsonAccount
	if err = c.decode(payload, &acc); err != nil {
		return nil, errors.Wrap(err, "fetchAccount[unmarshal]")
	}

	return &acc, nil
}

// FetchBlock bolaxy fetch block data api
// the parameter index is the blockheight, begin at 0
// if no block exist, will return error
func (c *Client) FetchBlock(index int) (*types.Block, error) {
	payload, err := get(c.host, blkSvcUrl, strconv.Itoa(index))
	if err != nil {
		return nil, errors.Wrap(err, "fetchBlock[get]")
	}

	var blk types.Block
	if err = c.decode(payload, &blk); err != nil {
		return nil, errors.Wrap(err, "fetchBlock[unmarshal]")
	}

	return &blk, nil
}

// FetchReceipt bolaxy fetch receipt api
// txhash is hex string that is transaction hash
func (c *Client) FetchReceipt(txhash string) (*JsonReceipt, error) {
	payload, err := get(c.host, receiptUrl, txhash)
	if err != nil {
		return nil, errors.Wrap(err, "fetchReceipt[get]")
	}

	var receipt JsonReceipt
	if err = c.decode(payload, &receipt); err != nil {
		return nil, errors.Wrap(err, "fetchReceipt[unmarshal]")
	}

	return &receipt, nil
}

// FetchChainInfo bolaxy fetch chain info api
// this api will get current block height and block chain status
func (c *Client) FetchChainInfo() (*ChainMeta, error) {
	payload, err := get(c.host, infoUrl)
	if err != nil {
		return nil, errors.Wrap(err, "fetchChainInfo[get]")
	}

	var meta ChainMeta
	if err = c.decode(payload, &meta); err != nil {
		return nil, errors.Wrap(err, "fetchChainInfo[unmarshal]")
	}

	return &meta, nil
}

// CallContract bolaxy call contract api
// the read only contract method can call this method
// {
// 	"from": "0x8F55dAa29339bB9685019D57ba70A638FE0040d9",
// 	"to": "0x3D5F11e6627422BFf4E5d8C8475d0c59C8521352",
// 	"gas": 100000,
// 	"gasPrice": 100,
// 	"value": 1000000,
// 	"data": "",
// 	"nonce": 1
// }
func (c *Client) CallContract(msg *SendTxArgs) ([]byte, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Wrap(err, "callContract[marshal sendtxargs]")
	}

	callURL := c.host + callSvcUrl
	payload, err = post(callURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.Wrap(err, "callContract[post]")
	}

	var jsonCallRes JsonCallRes
	if err = c.decode(payload, &jsonCallRes); err != nil {
		return nil, errors.Wrap(err, "callContract[unmarshal call json]")
	}

	result, err := hexutil.Decode(jsonCallRes.Data)
	if err != nil {
		return nil, errors.Wrap(err, "callContract[decode hex]")
	}

	return result, nil
}

func (c *Client) decode(data []byte, output interface{}) error {
	decoder := c.decoderPool.Get().(Decoder)
	defer c.decoderPool.Put(decoder)
	decoder.Reset()
	return decoder.Decode(data, output)
}

func get(getUrl ...string) ([]byte, error) {
	// fmt.Printf("get url: %s\n", strings.Join(getUrl, ""))
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

func float64ToBigInt(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Float64 || t != reflect.TypeOf(typeBig) {
		return data, nil
	}

	var z big.Int
	z.SetString(decimal.NewFromFloat(data.(float64)).String(), 10)
	return &z, nil
}

func float64ToUint64(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Float64 || t.Kind() != reflect.Uint64 {
		return data, nil
	}

	return uint64(data.(float64)), nil
}

func hexToHash(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(common.Hash{}) {
		return data, nil
	}

	if data.(string) == "null" {
		return common.Hash{}, nil
	}

	return common.HexToHash(data.(string)), nil
}

func hexToAddress(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

func base64ToSlice(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

func base64ToArray(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

func hexToBloom(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

func hexToUint64OrUint(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

type responseDecoder struct {
	hooks  []mapstructure.DecodeHookFunc
	result map[string]interface{}
}

func (r *responseDecoder) Reset() {
	r.result = make(map[string]interface{})
}

func (r *responseDecoder) Decode(result []byte, output interface{}) error {
	if r.result == nil {
		r.result = make(map[string]interface{})
	}

	// fmt.Printf("decode --> %s\n", result)
	if err := json.Unmarshal(result, &r.result); err != nil {
		return errors.Wrap(err, "json unmarshal")
	}

	if r.result["Err"] != "" {
		return errors.New(fmt.Sprintf("response err -> %s", r.result["Err"].(string)))
	}

	config := &mapstructure.DecoderConfig{Result: output}
	if len(r.hooks) == 1 {
		config.DecodeHook = r.hooks[0]
	} else if len(r.hooks) > 1 {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(r.hooks...)
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(r.result["Data"])
}