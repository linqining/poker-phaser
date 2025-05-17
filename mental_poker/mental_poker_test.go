package mental_poker

import (
	"github.com/ecodeclub/ekit/slice"
	"testing"
)

func TestGenerate(t *testing.T) {
	initialDeck, err := InitializeDeck()
	if err != nil {
		t.Fatal(err)
	}
	intialCardMap := slice.ToMapV(initialDeck.Cards, func(element InitialCard) (string, ClassicCard) {
		return element.Card, element.ClassicCard
	})

	game := NewGame(initialDeck.Cards, initialDeck.SeedHex)
	//t.Log(initialDeck)

	andrija := NewPlayer(game)
	andrija.Setup()
	kobi := NewPlayer(game)
	kobi.Setup()
	nico := NewPlayer(game)
	nico.Setup()
	tom := NewPlayer(game)
	tom.Setup()

	players := []*Player{
		andrija, kobi, nico, tom,
	}
	aggPlayers := []*AggPlayer{
		andrija.ToAggPlayer(), kobi.ToAggPlayer(), nico.ToAggPlayer(), tom.ToAggPlayer(),
	}

	// each player compute aggkey
	for _, player := range players {
		aggResp, err := player.ComputeAggregatekey(aggPlayers)
		if err != nil {
			t.Fatal(err)
		}
		player.SetJoinedKey(aggResp.JoinedKey)
	}

	maskResp, err := players[0].Mask()
	t.Log(maskResp)
	t.Log(err)
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
			t.Fatal(err)
		}
		for _, p := range players {
			_, verifyShuffleErr := p.VerifyShuffle(originCards, shuffleResp.Cards, shuffleResp.ShuffleProof)
			if verifyShuffleErr != nil {
				t.Fatal(verifyShuffleErr)
			}
		}
		originCards = shuffleResp.Cards
		finalCards = shuffleResp.Cards
		finalProof = shuffleResp.ShuffleProof
	}
	t.Log("shuffle complete", finalProof, finalCards)
	game.SetShuffleCards(finalCards)
	for i := 0; i < 4; i++ {
		card := finalCards[i]
		player := players[i]
		tokens := []RevealTokenAndProof{}
		for _, p := range players {
			if player.GameUserID != p.GameUserID {
				resp, err := p.ComputeRevealToken([]string{card})
				if err != nil {
					t.Fatal(err)
				}
				val := resp.TokenMap[card]
				tokens = append(tokens, val)
			}
		}
		t.Log("reveal token", tokens)
		player.ReceiveCard(card, tokens)
	}

	for _, player := range players {
		peekResp, err := player.PeekCards(player.ReceiveCards)
		if err != nil {
			t.Fatal(err)
		}
		for _, card := range peekResp.CardMap {
			userCard := intialCardMap[card]
			t.Log(userCard)
		}
	}

	//shuffledCards := shuffleResp.Cards

	//// each player maskcard
	//for _, player := range players {
	//	player.Mask()
	//}

}

//
//func TestInitialize(t *testing.T) {
//	data, err := InitializeDeck()
//	t.Log(err)
//	t.Log(data)
//	gameID := "game123"
//	andrija := NewPlayer()
//	andrija.Setup()
//}
