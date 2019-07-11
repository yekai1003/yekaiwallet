package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	Name       string
	NetworkUrl string
}

func (cli *CLI) Usage() {
	fmt.Println("./yekaiwallet createwallet -- for create a new wallet")
	fmt.Println("./yekaiwallet balance -address ADDRESS -- for get ether balance of a address")
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

	addressStr := balancecmd.String("address", "", "ADDRESS")

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
		fmt.Println("call create wallet")
	}

	if balancecmd.Parsed() {
		fmt.Println("call get balance of : ", *addressStr)
	}

}
