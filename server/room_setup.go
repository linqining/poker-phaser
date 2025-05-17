package poker

import (
	"log"
	"mental-poker/mental_poker"
)

func (room *Room) SetUpGame() error {
	initialDeckResp, err := mental_poker.InitializeDeck()
	if err != nil {
		return err
	}
	game := mental_poker.NewGame(initialDeckResp.Cards, initialDeckResp.SeedHex)
	room.game = game
	return nil
}

func (room *Room) setup() error {
	room.SetUpGame()
	players := []*mental_poker.Player{}

	room.Each(0, func(o *Occupant) bool {
		player := mental_poker.NewPlayer(room.game)
		player.Setup()
		players = append(players, player)
		o.SetPlayer(player)
		return true
	})
	aggPlayers := []*mental_poker.AggPlayer{}
	for _, player := range players {
		aggPlayers = append(aggPlayers, player.ToAggPlayer())
	}

	// each player compute aggkey
	for _, player := range players {
		aggResp, err := player.ComputeAggregatekey(aggPlayers)
		if err != nil {
			return err
		}
		player.SetJoinedKey(aggResp.JoinedKey)
	}

	maskResp, err := players[0].Mask()
	if err != nil {
		return err
	}
	log.Println(maskResp)
	cards := []string{}
	for _, card := range maskResp.Cards {
		cards = append(cards, card.MaskedCard)
	}
	originCards := cards
	finalCards := []string{}
	finalProof := ""
	for _, player := range players {
		shuffleResp, err := player.Shuffle(originCards)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, p := range players {
			_, verifyShuffleErr := p.VerifyShuffle(originCards, shuffleResp.Cards, shuffleResp.ShuffleProof)
			if verifyShuffleErr != nil {
				log.Println(verifyShuffleErr)
				return verifyShuffleErr
			}
		}
		originCards = shuffleResp.Cards
		finalCards = shuffleResp.Cards
		finalProof = shuffleResp.ShuffleProof
	}
	log.Println("shuffle complete", finalProof, finalCards)
	room.game.SetShuffleCards(finalCards)
	room.maskedDeck = NewDeckMasked(room.game.InitialCards, room.game.ShuffleCards)
	return nil
}

//func( room *Room) dealCard(player) error {
//	for i := 0; i < 4; i++ {
//		card := finalCards[i]
//		player := players[i]
//		tokens := []RevealTokenAndProof{}
//		for _, p := range players {
//			if player.GameUserID != p.GameUserID {
//				resp, err := p.ComputeRevealToken(card)
//				if err != nil {
//					t.Fatal(err)
//				}
//				val := resp.TokenMap[card]
//				tokens = append(tokens, val)
//			}
//		}
//		t.Log("reveal token", tokens)
//		player.ReceiveCard(card, tokens)
//	}
//
//	for _, player := range players {
//		peekResp, err := player.PeekCards(player.ReceiveCards[0])
//		if err != nil {
//			t.Fatal(err)
//		}
//		for _, card := range peekResp.CardMap {
//			userCard := intialCardMap[card]
//			t.Log(userCard)
//		}
//	}
//}
