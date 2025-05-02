package mental_poker

type ReceiveCard struct {
	Card        string                `json:"card"`
	RevealToken []RevealTokenAndProof `json:"reveal_tokens"`
}
