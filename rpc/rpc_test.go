package rpc_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/bolaxytools/tool-sdk/rpc"
)

const (
	testAddress         = "0xbf0c265f0d1b3df1229f34486b62fee1e99f0d10"
	testContractAddress = "0x599d7abdb0a289f85aaca706b55d1b96cc07f348"
	testTxHash = "0xe349b239e5b2fbb8ebe96556c3caa4c2b419f9a51af5e497bba0735c88a48b6d"
)

var (
	client = rpc.Dial("http://192.168.10.189:8082")
)

func TestClient_FetchNonce(t *testing.T) {
	nonce, err := client.FetchNonce(testAddress)
	if err != nil {
		t.Logf("FetchNonce: %v", err)
		t.FailNow()
	}
	t.Logf("nonce: %d", nonce)
}

func TestClient_FetchChainInfo(t *testing.T) {
	meta, err := client.FetchChainInfo()
	if err != nil {
		t.Logf("FetchChainInfo: %v", err)
		t.FailNow()
	}

	spew.Dump(meta)
}

func TestClient_FetchAccount(t *testing.T) {
	acc, err := client.FetchAccount(testAddress)
	if err != nil {
		t.Logf("FetchAccount: %v", err)
		t.FailNow()
	}

	spew.Dump(acc)
}

func TestClient_FetchBalance(t *testing.T) {
	blk, err := client.FetchBlock(1)
	if err != nil {
		t.Logf("FetchBalance: %v", err)
		t.FailNow()
	}

	spew.Dump(blk)
}

func TestClient_IsContract(t *testing.T) {
	chk, err := client.IsContract(testContractAddress)
	if err != nil {
		t.Logf("IsContract: %v", err)
		t.FailNow()
	}

	if !chk {
		t.Logf("unexpect value")
		t.FailNow()
	}
}

func TestClient_FetchReceipt(t *testing.T) {
	receipt, err := client.FetchReceipt(testTxHash)
	if err != nil {
		t.Logf("FetchReceipt: %v", err)
		t.FailNow()
	}

	spew.Dump(receipt)
}