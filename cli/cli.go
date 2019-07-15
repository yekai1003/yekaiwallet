package client

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"yekaiwallet/hdkeystore"
	"yekaiwallet/hdwallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/howeyc/gopass"
)

type CLI struct {
	DataPath   string
	NetworkUrl string
}

func NewCLI(path, url string) *CLI {
	return &CLI{
		DataPath:   path,
		NetworkUrl: url,
	}
}

func (cli *CLI) Usage() {
	fmt.Println("./yekaiwallet createwallet -name ACCOUNT_NAME -- for create a new wallet")
	fmt.Println("./yekaiwallet balance -name ACCOUNT_NAME -- for get ether balance of a address")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.Usage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	cwcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	balancecmd := flag.NewFlagSet("balance", flag.ExitOnError)

	accountName := cwcmd.String("name", "yekai", "ACCOUNT_NAME")

	//addressStr := balancecmd.String("address", "", "ADDRESS")
	ba_account := balancecmd.String("name", "yekai", "ACCOUNT_NAME")

	switch os.Args[1] {
	case "createwallet":
		err := cwcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse createwallet params:", err)
		}
	case "balance":
		err := balancecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse balancecmd params:", err)
		}
	default:
		cli.Usage()
		os.Exit(1)
	}

	if cwcmd.Parsed() {

		if !cli.checkPath(*accountName) {
			fmt.Println("the keystore director is not null,you can not create wallet!")
			os.Exit(1)
		}
		fmt.Println("call create wallet, Please input your password for keystore")
		pass, err := gopass.GetPasswd()
		if err != nil {
			log.Panic("failed to get your password:", err)
		}

		cli.CreateWallet(*accountName, string(pass))
	}

	if balancecmd.Parsed() {
		fmt.Println("call get balance of accouname is ", *ba_account)
		cli.getBalances(*ba_account)
	}

}

func (cli *CLI) checkPath(name string) bool {
	infos, err := ioutil.ReadDir(cli.DataPath + "/" + name)
	if err != nil {
		fmt.Println("failed to ReadDir:", err)
		return true
	}
	if len(infos) > 0 {
		return false
	}
	return true

}

func (cli *CLI) getBalances(name string) {
	rclient, err := rpc.Dial(cli.NetworkUrl)
	if err != nil {
		log.Panic("failed to Dial:", err)
	}
	defer rclient.Close()
	infos, err := ioutil.ReadDir(cli.DataPath + "/" + name)
	if err != nil {
		fmt.Println("failed to ReadDir:", err)
		return
	}
	for _, info := range infos {
		//fmt.Printf("The balance of %s is %s\n", info.Name(), cli.getBalance(info.Name(), rclient))
		fmt.Printf("The balance of %s is %v\n", info.Name(), cli.getBalance("0x29155963f8632EaeD108f6A81eA65c75C62e77c0", rclient))
	}
}

func hex2bigInt(hex string) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(hex[2:], 16)
	return n
}

func (cli *CLI) getBalance(address string, client *rpc.Client) *big.Int {
	var resutl string
	err := client.Call(&resutl, "eth_getBalance", common.HexToAddress(address), "latest")
	if err != nil {
		log.Panic("failed to call eth_getBalance:", err)
	}
	return hex2bigInt(resutl)
}

func (cli *CLI) CreateWallet(name, pass string) {

	mnemonic, err := hdwallet.NewMnemonic(160)
	if err != nil {
		log.Panic("failed to NewMnemonic:", err)
	}

	fmt.Printf("Please remember the mnemonic:\n[%s]\n\n", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, "")
	if err != nil {
		log.Panic("failed to NewFromMnemonic:", err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Panic("failed to Derive:", err)
	}
	fmt.Println(account.Address.Hex())

	pkey, err := wallet.PrivateKey(account)
	if err != nil {
		log.Panic("failed to PrivateKey:", err)
	}

	hdks := hdkeystore.NewHDKeyStore(cli.DataPath+"/"+name, pkey)

	err = hdks.StoreKey(account.Address.Hex(), pass)
	if err != nil {
		log.Panic("failed to store key:", err)
	}
}
