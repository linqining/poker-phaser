package poker

type Agent struct {
	Addr string
}

type PokerAgent interface {
}

func NewAgent(addr string) *Agent {
	return &Agent{Addr: addr}
}

func (a *Agent) SendInitialDeck(cards []MaskedCard) error {

}
