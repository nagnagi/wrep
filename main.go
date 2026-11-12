package main

import (
	"os"

	"wrep/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
