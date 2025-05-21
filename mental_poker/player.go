package mental_poker

import (
	"bytes"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/mozillazg/request"
	"io"
	"log"
	"net/http"
)

// 实例化一个新的color对象，设置前景色为红色，背景色为绿色，文字斜体
var colorPrint *color.Color

func init() {
	colorPrint = color.New()
	colorPrint.Add(color.FgRed) // 设置前景色为红色
	//colorPrint.Add(color.Italic)  // 设置文字为斜体
	colorPrint.Add(color.BgCyan) // 设置背景色为绿色
}

type UserKeyProof struct {
	Commit  string `json:"commit"`
	Opening string `json:"opening"`
}

type Player struct {
	GameUserID   string
	Game         *Game
	PublicKey    string
	UserKeyProof UserKeyProof
	JoinedKey    string
	ReceiveCards []ReceiveCard
}

func (p *Player) ToAggPlayer() *AggPlayer {
	return &AggPlayer{
		GameID:        p.Game.GameID,
		GameUserID:    p.GameUserID,
		UserKeyProof:  p.UserKeyProof,
		UserPublicKey: p.PublicKey,
	}
}

const (
	baseUrl          = "http://127.0.0.1:8000"
	setUpUrl         = baseUrl + "/deck/setup"
	clearUrl         = baseUrl + "/deck/clear"
	initialUrl       = baseUrl + "/deck/initialize"
	computeAggUrl    = baseUrl + "/deck/compute_aggregate_key"
	maskUrl          = baseUrl + "/deck/mask"
	shuffleUrl       = baseUrl + "/deck/shuffle"
	verifyShuffleUrl = baseUrl + "/deck/verify_shuffle"
	revelTokenUrl    = baseUrl + "/deck/reveal_token"
	peekCardsUrl     = baseUrl + "/deck/peek_cards"
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
	colorPrint.Println("SetUp initial public parameters:")
	log.Println(string(data))
	ret := &InitializeDeckResp{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewPlayer(game *Game) *Player {
	return &Player{
		GameUserID: uuid.New().String(),
		Game:       game,
	}
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

func (p *Player) Setup() (*SetUpResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]string{
		"user_id":      "123",
		"game_id":      p.Game.GameID,
		"game_user_id": p.GameUserID,
		"seed_hex":     p.Game.SeedHex,
	}
	resp, err := req.Post(setUpUrl)
	if err != nil {
		return nil, err
	}

	data, _ := io.ReadAll(resp.Body)

	colorPrint.Println("Player %s setUp publicKey and provide publicKey proof:", p.GameUserID)
	log.Println(string(data))
	setUpResponse := new(SetUpResponse)
	err = json.Unmarshal(data, setUpResponse)
	if err != nil {
		return nil, err
	}
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
		"seed_hex": p.Game.SeedHex,
	}
	resp, err := req.Post(computeAggUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	colorPrint.Println("HomomorphicEncryption setup initial public parameters:")
	log.Println(string(data))
	aggResponse := new(ComputeAggKeyResp)
	err = json.Unmarshal(data, aggResponse)
	if err != nil {
		return nil, err
	}
	return aggResponse, nil
}

type MaskResponse struct {
	Cards []struct {
		MaskedCard string `json:"masked_card"`
		Proof      struct {
			A string `json:"a"`
			B string `json:"b"`
			R string `json:"r"`
		} `json:"proof"`
	} `json:"cards"`
}

func (p *Player) Mask() (*MaskResponse, error) {
	cards := slice.Map(p.Game.InitialCards, func(idx int, src InitialCard) string {
		return src.Card
	})
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"seed_hex":   p.Game.SeedHex,
		"cards":      cards,
		"joined_key": p.JoinedKey,
	}
	resp, err := req.Post(maskUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	colorPrint.Println("Mask the card before each player shuffle:")
	log.Println(string(data))
	maskResp := new(MaskResponse)
	err = json.Unmarshal(data, maskResp)
	if err != nil {
		return nil, err
	}
	return maskResp, nil
}

type ShuffleResponse struct {
	Cards        []string `json:"cards"`
	ShuffleProof string   `json:"shuffle_proof"`
}

func (p *Player) Shuffle(cards []string) (*ShuffleResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"seed_hex":   p.Game.SeedHex,
		"cards":      cards,
		"joined_key": p.JoinedKey,
	}
	resp, err := req.Post(shuffleUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	colorPrint.Println("Player %s shuffle the deck and provide proof:", p.GameUserID)
	log.Println(string(data))
	shuffleResp := new(ShuffleResponse)
	err = json.Unmarshal(data, shuffleResp)
	if err != nil {
		return nil, err
	}
	return shuffleResp, nil
}

type VerifyShuffleResponse struct {
}

func (p *Player) VerifyShuffle(originCards []string, shuffledCards []string, shuffleProof string) (*VerifyShuffleResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"proof":          shuffleProof,
		"joined_key":     p.JoinedKey,
		"seed_hex":       p.Game.SeedHex,
		"origin_cards":   originCards,
		"shuffled_cards": shuffledCards,
	}
	resp, err := req.Post(verifyShuffleUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	colorPrint.Println("Player %s verify the shuffle proof:", p.GameUserID)
	log.Println(string(data))
	shuffleResp := new(VerifyShuffleResponse)
	err = json.Unmarshal(data, shuffleResp)
	if err != nil {
		return nil, err
	}
	return shuffleResp, nil
}

type PedersenProof struct {
	A string `json:"a"`
	B string `json:"b"`
	R string `json:"r"`
}

type RevealTokenAndProof struct {
	Token         string        `json:"token"`
	PedersenProof PedersenProof `json:"proof"`
	PublicKey     string        `json:"public_key"`
}

type RevealTokenResponse struct {
	TokenMap map[string]RevealTokenAndProof `json:"token_map"`
}

func (p *Player) ComputeRevealToken(cards []string) (*RevealTokenResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"game_user_id": p.GameUserID,
		"seed_hex":     p.Game.SeedHex,
		"reveal_cards": cards,
	}
	resp, err := req.Post(revelTokenUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	colorPrint.Println("Player %s compute reveal token for other player:", p.GameUserID)
	log.Println(string(data))
	shuffleResp := new(RevealTokenResponse)
	err = json.Unmarshal(data, shuffleResp)
	if err != nil {
		return nil, err
	}
	return shuffleResp, nil
}

type PeekCardsResponse struct {
	CardMap map[string]string `json:"card_map"`
}

func (p *Player) PeekCards(receiveCards []ReceiveCard) (*PeekCardsResponse, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]interface{}{
		"game_user_id": p.GameUserID,
		"seed_hex":     p.Game.SeedHex,
		"peek_cards":   receiveCards,
	}
	resp, err := req.Post(peekCardsUrl)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	colorPrint.Println("Player %s peek his card:", p.GameUserID)
	log.Println(string(data))
	shuffleResp := new(PeekCardsResponse)
	err = json.Unmarshal(data, shuffleResp)
	if err != nil {
		return nil, err
	}
	return shuffleResp, nil
}

func (p *Player) ReceiveCard(card string, tokens []RevealTokenAndProof) error {
	p.ReceiveCards = append(p.ReceiveCards, ReceiveCard{
		Card:        card,
		RevealToken: tokens,
	})
	return nil
}

func (p *Player) Clear() error {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = map[string]string{
		"user_id":      "123",
		"game_id":      p.Game.GameID,
		"game_user_id": p.GameUserID,
	}
	resp, err := req.Post(clearUrl)
	if err != nil {
		return err
	}
	data, _ := io.ReadAll(resp.Body)

	colorPrint.Println("Game over clear player %s data:")
	log.Println(string(data))

	return nil
}
