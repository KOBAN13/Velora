package packets

import "Velora/server/Internal/server/lobby"

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

func NewJoinRoomRequest(roomId uint32) Msg {
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

func NewRoomPlayerMessage(roomPlayer *lobby.RoomPlayer) *RoomPlayerMessage {
	return &RoomPlayerMessage{
		UserId:   roomPlayer.UserID,
		ClientId: roomPlayer.ClientID,
		Username: roomPlayer.Username,
		IsReady:  roomPlayer.IsReady,
		Owner:    roomPlayer.IsOwner,
	}
}

func NewMatchStarting(roomId uint64, startsAtUnixMs int64) Msg {
	return &Packet_MatchStarting{
		MatchStarting: &MatchStartingMessage{
			RoomId:         roomId,
			StartsAtUnixMs: startsAtUnixMs,
		},
	}
}

func NewMatchStarted(roomId uint64, matchId uint64) Msg {
	return &Packet_MatchStarted{
		MatchStarted: &MatchStartMessage{
			RoomId:  roomId,
			MatchId: matchId,
		},
	}
}
