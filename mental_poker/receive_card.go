package mental_poker

type ReceiveCard struct {
	Card        string   `json:"card"`
	RevealToken []string `json:"reveal_token"`
}
