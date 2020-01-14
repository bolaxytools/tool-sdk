package sdk

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"

	"github.com/bolaxy/accounts/keystore"
	"github.com/bolaxy/common"
	"github.com/bolaxy/common/hexutil"
	"github.com/bolaxy/crypto"
	"github.com/bolaxy/eth/types"
)

const (
	n = keystore.StandardScryptN
	p = keystore.StandardScryptP
)

var (
	signer = types.NewEIP155Signer(big.NewInt(1))
)

/* GenerateKey 生成私钥的方法。每次调用会生成一个新的密钥
 * @param passphrase 用户密码
 * @return {
 * 	 *Key 密钥结构，参考Key结构体说明
 *   error 生成失败时返回
 * }
 */
func GenerateKey(passphrase string) (*Key, error) {
	pk, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate key")
	}

	key := &keystore.Key{
		Alias:      "anonymity",
		Id:         uuid.NewRandom(),
		Address:    common.Address{},
		PrivateKey: pk,
	}

	blob, err := keystore.EncryptKey(key, passphrase, n, p)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt key")
	}

	return newKey(pk, passphrase, blob), nil
}

/* NewKey 初始化已知的私钥。
 * @param address 用户账户地址，从属于私钥
 * @param pubkey 公钥，从属于私钥
 * @param jsonkey keystore格式的JSON字符串。经过passphrase加密过。
 * @param passphrase 用户密码，用于恢复jsonkey成私钥
 * @return {
 * 	 *Key 密钥结构，参考Key结构体说明
 *   error 生成失败时返回
 * }
 *
 */
func NewKey(jsonkey []byte, passphrase string) (*Key, error) {
	if len(jsonkey) == 0 {
		return nil, errors.New("wrong params")
	}

	key := &Key{
		Address:    "",
		Pubkey:     "",
		JsonKey:    jsonkey,
		passphrase: passphrase,
	}

	k, err := keystore.DecryptKey(jsonkey, passphrase)
	if err != nil {
		return nil, err
	}

	key.address = k.Address
	key.pubKey = &k.PrivateKey.PublicKey
	return key, nil
}

// GetAddress 获取hex格式账户地址字符串
func (k *Key) GetStringAddress() string {
	if len(k.Address) == 0 {
		k.Address = k.address.String()
	}
	return k.Address
}

func (k *Key) GetAddress() common.Address {
	return k.address
}

// GetPubkey 获取压缩公钥的hex格式字符串
func (k *Key) GetPubkey() string {
	if len(k.Pubkey) == 0 && k.pubKey != nil {
		k.Pubkey = hexutil.Encode(crypto.CompressPubkey(k.pubKey))
	}
	return k.Pubkey
}

// SignTx 用密钥对交易数据签名。返回签名后的交易数据。
func (k *Key) SignTx(tx *types.Transaction) (*types.Transaction, error) {
	var (
		keyStore *keystore.Key
		err      error
	)

	if k.passphrase != "" && len(k.JsonKey) > 0 {
		keyStore, err = keystore.DecryptKey(k.JsonKey, k.passphrase)
		if err != nil {
			return nil, errors.Wrap(err, "SignTx")
		}
	}
	defer func() {
		keyStore = nil
	}()

	return types.SignTx(tx, signer, keyStore.PrivateKey)
}

// Key 密钥，密钥的账户、压缩🗜️公钥和Keystore字符串
type Key struct {
	Address string
	Pubkey  string
	JsonKey []byte

	address    common.Address
	pubKey     *ecdsa.PublicKey
	passphrase string
}

func newKey(pk *ecdsa.PrivateKey, passphrase string, jsonKey []byte) *Key {
	key := &Key{
		JsonKey:    jsonKey,
		pubKey:     &pk.PublicKey,
		address:    crypto.PubkeyToAddress(pk.PublicKey),
		passphrase: passphrase,
	}
	return key
}
