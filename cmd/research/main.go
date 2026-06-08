package main

import (
	"os"

	"github.com/ZenanH/research/internal/app"
)

var version = "dev"

func main() {
	os.Exit(app.Run(os.Args[1:], version))
}
