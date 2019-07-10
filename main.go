package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/davecgh/go-spew/spew"

	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tyler-smith/go-bip39"
)

func test_mnemonic() {

	b, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic("failed to NewEntropy:", err)
	}
	fmt.Println(b)

	mnemonic, err := bip39.NewMnemonic(b)
	if err != nil {
		log.Panic("failed to NewMnemonic:", err)
	}

	fmt.Println(mnemonic)

	seed := bip39.NewSeed(mnemonic, "123")

	fmt.Printf("%x\n", seed)
}

func test2() {
	mne := "august human human affair mechanic night verb metal embark marine orient million"

	wallet, err := NewFromMnemonic(mne)
	if err != nil {
		log.Panic(err)
	}

	path := MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(account.Address.Hex())

	path = MustParseDerivationPath("m/44'/60'/0'/0/1")
	account, err = wallet.Derive(path, false)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(account.Address.Hex())
}

func test_sign() {
	mnemonic := "august human human affair mechanic night verb metal embark marine orient million"
	wallet, err := NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	path := MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(0)
	value := big.NewInt(1000000000000000000) // 1个以太
	toAddress := common.HexToAddress("0x29155963f8632EaeD108f6A81eA65c75C62e77c0")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		log.Fatal(err)
	}
	//为了展示
	spew.Dump(signedTx)
}

func test_sendTransaction() {
	cli, err := ethclient.Dial("HTTP://127.0.0.1:7545") //注意地址变化 8545
	if err != nil {
		log.Panic(err)
	}

	defer cli.Close()

	mnemonic := "august human human affair mechanic night verb metal embark marine orient million"
	wallet, err := NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	path := MustParseDerivationPath("m/44'/60'/0'/0/1") //第2个账户地址
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(2)
	value := big.NewInt(3000000000000000000)
	toAddress := common.HexToAddress("0x44f4CD617655104649C1b866D20D5EAE198deD38")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		log.Fatal("failed to signed tx:", err)
	}

	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Panic("failed to send transaction:", err)
	}

	// bl, err := cli.BalanceAt(cli, toAddress, big.NewInt(0))
	// fmt.Println(bl.Int64(), err)
}

func main() {
	//test_mnemonic()
	// test2()
	// test_sign()
	test_sendTransaction()
}
