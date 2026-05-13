package lobby

type RoomStatus int

const (
	ROOM_STATUS_WAITING RoomStatus = iota
	ROOM_STATUS_STARTING
	ROOM_STATUS_STARTED
)
