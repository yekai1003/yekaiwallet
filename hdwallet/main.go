package hdwallet

import (
	"fmt"
	"log"
	"math/big"

	"github.com/davecgh/go-spew/spew"

	"context"

	"yekaiwallet/hdkeystore"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
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

	wallet, err := NewFromMnemonic(mne, "")
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
	wallet, err := NewFromMnemonic(mnemonic, "")
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

	// mnemonic := "august human human affair mechanic night verb metal embark marine orient million"
	// wallet, err := NewFromMnemonic(mnemonic, "")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// path := MustParseDerivationPath("m/44'/60'/0'/0/1") //第2个账户地址
	// account, err := wallet.Derive(path, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// nonce := uint64(3)
	// value := big.NewInt(3000000000000000000)
	// toAddress := common.HexToAddress("0x44f4CD617655104649C1b866D20D5EAE198deD38")
	// gasLimit := uint64(21000)
	// gasPrice := big.NewInt(21000000000)
	// var data []byte

	// tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	// signedTx, err := wallet.SignTx(account, tx, nil)
	// if err != nil {
	// 	log.Fatal("failed to signed tx:", err)
	// }

	// err = cli.SendTransaction(context.Background(), signedTx)
	// if err != nil {
	// 	log.Panic("failed to send transaction:", err)
	// }
	toAddress := common.HexToAddress("0x47E45e6E5336aE450A8aB1657CFCfb210b2661D6")
	bl, err := cli.BalanceAt(context.Background(), toAddress, big.NewInt(4))
	fmt.Println(bl.Int64(), bl.Uint64(), err)
}

func test_keystore() {

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

	wallet, err := NewFromMnemonic(mnemonic, "123")
	if err != nil {
		log.Fatal(err)
	}

	path := MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(account.Address.Hex())

	privateKey, err := wallet.derivePrivateKey(path)
	if err != nil {
		log.Panic("failed to derivePrivateKey:", err)
	}

	key := hdkeystore.NewKeyFromECDSA(privateKey)

	fmt.Println(key.Id.String())

	ks := hdkeystore.NewHDKeyStore("./data/"+account.Address.Hex(), nil)

	err = ks.StoreKey(ks.KeysDirPath, "456")
	if err != nil {
		log.Panic("failed to StoreKey:", err)
	}

}

func test_keystore2() {

	mnemonic := "august human human affair mechanic night verb metal embark marine orient million"

	wallet, err := NewFromMnemonic(mnemonic, "")
	if err != nil {
		log.Fatal(err)
	}

	path := MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(account.Address.Hex())

	privateKey, err := wallet.derivePrivateKey(path)
	if err != nil {
		log.Panic("failed to derivePrivateKey:", err)
	}

	key := hdkeystore.NewKeyFromECDSA(privateKey)

	fmt.Println(key.Id.String())

	ks := hdkeystore.NewHDKeyStore("./data/", nil)
	fmt.Println(ks)
	//ks.JoinPath(account.Address.Hex())
	err = ks.StoreKey(ks.JoinPath(account.Address.Hex()), "456")
	if err != nil {
		log.Panic("failed to StoreKey:", err)
	}

}

func test_login_keystore(addr string) {
	//0x44f4CD617655104649C1b866D20D5EAE198deD38

	ks := hdkeystore.NewHDKeyStore("./data/"+addr, nil)
	pkey, err := ks.GetKey(common.HexToAddress(addr), ks.KeysDirPath, "123")
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(pkey.Id.String())

	nonce := uint64(1)
	value := big.NewInt(3000000000000000000)
	toAddress := common.HexToAddress("0xf7066ce6e19351fb463e8c19f4aebbb3749e8167")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	account := accounts.Account{
		Address: common.HexToAddress(addr),
		URL: accounts.URL{
			Scheme: "",
			Path:   "m/44'/60'/0'/0/0",
		},
	}

	signedTx, err := ks.SignTx(account, tx, nil)
	if err != nil {
		log.Fatal("failed to signed tx:", err)
	}

	cli, err := ethclient.Dial("HTTP://127.0.0.1:8545") //注意地址变化 8545
	if err != nil {
		log.Panic(err)
	}

	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Panic("failed to send transaction:", err)
	}

}

func test_getBalance1(addr string) {
	cli, err := ethclient.Dial("HTTP://127.0.0.1:8545") //注意地址变化 8545
	if err != nil {
		log.Panic(err)
	}
	defer cli.Close()
	data, err := cli.BalanceAt(context.Background(), common.HexToAddress(addr), nil)

	fmt.Println(err, data)

}

func test_getBalance2(addr string) {
	cli, err := rpc.Dial("HTTP://127.0.0.1:8545")
	//cli, err := ethclient.Dial("HTTP://127.0.0.1:8545") //注意地址变化 8545
	if err != nil {
		log.Panic(err)
	}
	defer cli.Close()
	var balance string
	err = cli.Call(&balance, "eth_getBalance", common.HexToAddress(addr), "latest")

	fmt.Println(err, balance)

}

func main() {
	//test_mnemonic()
	//test2()
	//test_sign()
	//test_sendTransaction()
	//cli := &CLI{"yekai", "http://localhost:8545"}
	//test_keystore2()
	//test_login_keystore("0x5d065bce6e212d3acb544dc58d3bf9756eb76c6c")
	test_getBalance1("0x5d065bce6e212d3acb544dc58d3bf9756eb76c6c")
	test_getBalance2("0x5d065bce6e212d3acb544dc58d3bf9756eb76c6c")
	//cli.Run()
}
