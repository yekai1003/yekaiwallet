package main

import (
	"yekaiwallet/cli"
)

func main() {
	cli := client.NewCLI("./data", "http://localhost:8545")

	cli.Run()
}
