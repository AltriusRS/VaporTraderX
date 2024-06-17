package services

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserEntitlements map[string]uint32

var Entitlements = UserEntitlements{
	"none":      0 << 0, // Enable the 1st bit
	"admin":     0 << 1, // Enable the 2nd bit
	"moderator": 0 << 2, // Enable the 3rd bit
	"developer": 0 << 3, // Enable the 4th bit
	"premium1":  0 << 4, // Enable the 5th bit
	"premium2":  0 << 5, // Enable the 6th bit
	"premium3":  0 << 6, // Enable the bit
	"wfm staff": 0 << 7, // Enable the bit
}

// A struct to represent a user's VaporTrader account
type User struct {
	gorm.Model
	ID                string `gorm:"primaryKey unique"`
	Name              string
	Entitlements      uint32
	Locale            sql.NullString
	WfmID             sql.NullString `gorm:"unique column:wfm_id"`
	WfmUsername       sql.NullString `gorm:"unique column:wfm_username"`
	PreferredPlatform sql.NullString
	FirstSeen         time.Time `gorm:"autoCreateTime"` // the first time this user was seen (either on the socket, or the bot)
	LastSeen          time.Time // the last time this user was seen (either on the socket, or the bot)
	Awards            []Award   `gorm:"foreignkey:UserId"`
	Alerts            []Alert   `gorm:"foreignkey:UserId"`
}

type Badge struct {
	gorm.Model
	ID          uint32 `gorm:"'type:Int4' primaryKey unique autoIncrement"` // The unique ID of this badge
	Name        string `gorm:"unique"`
	Description string `gorm:"unique"`
	Icon        string `gorm:"unique"`
	Achievable  bool
}

type Award struct {
	gorm.Model
	ID      uint32 `gorm:"'type:Int4' primaryKey unique autoIncrement"` // The unique ID of this award
	UserId  string `gorm:"index"`
	User    User   `gorm:"references:ID"`
	BadgeId int32  `gorm:"index"`
	Badge   Badge  `gorm:"references:ID"`
}

// A struct to represent a user's VaporTrader alerts
type Alert struct {
	gorm.Model
	ID            uint32 `gorm:"'type:Int4' primaryKey unique autoIncrement"` // The unique ID of this alert
	UserId        string `gorm:"index"`
	User          User   `gorm:"references:ID"`
	ItemId        string `gorm:"index"`
	Item          Item   `gorm:"references:ID"`
	OrderType     string // The order type of this alert ()
	PriceMode     string // The price mode of this alert ()
	LowerPrice    uint32 // The minimum price of this alert (inclusive)
	UpperPrice    uint32 // The maximum price of this alert (inclusive)
	PriceVariance uint32 // The price variance to trigger the alert
	Platform      string // The platform this alert applies to
	Hits          int32  `gorm:"'type:Int4' 'default:0'"`
	Active        bool   `gorm:"default:true"`
}

// A struct to represent a trade
type Trade struct {
	gorm.Model
	ID     string `gorm:"primaryKey unique autoIncrement"` // The unique ID of this trade
	UserId string `gorm:"index"`
	ItemId string `gorm:"index"`
	Item   Item   `gorm:"references:ID"`
	Kind   uint8
	Price  uint32 `gorm:"type:Int4"` // The price of this trade
}

// A struct to represent an item from the API
type Item struct {
	gorm.Model
	ID           string `gorm:"primaryKey unique"` // The unique ID of this item - Assigned by WFM
	Slug         string `gorm:"unique"`
	IsSet        bool   `gorm:"default:false"`
	PartOf       sql.NullString
	Icon         sql.NullString
	SubIcon      sql.NullString
	Thumbnail    sql.NullString
	IconFormat   sql.NullString
	NumberForSet uint8             `gorm:"default:1"`
	MasteryLevel uint8             `gorm:"default:0"`
	Ducats       uint32            `gorm:"'type:Int4' 'default:0'"`
	TradeTax     uint32            `gorm:"'type:Int4' 'default:0'"`
	Tags         pq.StringArray    `gorm:"type:text[]"`
	Vaulted      bool              `gorm:"default:false"`
	Translations []ItemTranslation `gorm:"foreignkey:ItemId"`
}

// A struct to represent a translation of an item
type ItemTranslation struct {
	gorm.Model
	ID          string `gorm:"primaryKey unique autoIncrement"` // The unique ID of this translation
	ItemId      string `gorm:"index"`
	Locale      string `gorm:"index"`
	Name        string
	Description string
	WikiLink    string
	Thumbnail   string
	Icon        string
	// DropTables  []string // This field is not populated by the api, so we don't need to worry about it
}

// A struct to represent the state of the bot
type StateInfo struct {
	gorm.Model
	ID         uint32 `gorm:"'type:Int4' primaryKey autoIncrement"`
	LastSynced time.Time
}

type TradeReports struct {
	gorm.Model
	Hour     time.Time ``
	Listings uint32    `gorm:"type:Int4"`
	Buys     uint32    `gorm:"type:Int4"`
	Sells    uint32    `gorm:"type:Int4"`
}
