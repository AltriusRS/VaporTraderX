package services

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"vaportrader/src/constants"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

type SocketClient struct {
	Socket      *websocket.Conn
	OrderHook   func(order *SubscriptionsNewOrder)
	MessageHook func(message SocketPrivateMessage)
	PMHook      func(message *NewMessage)
	PendingPMs  map[string]*PendingMessage
	PMQueue     chan *SendMessage
	Status      UserStatus
	Session     *discordgo.Session
}

func NewSocketClient(s *discordgo.Session) *SocketClient {
	return &SocketClient{
		Socket:      nil,
		OrderHook:   nil,
		MessageHook: nil,
		PMHook:      nil,
		PMQueue:     make(chan *SendMessage),
		PendingPMs:  make(map[string]*PendingMessage),
		Status:      UserStatusUnknown,
		Session:     s,
	}
}

func (s *SocketClient) SendPM(message string, userID string) (*MessageAcknowledgement, error) {
	pm, channel, err := SendPM(message, userID)

	if err != nil {
		return nil, err
	}

	// Add the pending message to the map
	s.PendingPMs[pm.TempID] = channel

	// Queue the message to be sent
	s.PMQueue <- pm

	// Wait for confirmation
	result := <-channel.Confirmation

	return result, nil
}

func (s *SocketClient) SetPMHook(hook func(message *NewMessage)) {
	s.PMHook = hook
}

func (s *SocketClient) SetOrderHook(hook func(order *SubscriptionsNewOrder)) {
	s.OrderHook = hook
}

func (s *SocketClient) SetMessageHook(hook func(message SocketPrivateMessage)) {
	s.MessageHook = hook
}

func (s *SocketClient) SetStatus(status UserStatus) {
	s.Socket.WriteJSON(SocketMessage[UserStatus]{
		Type: "@WS/USER/SET_STATUS",
		Data: status,
	})
}

func (s *SocketClient) Subscribe(event string) {
	s.Socket.WriteJSON(SocketMessage[any]{
		Type: "@WS/SUBSCRIBE/" + event,
	})
}

var Socket *SocketClient

func InitSocket(s *discordgo.Session) {
	Socket = NewSocketClient(s)
	ConfigureSocket()
}

