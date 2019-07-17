package client

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"yekaiwallet/abi"
	"yekaiwallet/hdkeystore"
	"yekaiwallet/hdwallet"
	"yekaiwallet/util"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/howeyc/gopass"
)

type CLI struct {
	DataPath   string
	NetworkUrl string
	TokensFile string
}

type TokenConfig struct {
	Symbol string `json:"symbol"`
	Addr   string `json:"addr"`
}

func NewCLI(path, url string) *CLI {
	return &CLI{
		DataPath:   path,
		NetworkUrl: url,
		TokensFile: "tokens.json",
	}
}

func (cli *CLI) Usage() {
	fmt.Println("./yekaiwallet createwallet -name ACCOUNT_NAME -- for create a new wallet")
	fmt.Println("./yekaiwallet balance -name ACCOUNT_NAME -- for get ether balance of a address")
	fmt.Println("./yekaiwallet transfer -name ACCOUNT_NAME -toaddress ADDRESS -value VALUE -- for send ether to ADDRESS")
	fmt.Println("./yekaiwallet addtoken -addr CONTRACT_ADDR -- for send ether to ADDRESS")
	fmt.Println("./yekaiwallet tokenbalance -name ACCOUNT_NAME -- for get token balances")
	fmt.Println("./yekaiwallet sendtoken -name ACCOUNT_NAME -symbol SYMBOL -toaddress ADDRESS -value VALUE -- for send tokens to ADDRESS ")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.Usage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	balancecmd := flag.NewFlagSet("balance", flag.ExitOnError)

	createwalletcmd_acct := createwalletcmd.String("name", "yekai", "ACCOUNT_NAME")

	//addressStr := balancecmd.String("address", "", "ADDRESS")
	ba_account := balancecmd.String("name", "yekai", "ACCOUNT_NAME")

	//addtoken
	addtokencmd := flag.NewFlagSet("addtoken", flag.ExitOnError)
	addtokencmd_addr := addtokencmd.String("addr", "", "CONTRACT_ADDR")

	//tokenbalance
	tokenbalancecmd := flag.NewFlagSet("tokenbalance", flag.ExitOnError)
	tokenbalancecmd_acct := tokenbalancecmd.String("name", "yekai", "ACCOUNT_NAME")

	//sendtoken
	sendtokencmd := flag.NewFlagSet("sendtoken", flag.ExitOnError)
	sendtokencmd_acct := sendtokencmd.String("name", "yekai", "ACCOUNT_NAME")
	sendtokencmd_sym := sendtokencmd.String("symbol", "", "SYMBOL")
	sendtokencmd_toaddr := sendtokencmd.String("toaddress", "", "ADDRESS")
	sendtokencmd_value := sendtokencmd.Int64("value", 0, "VALUE")

	//transfer
	transfercmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	transfercmd_acct := transfercmd.String("name", "yekai", "ACCOUNT_NAME")
	transfercmd_toaddr := transfercmd.String("toaddress", "", "ADDRESS")
	transfercmd_value := transfercmd.Int64("value", 0, "VALUE")

	switch os.Args[1] {
	case "createwallet":
		err := createwalletcmd.Parse(os.Args[2:])
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
	case "tokenbalance":
		err := tokenbalancecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse tokenbalancecmd params:", err)
		}
	case "sendtoken":
		err := sendtokencmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse sendtokencmd params:", err)
		}
	case "transfer":
		err := transfercmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("failed to Parse transfercmd params:", err)
		}
	default:
		cli.Usage()
		os.Exit(1)
	}

	if createwalletcmd.Parsed() {

		if !cli.checkPath(*createwalletcmd_acct) {
			fmt.Println("the keystore director is not null,you can not create wallet!")
			os.Exit(1)
		}
		fmt.Println("call create wallet, Please input your password for keystore")
		pass, err := gopass.GetPasswd()

		if err != nil {
			log.Panic("failed to get your password:", err)
		}

		cli.CreateWallet(*createwalletcmd_acct, string(pass))
	}

	if balancecmd.Parsed() {
		fmt.Println("call get balance of accouname is ", *ba_account)
		cli.getBalances(*ba_account)
	}

	if addtokencmd.Parsed() {

		if *addtokencmd_addr == "" {
			fmt.Println("token's address is err")
			cli.Usage()
			os.Exit(1)
		}
		cli.addToken(*addtokencmd_addr)
	}

	if tokenbalancecmd.Parsed() {

		cli.GetTokensBalance(*tokenbalancecmd_acct)
	}

	if sendtokencmd.Parsed() {
		if *sendtokencmd_sym == "" || *sendtokencmd_toaddr == "" || *sendtokencmd_value == 0 {
			fmt.Println("params err for sendtoken!")
			cli.Usage()
			os.Exit(1)
		}
		//调用token转账函数
		cli.SendToken(*sendtokencmd_sym, *sendtokencmd_toaddr, *sendtokencmd_acct, *sendtokencmd_value)
	}

	//转账etc
	if transfercmd.Parsed() {
		if *transfercmd_toaddr == "" || *transfercmd_value == 0 {
			fmt.Println("params err for transfer!")
			cli.Usage()
			os.Exit(1)
		}
		//调用token转账函数
		cli.Transfer(*transfercmd_acct, *transfercmd_toaddr, *transfercmd_value)
	}

}

