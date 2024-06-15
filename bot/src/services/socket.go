package services

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
)

type SocketClient struct {
	Socket      *websocket.Conn
	OrderHook   func(order *SubscriptionsNewOrder)
	MessageHook func(message SocketPrivateMessage)
	PMHook      func(message *NewMessage)
	PendingPMs  map[string]*PendingMessage
	PMQueue     chan *SendMessage
	Status      UserStatus
}

func NewSocketClient() *SocketClient {
	return &SocketClient{
		Socket:      nil,
		OrderHook:   nil,
		MessageHook: nil,
		PMHook:      nil,
		PMQueue:     make(chan *SendMessage),
		PendingPMs:  make(map[string]*PendingMessage),
		Status:      UserStatusUnknown,
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

var Socket = NewSocketClient()

type UserStatus string

const (
	UserStatusOnline    = UserStatus("online")
	UserStatusInvisible = UserStatus("invisible")
	UserStatusUnknown   = UserStatus("ingame")
)

type OrderType string

const (
	OrderTypeSell  = OrderType("sell")
	OrderTypeBuy   = OrderType("buy")
	OrderTypeTrade = OrderType("trade")
)

func InitSocket() {
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
				log.Printf("error reading socket: %s", err)
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
				Socket.MessageHook(pm)
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
				Socket.PMHook(&nmsg.Data)
			}

			// If we receive a new order, check if we have a hook for it and process accordingly
			if msg.Type == "@WS/SUBSCRIPTIONS/MOST_RECENT/NEW_ORDER" {
				if Socket.OrderHook != nil {
					order, err := PayloadFrom[SubscriptionsNewWrappedOrder](raw)
					if err != nil {
						log.Printf("error unmarshaling message: %s", err)
						return
					}
					Socket.OrderHook(&order.Data.Order)
				}
			}

			// Only check for new messages every 100 milliseconds
			// This is to avoid over-utilising the thread
			time.Sleep(time.Millisecond * 100)
		}
	}()

	go func() {
		for pm := range Socket.PMQueue {
			log.Printf("sending pm to channel %s", pm.ChatID)
			Socket.Socket.WriteJSON(SocketMessage[*SendMessage]{
				Type: "@WS/chats/SEND_MESSAGE",
				Data: pm,
			})
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

type SocketMessage[T any] struct {
	Type string `json:"type"`
	Data T      `json:"payload"`
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

// Platform Orders

type SubscriptionsNewWrappedOrder struct {
	Order SubscriptionsNewOrder `json:"order"`
}

type SubscriptionsNewOrder struct {
	ID           string       `json:"id"`            // The ID of the order.
	CreationDate time.Time    `json:"creation_date"` // The date the order was created.
	LastModified time.Time    `json:"last_update"`   // The date the order was last updated.
	ModRank      int          `json:"mod_rank"`      // The rank of the user in the order.
	Price        int          `json:"platinum"`      // The price of the order.
	OrderType    string       `json:"order_type"`    // The type of the order. Can be "sell", or "buy".
	Quantity     int          `json:"quantity"`      // The quantity of the order.
	Visible      bool         `json:"visible"`       // Whether the order is visible.
	Region       string       `json:"region"`        // The region of the order.
	Platform     string       `json:"platform"`      // The platform of the order.
	User         PlatformUser `json:"user"`          // The user who placed the order.
	Item         PlatformItem `json:"item"`          // The item being ordered.
}

// Platform User Info

type PlatformUser struct {
	ID         string    `json:"id"`          // The user's ID.
	Locale     string    `json:"locale"`      // The user's locale.
	Avatar     string    `json:"avatar"`      // The user's avatar.
	LastSeen   time.Time `json:"last_seen"`   // The time the user was last seen.
	GameName   string    `json:"ingame_name"` // The name of the player in-game.
	Reputation int       `json:"reputation"`  // The user's reputation.
	Region     string    `json:"region"`      // The region of the user.
	Status     string    `json:"status"`      // The status of the user. Can be "online", "invisible", or "ingame".
}

// Platform Item Info

type PlatformItem struct {
	ID         string              `json:"id"`           // The ID of the item.
	Thumbnail  string              `json:"thumb"`        // The thumbnail of the item.
	Icon       string              `json:"icon"`         // The icon of the item.
	SubIcon    string              `json:"sub_icon"`     // The sub-icon of the item.
	IconFormat string              `json:"icon_format"`  // The format of the icon.
	Tags       []string            `json:"tags"`         // The tags of the item.
	MaxModRank int                 `json:"mod_max_rank"` // The maximum rank of the mod being sold.
	Slug       string              `json:"url_name"`     // The URL slug of the item.
	EN         PlatformTranslation `json:"en"`           // Platform translations for the English language
	RU         PlatformTranslation `json:"ru"`           // Platform translations for the Russian language
	KO         PlatformTranslation `json:"ko"`           // Platform translations for the Korean language
	FR         PlatformTranslation `json:"fr"`           // Platform translations for the French language
	SV         PlatformTranslation `json:"sv"`           // Platform translations for the Swedish language
	DE         PlatformTranslation `json:"de"`           // Platform translations for the German language
	ZHHant     PlatformTranslation `json:"zh-hant"`      // Platform translations for the Chinese (Traditional) language
	ZHHans     PlatformTranslation `json:"zh-hans"`      // Platform translations for the Chinese (Simplified) language
	PT         PlatformTranslation `json:"pt"`           // Platform translations for the Portuguese language
	ES         PlatformTranslation `json:"es"`           // Platform translations for the Spanish language
	PL         PlatformTranslation `json:"pl"`           // Platform translations for the Polish language
	CS         PlatformTranslation `json:"cs"`           // Platform translations for the Czech language
	UK         PlatformTranslation `json:"uk"`           // Platform translations for the Ukrainian language
}

type PlatformTranslation struct {
	Name string `json:"item_name"` // The name of the item in the given language.
}

// Platform Direct Messages

type SendMessage struct {
	ChatID  string `json:"chat_id"` // The ID of the chat to send the message in.
	Message string `json:"message"` // The message to send.
	TempID  string `json:"temp_id"` // The temporary ID of the message.
}

func SendPM(message string, chatID string) (*SendMessage, *PendingMessage, error) {
	ack := &PendingMessage{
		Confirmation: make(chan *MessageAcknowledgement),
		Time:         time.Now(),
		Success:      false,
	}
	tid, err := shortid.Generate()

	if err != nil {
		return nil, nil, err
	}

	return &SendMessage{
		ChatID:  chatID,
		Message: message,
		TempID:  tid,
	}, ack, nil
}

type NewMessage struct {
	Message     string `json:"message"`      // The message formatted with HTML.
	RawMessage  string `json:"raw_message"`  // The message as it was typed.
	MessageFrom string `json:"message_from"` // The user who sent the message. (ID)
	Timestamp   string `json:"send_date"`    // The date the message was sent.
	ID          string `json:"id"`           // The ID of the message.
	ChatID      string `json:"chat_id"`      // The ID of the chat the message was sent in.
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
	return Socket.SendPM(message, m.ChatID)
}

type ReadMessage struct {
	MessageID string `json:"message_id"` // The Id of the message to mark as read.
}

type MessageAcknowledgement struct {
	Message *NewMessage `json:"message"` // The message that was sent.
	TempID  string      `json:"temp_id"` // The temporary ID of the message.
	Success bool        `json:"success"` // Whether the message was sent successfully.
}

type PendingMessage struct {
	Confirmation chan *MessageAcknowledgement // Channel to send confirmation messages to
	Time         time.Time                    // Time the message was scheduled to be sent
	Success      bool                         // Whether the message was sent successfully
}

type SocketPrivateMessage struct {
	inner  NewMessage
	Text   string
	Author string
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
