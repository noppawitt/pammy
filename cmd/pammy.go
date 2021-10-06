package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/noppawitt/pammy"
)

func main() {
	var token string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	bot, err := pammy.NewClient(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.Start(); err != nil {
		log.Fatal(err)
	}
}
