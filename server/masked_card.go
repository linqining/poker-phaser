package poker

import (
	"encoding/json"
	"io"
	"net/http"
)

type ClassicCard struct {
	Value string `json:"value"`
	Suite string `json:"suite"`
}

type MaskedCard struct {
	ClassicCard ClassicCard `json:"classic_card"`
	Card        []byte      `json:"card"`
}

func (m MaskedCard) ToCard() Card {
	rank := 0
	suit := 0

	switch m.ClassicCard.Suite {
	case "Heart":
		suit = Heart
	case "Club":
		suit = Club
	case "Spade":
		suit = Spade
	case "Diamond":
		suit = Diamond
	default:
		panic(m.ClassicCard.Suite)
	}

	switch m.ClassicCard.Value {
	case "Two":
		rank = Deuce
	case "Three":
		rank = Trey
	case "Four":
		rank = Four
	case "Five":
		rank = Five
	case "Six":
		rank = Six
	case "Seven":
		rank = Seven
	case "Eight":
		rank = Eight
	case "Nine":
		rank = Nine
	case "Ten":
		rank = Ten
	case "Jack":
		rank = Jack
	case "Queen":
		rank = Queen
	case "King":
		rank = King
	case "Ace":
		rank = Ace
	default:
		panic(m.ClassicCard.Value)
	}
	if suit == 0 || rank < 0 {
		return NilCard
	}

	card := (1 << uint32(16+rank)) | suit | (rank << 8) | primes[rank]
	return Card(card)
}

type DeckMasked struct {
	Cards []MaskedCard `json:"cards"`
	pos   int
}

type InitialDeckResponse struct {
	Cards []MaskedCard `json:"cards"`
}

func NewDeckMasked() (*DeckMasked, error) {
	// todo replace address
	response, err := http.Get("http://127.0.0.1:8000/deck/initialize")
	if err != nil {
		return nil, err
	}
	initialDeck := new(InitialDeckResponse)
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, initialDeck)
	if err != nil {
		return nil, err
	}
	return &DeckMasked{Cards: initialDeck.Cards}, nil
}

func (deck *DeckMasked) Find(rank, suit int) Card {
	for _, maskedCard := range deck.Cards {
		card := maskedCard.ToCard()
		if card.Rank() == rank && card.Suit() == suit {
			return card
		}
	}

	return NilCard
}

//func (deck *Deck) Shuffle() {
//	deck.pos = 0
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
//	a := r.Perm(52)
//	for i, v := range a {
//		deck.cards[i], deck.cards[v] = deck.cards[v], deck.cards[i]
//	}
//}

func (deck *DeckMasked) Take() Card {
	if deck.pos >= NumCard {
		return NilCard
	}

	card := deck.Cards[deck.pos]
	deck.pos++
	return card.ToCard()
}
