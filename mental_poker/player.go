package mental_poker

import (
	"bytes"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/mozillazg/request"
	"io"
	"log"
	"net/http"
)

type UserKeyProof struct {
	Commit  string `json:"commit"`
	Opening string `json:"opening"`
}

type Player struct {
	PublicKey    string
	GameUserID   string
	GameID       string
	UserKeyProof UserKeyProof
	JoinedKey    string
	SeedHex      string
	Cards        []InitialCard
}

func (p *Player) ToAggPlayer() *AggPlayer {
	return &AggPlayer{
		GameID:        p.GameID,
		GameUserID:    p.GameUserID,
		UserKeyProof:  p.UserKeyProof,
		UserPublicKey: p.PublicKey,
	}
}

const (
	baseUrl       = "http://127.0.0.1:8000"
	setUpUrl      = baseUrl + "/deck/setup"
	initialUrl    = baseUrl + "/deck/initialize"
	computeAggUrl = baseUrl + "/deck/compute_aggregate_key"
	maskUrl       = baseUrl + "/deck/mask"
)

type ClassicCard struct {
	Suite string `json:"suite"`
	Value string `json:"value"`
}

type InitialCard struct {
	Card        string      `json:"card"`
	ClassicCard ClassicCard `json:"classic_card"`
}

type InitializeDeckResp struct {
	Cards   []InitialCard `json:"cards"`
	SeedHex string        `json:"seed_hex"`
}

func InitializeDeck() (*InitializeDeckResp, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	resp, err := req.Get(initialUrl)
	if err != nil {
		return nil, err
	}
	bytes.NewReader([]byte{})
	data, err := io.ReadAll(resp.Body)
	ret := &InitializeDeckResp{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewPlayer() *Player {
	return &Player{}
}

func (c *Player) SetJoinedKey(publicKey string) {
	c.JoinedKey = publicKey
}

type SetUpResponse struct {
	UserID        string `json:"user_id"`
	GameID        string `json:"game_id"`
	GameUserID    string `json:"game_user_id"`
	UserPublicKey string `json:"user_public_key"`
	UserKeyProof  struct {
		Commit  string `json:"commit"`
		Opening string `json:"opening"`
	} `json:"user_key_proof"`
}

func (p *Player) Setup(gameID string, gameUserID string, initialDeck *InitializeDeckResp) (*SetUpResponse, error) {
	p.SeedHex = initialDeck.SeedHex
	p.Cards = initialDeck.Cards

	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]string{
		"user_id":      "123",
		"game_id":      gameID,
		"game_user_id": gameUserID,
		"seed_hex":     initialDeck.SeedHex,
	}
	resp, err := req.Post(setUpUrl)
	if err != nil {
		return nil, err
	}
	//j, err := resp.Json()
	//if err != nil {
	//	return nil, err
	//}

	data, _ := io.ReadAll(resp.Body)
	log.Println(string(data))
	setUpResponse := new(SetUpResponse)
	err = json.Unmarshal(data, setUpResponse)
	if err != nil {
		return nil, err
	}
	p.GameID = gameID
	p.GameUserID = gameUserID
	p.PublicKey = setUpResponse.UserPublicKey
	p.UserKeyProof = UserKeyProof{
		Commit:  setUpResponse.UserKeyProof.Commit,
		Opening: setUpResponse.UserKeyProof.Opening,
	}
	return setUpResponse, nil
}

type AggPlayer struct {
	GameID        string       `json:"game_id"`
	GameUserID    string       `json:"game_user_id"`
	UserKeyProof  UserKeyProof `json:"user_key_proof"`
	UserPublicKey string       `json:"public_key"`
}

type ComputeAggKeyResp struct {
	JoinedKey string `json:"joined_key"`
}

func (p *Player) ComputeAggregatekey(players []*AggPlayer) (*ComputeAggKeyResp, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"players":  players,
		"seed_hex": p.SeedHex,
	}
	resp, err := req.Post(computeAggUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	aggResponse := new(ComputeAggKeyResp)
	err = json.Unmarshal(data, aggResponse)
	if err != nil {
		return nil, err
	}
	return aggResponse, nil
}

func (p *Player) Mask() (*ComputeAggKeyResp, error) {
	cards := slice.Map(p.Cards, func(idx int, src InitialCard) string {
		return src.Card
	})
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"seed_hex": p.SeedHex,
		"cards":    cards,
	}
	resp, err := req.Post(computeAggUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	aggResponse := new(ComputeAggKeyResp)
	err = json.Unmarshal(data, aggResponse)
	if err != nil {
		return nil, err
	}
	return aggResponse, nil
}
