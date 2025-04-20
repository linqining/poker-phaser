package poker

import (
	"testing"
)

func TestNewDeckMasked(t *testing.T) {
	deck, err := NewDeckMasked()
	if err != nil {
		t.Fatal(err)
	}
	suiteSet := make(map[string]struct{})
	valueSet := make(map[string]struct{})
	for _, card := range deck.Cards {
		suiteSet[card.ClassicCard.Suite] = struct{}{}
		valueSet[card.ClassicCard.Value] = struct{}{}
		t.Log(card.ClassicCard)
		t.Log(card.ToCard())
	}
	for suiteName, _ := range suiteSet {
		t.Log(suiteName)
	}
	for valueName, _ := range valueSet {
		t.Log(valueName)
	}
	t.Log(deck.Cards)
}
