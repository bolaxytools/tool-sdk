package sdk

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"github.com/pkg/errors"

	"github.com/bolaxy/accounts/keystore"
	"github.com/bolaxy/common"
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

func GenerateKey() (*Key, error) {
	pk, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generate key")
	}

	return newKey(pk), nil
}

func RecoverKey(raw []byte) (*Key, error) {
	buff := bytes.NewBuffer(raw)
	pk, err := ecdsa.GenerateKey(crypto.S256(), buff)
	if err != nil {
		return nil, err
	}
	return newKey(pk), nil
}

func (k *Key) ExportPrivateKey() []byte {
	return k.PK.D.Bytes()
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

// SignTx 用密钥对交易数据签名。返回签名后的交易数据。
func (k *Key) SignTx(tx *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(tx, signer, k.PK)
}

type Key struct {
	Address string
	PK      *ecdsa.PrivateKey

	address common.Address
}

func newKey(pk *ecdsa.PrivateKey) *Key {
	key := &Key{
		PK:      pk,
		address: crypto.PubkeyToAddress(pk.PublicKey),
	}
	return key
}
