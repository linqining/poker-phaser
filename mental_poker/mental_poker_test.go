package mental_poker

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	player := NewPlayer()
	setUpResp, err := player.Setup()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(setUpResp.GameUserID)
	//andrija := NewPlayer()
	//kobi := NewPlayer()
	//nico := NewPlayer()
	//tom := NewPlayer()

}
