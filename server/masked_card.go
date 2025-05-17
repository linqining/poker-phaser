package poker

import (
	"mental-poker/mental_poker"
)

type ClassicCard struct {
	Value string `json:"value"`
	Suite string `json:"suite"`
}

type MaskedCard struct {
	ClassicCard ClassicCard `json:"classic_card"`
	Card        string      `json:"card"`
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
	CardMap     map[string]MaskedCard
	pos         int
	MaskedCards []string
}

type InitialDeckResponse struct {
	Cards []MaskedCard `json:"cards"`
}

func NewDeckMasked(cards []mental_poker.InitialCard, shuffledCards []string) *DeckMasked {
	cardMap := make(map[string]MaskedCard)
	for _, card := range cards {
		classicCard := ClassicCard{Value: card.ClassicCard.Value, Suite: card.ClassicCard.Suite}
		cardMap[card.Card] = MaskedCard{ClassicCard: classicCard, Card: card.Card}
	}
	return &DeckMasked{CardMap: cardMap, MaskedCards: shuffledCards}
}

//func (deck *DeckMasked) Find(rank, suit int) Card {
//	for _, maskedCard := range deck.Cards {
//		card := maskedCard.ToCard()
//		if card.Rank() == rank && card.Suit() == suit {
//			return card
//		}
//	}
//
//	return NilCard
//}

//func (deck *Deck) Shuffle() {
//	deck.pos = 0
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
//	a := r.Perm(52)
//	for i, v := range a {
//		deck.cards[i], deck.cards[v] = deck.cards[v], deck.cards[i]
//	}
//}

func (deck *DeckMasked) Take() string {
	if deck.pos >= NumCard {
		return ""
	}
	card := deck.MaskedCards[deck.pos]
	deck.pos++
	return card
}
