package main

import (
	"fmt"

	"github.com/gen2brain/dlgs"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func test() {
	// TEST
	passwd, _, err := dlgs.Password("Password", "Enter your API key:")
	if err != nil {
		panic(err)
	}
	fmt.Println(passwd)
}
