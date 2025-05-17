package mental_poker

import (
	"github.com/google/uuid"
)

type Game struct {
	InitialCards []InitialCard `json:"initial_cards"`
	SeedHex      string        `json:"seed_hex"`
	ShuffleCards []string      `json:"shuffle_cards"`
	GameID       string        `json:"game_id"`
}

func NewGame(cards []InitialCard, seedHex string) *Game {
	return &Game{
		InitialCards: cards,
		SeedHex:      seedHex,
		GameID:       uuid.New().String(),
	}
}

func (g *Game) SetShuffleCards(shuffleCards []string) {
	g.ShuffleCards = shuffleCards
}

//func (g *Game) ComputeAggregatekey(players []*AggPlayer) (*ComputeAggKeyResp, error) {
//	c := new(http.Client)
//	req := request.NewRequest(c)
//	req.Json = map[string]interface{}{
//		"players":  players,
//		"seed_hex": g.SeedHex,
//	}
//	resp, err := req.Post(computeAggUrl)
//	if err != nil {
//		return nil, err
//	}
//	data, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//	aggResponse := new(ComputeAggKeyResp)
//	err = json.Unmarshal(data, aggResponse)
//	if err != nil {
//		return nil, err
//	}
//	return aggResponse, nil
//}
