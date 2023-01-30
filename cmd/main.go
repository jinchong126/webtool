package main

import (
	"os"
	"webtool/cmd/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
