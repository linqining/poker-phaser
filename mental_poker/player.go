package mental_poker

import (
	"encoding/json"
	"github.com/mozillazg/request"
	"io"
	"log"
	"net/http"
)

type Player struct {
}

const (
	baseUrl  = "http://127.0.0.1:8000"
	setUpUrl = baseUrl + "/deck/setup"
)

func NewPlayer() *Player {
	return &Player{}
}

type SetUpResponse struct {
	GameID        string `json:"game_id"`
	GameUserID    string `json:"game_user_id"`
	UserID        string `json:"user_id"`
	UserKeyProof  []byte `json:"user_key_proof"`
	UserPublicKey []byte `json:"user_public_key"`
}

func (p *Player) Setup() (*SetUpResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]string{
		"user_id":      "123",
		"game_id":      "game_id123",
		"game_user_id": "game_user_id123",
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
	return setUpResponse, nil
}
