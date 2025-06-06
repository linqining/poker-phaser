package main

import (
	"flag"
	poker "mental-poker/server"
)

func init() {
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) < 5 {
		return
	}

	if len(args) == 5 {
		var cards [5]poker.Card
		for i, v := range args {
			card := poker.ParseCard(v)
			if card == poker.NilCard {
				return
			}
			cards[i] = card
		}
		hand := poker.Eva5Hand(cards)
		_ = hand
		//fmt.Println(hand)
	}

	if len(args) == 7 {
		var cards [7]poker.Card
		for i, v := range args {
			card := poker.ParseCard(v)
			if card == poker.NilCard {
				return
			}
			cards[i] = card
		}
		//fmt.Println(poker.Eva7Hand(cards))
	}
}
