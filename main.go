package main

import (
	"fmt"
	"yekaiwallet/cli"
)

func main() {
	fmt.Println("hello world")
	cli := client.NewCLI("./data", "http://localhost:7545")

	cli.Run()
}
