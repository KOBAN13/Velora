package states

import (
	"Velora/server/Internal/server"
	"Velora/server/Internal/server/db"
	"Velora/server/pkg/packets"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Connection struct {
	client server.ClientInterface
	logger *log.Logger
}

func (conn *Connection) Name() string {
	return "Connection"
}

func (conn *Connection) SetClientInterface(client server.ClientInterface) {
	conn.client = client

	var loggerPrefix = fmt.Sprintf("Client %d [%s]", client.Id(), conn.Name())

	conn.logger = log.New(log.Writer(), loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (conn *Connection) HandleMessage(id uint64, msg packets.Msg) {
	switch message := msg.(type) {
	case *packets.Packet_LoginRequest:
		conn.handleLoginRequest(id, message.LoginRequest)
		break
	case *packets.Packet_RegisterRequest:
		conn.handleRegisterRequest(id, message.RegisterRequest)
		break
	case *packets.Packet_Chat:
		conn.handleChatMessage(id, message)
	}
}

func (conn *Connection) OnEnter() {
	var id = conn.client.Id()

	var clientId = packets.NewId(id)

	conn.client.SocketSend(clientId)
	conn.logger.Printf("Client initialized and send to client id: %v", clientId)
}

func (conn *Connection) OnLeave() {

}

func (conn *Connection) handleChatMessage(id uint64, msg packets.Msg) {
	if id == conn.client.Id() {
		conn.client.Broadcast(msg)
		return
	}

	conn.client.SocketSendAs(msg, id)
}

func (conn *Connection) handleLoginRequest(senderId uint64, msg *packets.LoginRequestMessage) {
	if senderId != conn.client.Id() {
		conn.logger.Printf("Invalid client id: %v", senderId)
		return
	}

	var username = msg.Username

	var denyMessage = packets.NewDenyResponse("Invalid username or password")

	var dBtX = conn.client.DbTx()

	var user, err = dBtX.UserRepository.GetUserByUsername(dBtX.Ctx, strings.ToLower(username))

	if err != nil {
		conn.logger.Printf("Error getting user by username: %v", err)
		conn.client.SocketSend(denyMessage)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.Password))

	if err != nil {
		conn.logger.Printf("Incorrect password for user: %v", username)
		conn.client.SocketSend(denyMessage)
		return
	}

	conn.logger.Printf("Successfully authenticated user: %v", username)
	conn.client.SocketSend(packets.NewOkResponse())
}

func (conn *Connection) handleRegisterRequest(senderId uint64, msg *packets.RegisterRequestMessage) {
	if senderId != conn.client.Id() {
		conn.logger.Printf("Invalid client id: %v", senderId)
		return
	}

	var username = msg.Username
	var password = msg.Password

	var err = validateUsername(username)

	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		conn.logger.Println(reason)
		conn.client.SocketSend(packets.NewDenyResponse(reason))
		return
	}

	var dBtX = conn.client.DbTx()

	_, err = dBtX.UserRepository.GetUserByUsername(dBtX.Ctx, strings.ToLower(username))

	if err == nil {
		var denyMessage = packets.NewDenyResponse("This user already exists")
		conn.logger.Printf("This user already exists : %v", err)
		conn.client.SocketSend(denyMessage)
		return
	}

	var denyMessage = packets.NewDenyResponse("Failed to register user (internal server error)")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		conn.logger.Printf("Error hashing password: %v", err)
		conn.client.SocketSend(denyMessage)
		return
	}

	_, err = dBtX.UserRepository.CreateUser(dBtX.Ctx, db.CreateUserParams{
		Username:     strings.ToLower(username),
		PasswordHash: string(passwordHash),
	})

	if err != nil {
		conn.logger.Printf("Failed to create user: %v", err)
		conn.client.SocketSend(denyMessage)
		return
	}

	conn.logger.Printf("Successfully register user: %v", username)
	conn.client.SocketSend(packets.NewOkResponse())
}

func validateUsername(username string) error {
	if len(username) <= 0 {
		return errors.New("username cannot be empty")
	}

	if len(username) > 20 {
		return errors.New("username cannot be longer than 20 characters")
	}

	if username != strings.TrimSpace(username) {
		return errors.New("username does not match")
	}

	return nil
}
