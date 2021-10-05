package main

import (
	"flag"
	"log"

	"github.com/noppawitt/pammy"
)

func main() {
	var token string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	bot, err := pammy.NewClient(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.Start(); err != nil {
		log.Fatal(err)
	}
}
