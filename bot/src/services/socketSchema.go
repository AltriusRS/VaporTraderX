package services

import (
	"time"

	"github.com/teris-io/shortid"
)

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

type SocketMessage[T any] struct {
	Type string `json:"type"`
	Data T      `json:"payload"`
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

type SocketOnlineCount struct {
	TotalUsers      int `json:"total_users"`
	RegisteredUsers int `json:"registered_users"`
}
