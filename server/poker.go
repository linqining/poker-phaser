package poker

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Poker struct {
	WebRoot string
	Addr    string
	OnAuth  func(conn *Conn, mechanism, text string) (*Occupant, error)
	OnExit  func(o *Occupant)
}

func (p *Poker) ListenAndServe() error {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	return origin == "https://github.com"
		//},
		MaxAge: 12 * time.Hour,
	}))

	r.GET("/reconnect/:user_addr", func(c *gin.Context) {
		address := c.Param("user_addr")
		playerInfo := GetPlayerInfo(address)
		if playerInfo != nil {
			c.JSON(http.StatusOK, gin.H{
				"room_id": playerInfo.RoomID,
				"user_id": playerInfo.UserID,
				"chips":   playerInfo.Chips,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"room_id": "",
			"user_id": "",
			"chips":   0,
		})
	})
	r.GET("/ws", func(c *gin.Context) {
		p.pokerHandler(c.Writer, c.Request)
	})
	return r.Run(fmt.Sprintf("%s", p.Addr)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (p *Poker) pokerHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer ws.Close()

	ws.SetReadLimit(maxMessageSize)
	ws.SetPongHandler(
		func(string) error {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	conn := NewConn(ws, 128)
	defer conn.Close()

	auth := &Auth{}
	if err := conn.ReadJSONTimeout(auth, readWait); err != nil {
		return
	}

	var o *Occupant
	if p.OnAuth != nil {
		o, err = p.OnAuth(conn, auth.Mechanism, auth.Text)
		if err != nil {
			log.Println(err)
			o = NewOccupant(strconv.FormatInt(time.Now().Unix(), 10), conn)
			o.Name = auth.Text
		}
		o.Chips = 10000
	}

	if err := conn.WriteJSON(o); err != nil {
		return
	}

	for {
		message, _ := o.GetMessage(0)
		if message == nil {
			break
		}

		switch message.Type {
		case MsgIQ:
			handleIQ(o, message)
		case MsgPresence:
			handlePresence(o, message)
		case MsgMessage:
		}
	}

	o.Leave()

	if p.OnExit != nil {
		p.OnExit(o)
	}
	log.Println(o.Name, "disconnected.")
}