func (cli *CLI) Transfer(acct_name, toaddr string, value int64) {
	//获得转出账户地址
	fromaddr, err := cli.getAcctAddr(acct_name)
	if err != nil {
		log.Panic("failed to getAcctAddr ", err)
	}

	client, err := ethclient.Dial(cli.NetworkUrl)
	if err != nil {
		log.Panic("failed to Transfer when Dial ", err)
	}
	defer client.Close()

	//获取当前nonce值
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(fromaddr), nil)
	if err != nil {
		log.Panic("failed to Transfer when NonceAt ", err)
	}
	fmt.Println(nonce)

	gasLimit := uint64(300000)
	gasPrice := big.NewInt(21000000000)
	tx := types.NewTransaction(nonce, common.HexToAddress(toaddr), big.NewInt(value), gasLimit, gasPrice, []byte("salary"))

	//去做签名
	hdks := hdkeystore.NewHDKeyStore(cli.DataPath+"/"+acct_name, nil)
	fmt.Println("Please input your pass:")
	pass, err := gopass.GetPasswd()
	if err != nil {
		log.Panic("failed to get pass ", err)
	}
	_, err = hdks.GetKey(common.HexToAddress(fromaddr), hdks.JoinPath(fromaddr), string(pass))
	if err != nil {
		log.Panic("failed to Transfer when GetKey ", err)
	}

	stx, err := hdks.SignTx(common.HexToAddress(fromaddr), tx, nil)
	if err != nil {
		log.Panic("failed to Transfer when SignTx ", err)
	}

	err = client.SendTransaction(context.Background(), stx)
	if err != nil {
		log.Panic("failed to Transfer when SendTransaction ", err)
	}
}

func (cli *CLI) SendToken(sym, toaddr, acct_name string, value int64) {
	//获得token的合约地址
	contract, err := cli.getContract(sym)
	if err != nil {
		log.Panic("failed to getContract ", err)
	}
	//获得发送方账户地址，以及交易签名
	fromaddr, err := cli.getAcctAddr(acct_name)
	if err != nil {
		log.Panic("failed to getAcctAddr ", err)
	}

	txopt, err := makeAuth(cli.DataPath + "/" + acct_name + "/" + fromaddr)
	if err != nil {
		log.Panic("failed to makeAuth ", err)
	}
	cli.sendTokenTx(cli.NetworkUrl, contract, toaddr, value, txopt)
}

func (cli *CLI) sendTokenTx(rawurl, contract, toaddr string, value int64, opt *bind.TransactOpts) {
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		log.Panic("failed to sendTokenTx ", err)
	}

	defer client.Close()

	erc20, err := erc20abi.NewErc20(common.HexToAddress(contract), client)
	if err != nil {
		log.Panic("failed to sendTokenTx when NewErc20 ", err)
	}

	txhash, err := erc20.Transfer(opt, common.HexToAddress(toaddr), big.NewInt(value))
	if err != nil {
		log.Panic("failed to Transfer ", err)
	}
	fmt.Println("sendtoken call ok,hash=", txhash.Hash().Hex())
}

func makeAuth(ksfile string) (*bind.TransactOpts, error) {
	keyin, err := os.Open(ksfile)
	if err != nil {
		log.Panic("failed to read keystore file ", err)
	}
	//ioutil.ReadAll(os.Stdin)
	fmt.Println("please input pass for tx:")
	pass, err := gopass.GetPasswd()
	fmt.Println(string(pass), "----------")
	if err != nil {
		log.Panic("failed to read GetPasswd file ", err)
	}
	return bind.NewTransactor(keyin, string(pass))
}

