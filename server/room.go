package poker

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/ecodeclub/ekit/mapx"
	"log"
	"mental-poker/mental_poker"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	actionWait = 20 * time.Second
	MaxN       = 10
	mnemonic   = "fly happy jungle remind dune replace deer travel good man sleep faint"
	// address 0x813f37c19631325b3bb81c4922e8244c53b42afd46c755ee3e8da779903bbfd9
	MovePkgID     = "0x63aeecd28b9bc9cf5476bac7a6ad3e62e5427c15f0a210335df271e75c89baa9"
	GameDataObjID = "0x0a84bc3803250c79790245cf2c8ea8f74f9aa48efe5f40a30a205d9233cc5126"
)

type Room struct {
	Id        string      `json:"id"`
	SB        int         `json:"sb"`
	BB        int         `json:"bb"`
	Cards     []Card      `json:"cards,omitempty"`
	Pot       []int       `json:"pot,omitempty"`
	Timeout   int         `json:"timeout,omitempty"`
	Button    int         `json:"button,omitempty"`
	Occupants []*Occupant `json:"occupants,omitempty"`
	Chips     []int       `json:"chips,omitempty"`
	Bet       int         `json:"bet,omitempty"`
	N         int         `json:"n"`
	Max       int         `json:"max"`
	MaxChips  int         `json:"maxchips"`
	MinChips  int         `json:"minchips"`
	remain    int
	allin     int
	EndChan   chan int `json:"-"`
	exitChan  chan interface{}
	startChan chan struct{}
	lock      sync.Mutex
	//deck       *Deck
	maskedDeck *DeckMasked
	game       *mental_poker.Game
}

func NewRoom(id string, max int, sb, bb int) *Room {
	if max <= 0 || max > MaxN {
		max = 9 // default 9 occupants
	}

	room := &Room{
		Id:        id,
		Occupants: make([]*Occupant, max, MaxN),
		Chips:     make([]int, max, MaxN),
		SB:        sb,
		BB:        bb,
		Pot:       make([]int, 1),
		Timeout:   10,
		Max:       max,
		lock:      sync.Mutex{},
		//deck:      NewDeck(),
		EndChan:   make(chan int),
		exitChan:  make(chan interface{}, 1),
		startChan: make(chan struct{}, 1),
	}
	go func() {
		timer := time.NewTimer(time.Second * 6)
		for {
			select {
			case <-timer.C:
				room.start()
				timer.Reset(time.Second * 6)
			case <-room.exitChan:
				return
			case <-room.startChan:
				room.start()
			}
		}
	}()

	return room
}

func (room *Room) Cap() int {
	return len(room.Occupants)
}

func (room *Room) Occupant(id string) *Occupant {
	for _, o := range room.Occupants {
		if o != nil && o.Id == id {
			return o
		}
	}

	return nil
}

func (room *Room) AddOccupant(o *Occupant) int {
	room.lock.Lock()
	defer room.lock.Unlock()

	// room not exists
	if len(room.Id) == 0 {
		return 0
	}

	for pos, _ := range room.Occupants {
		if room.Occupants[pos] == nil {
			room.Occupants[pos] = o
			room.N++
			log.Println(room.N)
			o.Room = room
			o.Pos = pos + 1
			break
		}
	}

	return o.Pos
}

