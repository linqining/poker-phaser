package poker

type PlayerInfo struct {
	RoomID string
	Chips  int
	UserID string
}

func GetPlayerInfo(addr string) *PlayerInfo {
	rooms.lock.RLock()
	defer rooms.lock.RUnlock()
	for roomid, room := range rooms.M {
		for _, occupant := range room.Occupants {
			if occupant != nil && occupant.Id == addr {
				return &PlayerInfo{
					RoomID: roomid,
					Chips:  occupant.Chips,
					UserID: occupant.Id,
				}
			}
		}
	}
	return nil
}