func (cli *CLI) GetTokensBalance(acct_name string) {
	//先读取要获取哪些token
	data, err := ioutil.ReadFile(cli.TokensFile)
	if err != nil {
		log.Panic("failed to GetTokensBalance when ReadFile ", err)
	}
	tokens := []TokenConfig{}
	err = json.Unmarshal(data, &tokens)
	if err != nil {
		log.Panic("failed to GetTokensBalance when Unmarshal ", err)
	}
	//再读取账户下有哪些地址
	acctaddr, err := cli.getAcctAddr(acct_name)
	if err != nil {
		log.Panic("failed to GetTokensBalance when getAcctAddr ", err)
	}
	//每种token，每个地址做一个处理
	for _, token := range tokens {
		amount, err := cli.getTokenBalance(token.Addr, acctaddr)
		if err != nil {
			fmt.Println("failed to getTokenBalance ", err)
			continue
		}
		fmt.Printf("%s:%v\n", token.Symbol, amount)
	}
}

func (cli *CLI) getTokenBalance(contract, account string) (*big.Int, error) {
	client, err := ethclient.Dial(cli.NetworkUrl)
	if err != nil {
		log.Panic("failed to getTokenBalance when dial", err)
	}
	erc20, err := erc20abi.NewErc20(common.HexToAddress(contract), client)
	if err != nil {
		log.Panic("failed to getTokenBalance when NewErc20", err)
	}
	return erc20.BalanceOf(nil, common.HexToAddress(account))
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

func (cli *CLI) getAcctAddr(account_name string) (string, error) {
	infos, err := ioutil.ReadDir(cli.DataPath + "/" + account_name)
	if err != nil {
		fmt.Println("failed to ReadDir:", err)
		return "", err
	}
	if len(infos) <= 0 {
		return "", errors.New("The wallet does'n has account")
	}
	return infos[0].Name(), nil

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
		fmt.Printf("The balance of %s is %v\n", info.Name(), cli.getBalance(info.Name(), rclient))
	}
}

func (cli *CLI) getContract(symbol string) (string, error) {
	data, err := ioutil.ReadFile(cli.TokensFile)
	if err != nil {
		log.Panic("failed to read tokensfile ", err)
	}
	tokens := []TokenConfig{}
	err = json.Unmarshal(data, &tokens)
	if err != nil {
		log.Panic("failed to Unmarshal data ", err)
	}

	for _, token := range tokens {
		if symbol == token.Symbol {
			return token.Addr, nil
		}
	}

	return "", errors.New("failed to get token,please add token first!")
}

func (cli *CLI) GetSymbol(rawurl, address string) (string, error) {
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		log.Panic("failed to GetSymbol when Dial:", err)
	}
	defer client.Close()
	erc20, err := erc20abi.NewErc20(common.HexToAddress(address), client)

	if err != nil {
		log.Panic("failed to GetSymbol when NewErc20", err)
	}

	return erc20.Symbol(nil)
}

func checkUnique(address string, tokens []TokenConfig) (bool, string) {
	for _, token := range tokens {
		if address == token.Addr {
			return true, token.Symbol
		}
	}
	return false, ""
}

func (cli *CLI) addToken(address string) {
	//形成一个数组，然后存储到文件中:json
	//tokens := make([]TokenConfig, 0)
	tokens := []TokenConfig{}

	//读历史token信息
	data, err := ioutil.ReadFile(cli.TokensFile)
	if err != nil {
		fmt.Println("failed to read tokensfile ", err)
	} else {
		err = json.Unmarshal(data, &tokens)
		if err != nil {
			log.Panic("failed to Unmarshal data ", err)
		}
	}

	fmt.Println(tokens)

	if ok, sym := checkUnique(address, tokens); ok {
		fmt.Println("token already exists, name is", sym)
		return
	}

	symbol, err := cli.GetSymbol(cli.NetworkUrl, address)

	if err != nil {
		log.Panic("failed to addToken when GetSymbol", err)
	}

	token := TokenConfig{symbol, address}
	tokens = append(tokens, token)
	fmt.Println(tokens)
	data, _ = json.Marshal(tokens)
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
