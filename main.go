package main

import (
	"flag"
	"fmt"

	"github.com/jonsth131/ctfd-cli/tui"
)

func main() {
	baseUrl := flag.String("baseurl", "", "Base URL for API requests")
	logging := flag.Bool("log", false, "Log to file")
	flag.Parse()

	if *baseUrl == "" {
		fmt.Println("Please provide a base URL with -baseurl")
		return
	}

	tui.StartTea(*baseUrl, *logging)
}
