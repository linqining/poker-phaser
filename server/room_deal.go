package poker

import (
	"mental-poker/mental_poker"
)

func (r *Room) DealCard(occupant *Occupant, num int) ([]Card, error) {
	receiveCards, cards, err := r.AskOccupantOpenCard(occupant, num)
	if err != nil {
		return nil, err
	}
	occupant.RevealCards = append(occupant.RevealCards, receiveCards...)
	occupant.Cards = append(occupant.Cards, cards...)
	return cards, nil
}

func (r *Room) AskOccupantOpenCard(occupant *Occupant, num int) ([]*mental_poker.ReceiveCard, []Card, error) {
	maskCards := []string{}
	for i := 0; i < num; i++ {
		card := r.maskedDeck.Take()
		maskCards = append(maskCards, card)
	}

	receiveCards, err := r.CollectRevealTokens(occupant, maskCards)
	if err != nil {
		return nil, nil, err
	}
	occupant.RevealCards = append(occupant.RevealCards, receiveCards...)

	revealCards := []mental_poker.ReceiveCard{}
	for _, card := range receiveCards {
		revealCards = append(revealCards, *card)
	}
	resp, err := occupant.player.PeekCards(revealCards)
	if err != nil {
		return nil, nil, err
	}
	cards := []Card{}
	for _, ucard := range receiveCards {
		initCard := resp.CardMap[ucard.Card]
		maskCard := r.maskedDeck.CardMap[initCard]
		cards = append(cards, maskCard.ToCard())
	}
	return receiveCards, cards, nil
}

func (r *Room) DealPublicCard(num int) ([]Card, error) {
	// randomly choose a occupant to revealpublic cards
	var occupant *Occupant
	for _, o := range r.Occupants {
		if o == nil {
			continue
		}
		occupant = o
		break
	}
	_, cards, err := r.AskOccupantOpenCard(occupant, num)
	if err != nil {
		return nil, err
	}
	return cards, nil
}
