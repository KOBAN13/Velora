package packets

type Msg = isPacket_Msg

func NewChat(msg string) Msg {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Msg: msg,
		},
	}
}

func NewId(id uint64) Msg {
	return &Packet_Id{
		Id: &IdMessage{
			Id: id,
		},
	}
}

func NewOkResponse() Msg {
	return &Packet_OkResponse{
		OkResponse: &OkResponseMessage{},
	}
}

func NewDenyResponse(reason string) Msg {
	return &Packet_DenyResponse{
		DenyResponse: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewCreateRoomRequest(maxPlayer uint32) Msg {
	return &Packet_CreateRoomRequest{
		CreateRoomRequest: &CreateRoomRequestMessage{
			MaxPlayer: maxPlayer,
		},
	}
}

func NewJoinRoomRequest(roomId uint64) Msg {
	return &Packet_JoinRoomRequest{
		JoinRoomRequest: &JoinRoomRequestMessage{
			RoomId: roomId,
		},
	}
}

func NewLeaveRoomRequest() Msg {
	return &Packet_LeaveRoomRequest{
		LeaveRoomRequest: &LeaveRoomRequestMessage{},
	}
}

func NewReadyRequest(isReady bool) Msg {
	return &Packet_ReadyRequest{
		ReadyRequest: &ReadyRequestMessage{
			IsReady: isReady,
		},
	}
}

func NewPlayerInputMessage(matchId uint64, position *Vector2Message) Msg {
	return &Packet_PlayerInput{
		PlayerInput: &PlayerInputMessage{
			MatchId:      matchId,
			MovePosition: position,
		},
	}
}

func NewMatchSnapshot(matchId uint64, serverTick uint64, phase MatchPhase, phaseTimeLeft int64) Msg {
	return &Packet_MatchSnapshot{
		MatchSnapshot: &MatchSnapshotMessage{
			MatchId:         matchId,
			ServerTick:      serverTick,
			Phase:           phase,
			PhaseTimeLeftMs: phaseTimeLeft,
		},
	}
}

func NewMatchStarted(roomId uint64, matchId uint64, playerId uint64, slot uint32, mapSeed uint64, startsAtUnixMs int64) Msg {
	return &Packet_MatchStarted{
		MatchStarted: &MatchStartMessage{
			RoomId:         roomId,
			MatchId:        matchId,
			PlayerId:       playerId,
			Slot:           slot,
			MapSeed:        mapSeed,
			StartsAtUnixMs: startsAtUnixMs,
		},
	}
}

func NewRoomSummaryMessage(roomId uint64, roomName string, players []*RoomPlayerMessage, maxPlayer uint32, status RoomStatus) *RoomSummaryMessage {
	return &RoomSummaryMessage{
		Name:      roomName,
		RoomId:    roomId,
		Players:   players,
		MaxPlayer: maxPlayer,
		Status:    status,
	}
}

func NewRoomListSnapshotMessage(rooms []*RoomSummaryMessage) Msg {
	return &Packet_RoomListSnapshot{
		RoomListSnapshot: &RoomListSnapshotMessage{
			Rooms: rooms,
		},
	}
}

func NewRoomStateSnapshot(roomId uint64, maxPlayer uint32, status RoomStatus, players []*RoomPlayerMessage) Msg {
	return &Packet_RoomStateSnapshot{
		RoomStateSnapshot: &RoomStateSnapshotMessage{
			RoomId:    roomId,
			MaxPlayer: maxPlayer,
			Status:    status,
			Player:    players,
		},
	}
}
