package client

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"yekaiwallet/hdkeystore"
	"yekaiwallet/hdwallet"
	"yekaiwallet/util"

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
	fmt.Println("./yekaiwallet transfer -name ACCOUNT_NAME -toaddress ADDRESS -value VALUE -- for send ether to ADDRESS")
	fmt.Println("./yekaiwallet addtoken -addr CONTRACT_ADDR -- for send ether to ADDRESS")
	fmt.Println("./yekaiwallet tokenbalance -name ACCOUNT_NAME -- for get token balances")
	fmt.Println("./yekaiwallet sendtoken -name ACCOUNT_NAME -tokenname SYMBOL -toaddress ADDRESS -value VALUE -- for send tokens to ADDRESS ")
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

	//addtoken
	addtokencmd := flag.NewFlagSet("addtoken", flag.ExitOnError)
	//addtokencmd_addr := addtokencmd.String("addr", "", "CONTRACT_ADDR")

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
	case "addtoken":
		err := addtokencmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse addtokencmd params:", err)
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

	if addtokencmd.Parsed() {
		cli.addToken()
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

type TokenConfig struct {
	Name string
	Addr string
}

func (cli *CLI) addToken() {
	//形成一个数组，然后存储到文件中:json
	//tokens := make([]TokenConfig, 0)
	tokens := []TokenConfig{}
	token := TokenConfig{"ykc", "0xb96a1f071727692e09acd337c8273a39398c0c70"}
	tokens = append(tokens, token)
	fmt.Println(tokens)
	data, _ := json.Marshal(tokens)
	utils.WriteKeyFile("tokens.json", data)

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