func ConfigureSocket() {
	u := url.URL{Scheme: "wss", Host: "warframe.market", Path: "/socket", RawQuery: "platform=pc"}
	log.Printf("connecting to %s", u.String())

	skt, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Cookie": []string{"JWT=" + os.Getenv("JWT")},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}

	Socket.Socket = skt

	done := make(chan struct{})

	Socket.Subscribe("MOST_RECENT")

	Socket.SetStatus(UserStatusInvisible)

	// time.Sleep(time.Second * 5)

	Socket.SetStatus(UserStatusOnline)

	log.Printf("connected to %s", u.String())

	go func() {
		defer close(done)
		for {
			msg, raw, err := Socket.read()
			if err != nil {
				log.Printf("Socket closed unexpectedly")
				log.Print(err.Error())
				go ConfigureSocket()
				return
			}

			// Send all messages to the message hook if present
			if Socket.MessageHook != nil {
				privMessage, err := PayloadFrom[NewMessage](raw)

				if err != nil {
					log.Printf("error unmarshaling message: %s", err)
					return
				}

				pm := ProcessPM(privMessage)
				go Socket.MessageHook(pm)
				continue
			}

			// Handle PM message delivery confirmations
			if msg.Type == "@WS/chats/MESSAGE_SENT" {
				ack, err := PayloadFrom[MessageAcknowledgement](raw)
				if err != nil {
					log.Printf("error unmarshaling message: %s", err)
					ack.Data.Success = false
					if _, ok := Socket.PendingPMs[ack.Data.TempID]; ok {
						Socket.PendingPMs[ack.Data.TempID].Success = true
						Socket.PendingPMs[ack.Data.TempID].Confirmation <- &ack.Data

						delete(Socket.PendingPMs, ack.Data.TempID)
					}
				} else {
					ack.Data.Success = true
					if _, ok := Socket.PendingPMs[ack.Data.TempID]; ok {
						Socket.PendingPMs[ack.Data.TempID].Success = true
						Socket.PendingPMs[ack.Data.TempID].Confirmation <- &ack.Data
						delete(Socket.PendingPMs, ack.Data.TempID)
					}
				}
			}

			// If we receive a new PM, check if we have a hook for it and process accordingly
			if Socket.PMHook != nil && msg.Type == "@WS/chats/NEW_MESSAGE" {
				nmsg, err := PayloadFrom[NewMessage](raw)
				if err != nil {
					log.Printf("error unmarshaling message: %s", err)
					return
				}
				go Socket.PMHook(&nmsg.Data)
				continue
			}

			// If we receive a new order, check if we have a hook for it and process accordingly
			if msg.Type == "@WS/SUBSCRIPTIONS/MOST_RECENT/NEW_ORDER" {
				if Socket.OrderHook != nil {
					order, err := PayloadFrom[SubscriptionsNewWrappedOrder](raw)
					if err != nil {
						log.Printf("error unmarshaling message: %s", err)
						return
					}
					go Socket.OrderHook(&order.Data.Order)
					continue
				}
			}

			log.Printf("unhandled message type: %s", msg.Type)
		}
	}()

	go func() {
		for pm := range Socket.PMQueue {
			log.Printf("sending pm to channel %s", pm.ChatID)
			log.Print(pm.Message)
			err := Socket.Socket.WriteJSON(SocketMessage[*SendMessage]{
				Type: "@WS/chats/SEND_MESSAGE",
				Data: pm,
			})

			if err != nil {
				log.Printf("error sending pm: %s", err)
			}
		}
	}()

	go func() {
		for {

			for key, pm := range Socket.PendingPMs {

				// Check if the message is older than 5 seconds
				cutoff := pm.Time.Add(time.Second * 5)

				if time.Now().After(cutoff) {
					pm.Confirmation <- &MessageAcknowledgement{
						Message: nil,
						TempID:  key,
						Success: false,
					}
					delete(Socket.PendingPMs, key)
				}
			}

			time.Sleep(time.Second * 1)
		}
	}()
}

func (s *SocketClient) read() (*SocketMessage[any], []byte, error) {
	_, message, err := s.Socket.ReadMessage()
	if err != nil {
		return nil, nil, err
	}

	var msg SocketMessage[any]

	err = json.Unmarshal(message, &msg)
	if err != nil {
		return nil, nil, err
	}

	return &msg, message, nil
}

func PayloadFrom[T any](msg []byte) (SocketMessage[T], error) {
	var msgStruct SocketMessage[T]
	err := json.Unmarshal(msg, &msgStruct)
	return msgStruct, err
}

func (s *SocketMessage[any]) Serialize() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func (m *NewMessage) Acknowledge() {
	Socket.Socket.WriteJSON(SocketMessage[ReadMessage]{
		Type: "@WS/chats/MESSAGE_WAS_READ",
		Data: ReadMessage{
			MessageID: m.ID,
		},
	})
}

func (m *NewMessage) Reply(message string) (*MessageAcknowledgement, error) {
	return Socket.SendPM(message+"\n\n"+constants.WFMFooter, m.ChatID)
}

func ProcessPM(msg SocketMessage[NewMessage]) SocketPrivateMessage {
	pm := SocketPrivateMessage{
		inner:  msg.Data,
		Text:   msg.Data.RawMessage,
		Author: msg.Data.MessageFrom,
	}

	pm.Acknowledge()

	return pm
}

func (pm *SocketPrivateMessage) Reply(message string) (*MessageAcknowledgement, error) {
	return Socket.SendPM(message, pm.inner.ChatID)
}

func (pm *SocketPrivateMessage) Acknowledge() {
	pm.inner.Acknowledge()
}
