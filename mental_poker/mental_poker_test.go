package mental_poker

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	initialDeck, err := InitializeDeck()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(initialDeck)

	gameID := "game123"
	andrija := NewPlayer()
	andrija.Setup(gameID, "andrija", initialDeck)
	kobi := NewPlayer()
	kobi.Setup(gameID, "kobi", initialDeck)
	nico := NewPlayer()
	nico.Setup(gameID, "nico", initialDeck)
	tom := NewPlayer()
	tom.Setup(gameID, "tom", initialDeck)

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
	//// each player maskcard
	//for _, player := range players {
	//	player.Mask()
	//}

}

func TestInitialize(t *testing.T) {
	data, err := InitializeDeck()
	t.Log(err)
	t.Log(data)
	gameID := "game123"
	andrija := NewPlayer()
	andrija.Setup(gameID, "andrija", data)
}
