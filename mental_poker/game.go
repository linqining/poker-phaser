package mental_poker

type Game struct {
	InitialCards []InitialCard `json:"initial_cards"`
	SeedHex      string        `json:"seed_hex"`
	ShuffleCards []string      `json:"shuffle_cards"`
}