func (room *Room) DelOccupant(o *Occupant) {
	if o == nil || o.Pos == 0 {
		return
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	room.Occupants[o.Pos-1] = nil
	room.N--
	if len(o.Cards) > 0 {
		room.remain--
	}
	/*
		if o.Action == ActAllin {
			room.allin--
		}
	*/

	if room.N == 0 {
		DelRoom(room)
		select {
		case room.exitChan <- 0:
		default:
		}
	}

	if room.remain <= 1 {
		select {
		case room.EndChan <- 0:
		default:
		}
	}
}

func (room *Room) Broadcast(message *Message) {
	for _, o := range room.Occupants {
		if o != nil {
			o.SendMessage(message)
		}
	}
}

// start starts from 0
func (room *Room) Each(start int, f func(o *Occupant) (isContinue bool)) {
	end := (room.Cap() + start - 1) % room.Cap()
	i := start
	for ; i != end; i = (i + 1) % room.Cap() {
		if room.Occupants[i] != nil && !f(room.Occupants[i]) {
			return
		}
	}

	// end
	if room.Occupants[i] != nil {
		f(room.Occupants[i])
	}
}

func (room *Room) start() {
	var dealer *Occupant
	// remove zero chips user
	// 合约交互
	room.Each(0, func(o *Occupant) bool {
		if o.Chips < room.BB {
			o.Leave()
		}
		return true
	})
	room.lock.Lock()
	if room.N < 2 {
		room.lock.Unlock()
		return
	}

	room.setup()
	// Select Dealer
	button := room.Button - 1
	room.Each((button+1)%room.Cap(), func(o *Occupant) bool {
		room.Button = o.Pos
		dealer = o
		return false
	})

	if dealer == nil {
		return
	}

	//room.deck.Shuffle()

	// Small Blind
	sb := dealer.Next()
	if room.N == 2 { // one-to-one
		sb = dealer
	}
	// Big Blind
	bb := sb.Next()
	bbPos := bb.Pos

	room.Pot = nil
	room.Chips = make([]int, room.Max)
	room.Bet = 0
	room.Cards = nil
	room.remain = 0
	room.allin = 0
	room.Each(0, func(o *Occupant) bool {
		o.Bet = 0
		cards, err := room.DealCard(o, 2)
		if err != nil {
			log.Println(err)
		}
		o.Cards = cards
		o.Hand = 0
		//o.Action = ActReady
		o.Action = ""
		room.remain++

		return true
	})
	room.lock.Unlock()

	room.Broadcast(&Message{
		From:   room.Id,
		Type:   MsgPresence,
		Action: ActButton,
		Class:  strconv.Itoa(room.Button),
	})

	room.betting(sb.Pos, room.SB)
	room.betting(bb.Pos, room.BB)

	// Round 1 : preflop
	room.Each(sb.Pos-1, func(o *Occupant) bool {
		o.SendMessage(&Message{
			From:   room.Id,
			Type:   MsgPresence,
			Action: ActPreflop,
			Class:  o.Cards[0].String() + "," + o.Cards[1].String(),
		})
		return true
	})
	var (
		err        error
		turnCards  []Card
		riverCards []Card
	)

	room.action(bbPos%room.Cap() + 1)
	if room.remain <= 1 {
		goto showdown
	}
	room.calc()

	// Round 2 : Flop
	room.ready()
	room.Cards, err = room.DealPublicCard(3)
	if err != nil {
		panic(err)
	}
	room.Each(0, func(o *Occupant) bool {
		var hand [5]Card
		if len(o.Cards) > 0 {
			cards := hand[0:0]
			//cards = append(cards, o.RevealCards...) // todo handle reveal cards
			cards = append(cards, o.Cards...) //TODO this action handle by user not by server
			log.Println(o.player.GameUserID, cards)
			cards = append(cards, room.Cards...)
			log.Println(o.player.GameUserID, hand)
			log.Println(o.player.GameUserID, cards)
			o.Hand = Eva5Hand(hand)
		}
		o.SendMessage(&Message{
			From:   room.Id,
			Type:   MsgPresence,
			Action: ActFlop,
			Class:  fmt.Sprintf("%s,%s,%s,%d", room.Cards[0], room.Cards[1], room.Cards[2], o.Hand>>16),
		})

		return true
	})

	room.action(0)

	if room.remain <= 1 {
		goto showdown
	}
	room.calc()

	// Round 3 : Turn
	room.ready()
	turnCards, err = room.DealPublicCard(1)
	if err != nil {
		panic(err)
	}
	room.Cards = append(room.Cards, turnCards...)
	room.Each(0, func(o *Occupant) bool {
		var hand [6]Card
		if len(o.Cards) > 0 {
			cards := hand[0:0]
			cards = append(cards, o.Cards...)
			cards = append(cards, room.Cards...)
			o.Hand = Eva6Hand(hand)
		}
		o.SendMessage(&Message{
			From:   room.Id,
			Type:   MsgPresence,
			Action: ActTurn,
			Class:  fmt.Sprintf("%s,%d", room.Cards[3], o.Hand>>16),
		})

		return true
	})
	room.action(0)
	if room.remain <= 1 {
		goto showdown
	}
	room.calc()

	// Round 4 : River
	room.ready()
	riverCards, err = room.DealPublicCard(1)
	if err != nil {
		panic(err)
	}
	room.Cards = append(room.Cards, riverCards...)
	room.Each(0, func(o *Occupant) bool {
		var hand [7]Card
		if len(o.Cards) > 0 {
			cards := hand[0:0]
			cards = append(cards, o.Cards...)
			cards = append(cards, room.Cards...)
			o.Hand = Eva7Hand(hand)
		}
		o.SendMessage(&Message{
			From:   room.Id,
			Type:   MsgPresence,
			Action: ActRiver,
			Class:  fmt.Sprintf("%s,%d", room.Cards[4], o.Hand>>16),
		})

		return true
	})
	room.action(0)

showdown:
	room.showdown()
	// Final : Showdown
	room.Broadcast(&Message{
		From:   room.Id,
		Type:   MsgPresence,
		Action: ActShowdown,
		Room:   room,
	})
	log.Println("showdown end ", room.Id)
	room.checkAndEndGame()
}

func (room *Room) AllPlayers() []*mental_poker.Player {
	players := []*mental_poker.Player{}
	for _, occu := range room.Occupants {
		if occu == nil {
			continue
		}
		players = append(players, occu.player)
	}
	return players
}

func (room *Room) CollectRevealTokens(o *Occupant, cards []string) ([]*mental_poker.ReceiveCard, error) {
	players := room.AllPlayers()
	dealCardMap := make(map[string]*mental_poker.ReceiveCard)
	for _, card := range cards {
		dealCardMap[card] = &mental_poker.ReceiveCard{
			Card: card,
		}
	}
	for _, player := range players {
		if player.GameUserID == o.player.GameUserID {
			continue
		}
		tokenResp, err := player.ComputeRevealToken(cards)
		if err != nil {
			return nil, err
		}
		// todo encrypt token with users pk
		for card, cardAndProof := range tokenResp.TokenMap {
			dealCardMap[card].RevealToken = append(dealCardMap[card].RevealToken, cardAndProof)
		}
	}
	return mapx.Values(dealCardMap), nil

}

func (room *Room) action(pos int) {
	if room.allin+1 >= room.remain {
		return
	}

	skip := 0
	if pos == 0 { // start from left hand of button
		pos = (room.Button)%room.Cap() + 1
	}

	for {
		raised := 0

		room.Each(pos-1, func(o *Occupant) bool {
			if room.remain <= 1 {
				return false
			}

			if o.Pos == skip || o.Chips == 0 || len(o.Cards) == 0 {
				return true
			}

			room.Broadcast(&Message{
				From:   room.Id,
				Type:   MsgPresence,
				Action: ActAction,
				Class:  fmt.Sprintf("%d,%d", o.Pos, room.Bet),
			})

			msg, _ := o.GetAction(time.Duration(room.Timeout) * time.Second)
			if room.remain <= 1 {
				return false
			}

			n := 0
			// timeout or leave
			if msg == nil || len(msg.Class) == 0 {
				n = -1
			} else {
				n, _ = strconv.Atoi(msg.Class)
			}

			if room.betting(o.Pos, n) {
				raised = o.Pos
				return false
			}

			return true
		})

		if raised == 0 {
			break
		}

		pos = raised
		skip = pos
	}
}

func (room *Room) calc() (pots []handPot) {
	pots = calcPot(room.Chips)
	room.Pot = nil
	var ps []string
	for _, pot := range pots {
		room.Pot = append(room.Pot, pot.Pot)
		ps = append(ps, strconv.Itoa(pot.Pot))
	}

	room.Broadcast(&Message{
		From:   room.Id,
		Type:   MsgPresence,
		Action: ActPot,
		Class:  strings.Join(ps, ","),
	})

	return
}

func (room *Room) showdown() {
	pots := room.calc()

	for i, _ := range room.Chips {
		room.Chips[i] = 0
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	for _, pot := range pots {
		maxHand := 0
		for _, pos := range pot.OPos {
			o := room.Occupants[pos-1]
			if o != nil && o.Hand > maxHand {
				maxHand = o.Hand
			}
		}

		var winners []int

		for _, pos := range pot.OPos {
			o := room.Occupants[pos-1]
			if o != nil && o.Hand == maxHand && len(o.Cards) > 0 {
				winners = append(winners, pos)
			}
		}

		if len(winners) == 0 {
			fmt.Println("!!!no winners!!!")
			return
		}

		for _, winner := range winners {
			room.Chips[winner-1] += pot.Pot / len(winners)
		}
		room.Chips[winners[0]-1] += pot.Pot % len(winners) // odd chips
	}

	for i, _ := range room.Chips {
		if room.Occupants[i] != nil {
			room.Occupants[i].Chips += room.Chips[i]
		}
	}
}

func (room *Room) ready() {
	room.Bet = 0
	room.lock.Lock()
	defer room.lock.Unlock()

	room.Each(0, func(o *Occupant) bool {
		o.Bet = 0
		/*
			if o.Action == ActAllin || o.Action == ActFold || o.Action == "" {
				return true
			}
			o.Action = ActReady
		*/
		return true
	})

}

func (room *Room) betting(pos, n int) (raised bool) {
	if pos <= 0 {
		return
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	o := room.Occupants[pos-1]
	if o == nil {
		return
	}
	raised = o.Betting(n)
	if o.Action == ActFold {
		room.remain--
	}
	if o.Action == ActAllin {
		room.allin++
	}

	room.Broadcast(&Message{
		Id:     room.Id,
		Type:   MsgPresence,
		From:   o.Id,
		Action: ActBet,
		Class:  o.Action + "," + strconv.Itoa(o.Bet) + "," + strconv.Itoa(o.Chips),
	})

	return
}

func (room *Room) checkAndEndGame() {
	type UserSettle struct {
		Player     string `json:"player"`
		ChipAmount int    `json:"chip_amount"`
	}
	userSettles := []UserSettle{}

	hasChipUserCnt := 0
	playerCnt := 0
	for _, occupant := range room.Occupants {
		if occupant != nil && occupant.player != nil {
			playerCnt++
			if occupant.Chips > 0 {
				hasChipUserCnt++
			}
			userSettles = append(userSettles, UserSettle{
				Player:     occupant.Id,
				ChipAmount: occupant.Chips,
			})
		}
	}
	log.Println("userSettles:", userSettles, playerCnt, hasChipUserCnt)
	if len(userSettles) != 2 {
		log.Println("num not match %d", len(userSettles))
		return
	}

	// gameover
	if playerCnt > 0 && hasChipUserCnt <= 1 {
		room.Each(0, func(o *Occupant) bool {
			o.Leave()
			return true
		})
		log.Println("checkAndEndGame leave users")
		// call contract endgame
		cli := sui.NewSuiClient(constant.SuiTestnetEndpoint)
		signerAccount, err := signer.NewSignertWithMnemonic(mnemonic)
		if err != nil {
			log.Panicln(err)
		}
		rsp, err := cli.MoveCall(context.Background(), models.MoveCallRequest{
			Signer:          signerAccount.Address,
			PackageObjectId: MovePkgID,
			Module:          "mental_poker",
			Function:        "end_game",
			TypeArguments:   []interface{}{},
			Arguments: []interface{}{
				room.game.GameID,
				GameDataObjID,
				userSettles[0].Player,
				strconv.Itoa(int(userSettles[0].ChipAmount)),
				userSettles[1].Player,
				strconv.Itoa(int(userSettles[1].ChipAmount)),
				"proof", // todo proof
			},
			//Gas:       &gasObj,
			GasBudget: "100000000",
		})

		if err != nil {
			log.Println(room.game.GameID, GameDataObjID)
			log.Println("checkAndEndGame MoveCall", err.Error())
			return
		}
		// see the successful transaction url: https://explorer.sui.io/txblock/CD5hFB4bWFThhb6FtvKq3xAxRri72vsYLJAVd7p9t2sR?network=testnet
		rsp2, err := cli.SignAndExecuteTransactionBlock(context.Background(), models.SignAndExecuteTransactionBlockRequest{
			TxnMetaData: rsp,
			PriKey:      signerAccount.PriKey,
			// only fetch the effects field
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
			},
			RequestType: "WaitForLocalExecution",
		})
		if err != nil {
			log.Println("checkAndEndGame SignAndExecuteTransactionBlock", err.Error())
			return
		}
		log.Println(rsp2)
	}
}

type roomlist struct {
	M       map[string]*Room
	counter int
	lock    sync.Mutex
}

func NewRoomList() *roomlist {
	return &roomlist{
		M:       make(map[string]*Room),
		counter: 1000,
		lock:    sync.Mutex{},
	}
}

var (
	rooms = NewRoomList()
)

func SetRoom(room *Room) {
	rooms.lock.Lock()
	defer rooms.lock.Unlock()

	setRoom(room)
}

func setRoom(room *Room) {
	//id, _ := strconv.Atoi(room.Id)
	//if id == 0 {
	//	rooms.counter++
	//	id = rooms.counter
	//	room.Id = strconv.Itoa(id)
	//}
	rooms.M[room.Id] = room
}

func GetRoom(id string) *Room {
	rooms.lock.Lock()
	defer rooms.lock.Unlock()

	room := rooms.M[id]
	if room == nil {
		for _, v := range rooms.M {
			if v.N < v.Max {
				return v
			}
		}
		room = NewRoom(id, 9, 500, 1000)
		room.SetUpGame()
		setRoom(room)
	}

	return room
}

func GetOrCreateRoom(id string) *Room {
	if room := GetRoom(id); room != nil {
		return room
	}
	room := NewRoom(id, 9, 500, 1000)

	//if message.Room != nil {
	//	if message.Room.SB > 0 {
	//		room.SB = message.Room.SB
	//	}
	//	if message.Room.BB > 0 {
	//		room.BB = message.Room.BB
	//	}
	//	if message.Room.Timeout > 0 {
	//		room.Timeout = message.Room.Timeout
	//	}
	//
	//	if message.Room.Max > 0 && message.Room.Max <= MaxN {
	//		room.Max = message.Room.Max
	//		room.Occupants = room.Occupants[:room.Max]
	//		room.Chips = room.Chips[:room.Max]
	//	}
	//}
	SetRoom(room)
	return room
}

func DelRoom(room *Room) {
	rooms.lock.Lock()
	defer rooms.lock.Unlock()

	//id, _ := strconv.Atoi(room.Id)
	delete(rooms.M, room.Id)
	room.Id = ""
}

func Rooms() (r []*Room) {
	rooms.lock.Lock()
	defer rooms.lock.Unlock()

	r = make([]*Room, 0, len(rooms.M))

	ids := make([]string, 0, len(rooms.M))
	for k := range rooms.M {
		ids = append(ids, k)
	}

	//sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	for _, id := range ids {
		r = append(r, rooms.M[id])
	}

	return
}
