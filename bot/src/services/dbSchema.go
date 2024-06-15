package services

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type UserEntitlement uint16

type UserEntitlements map[string]UserEntitlement

var Entitlements = UserEntitlements{
	"none":      UserEntitlement(0 << 0), // Enable the 1st bit
	"admin":     UserEntitlement(0 << 1), // Enable the 2nd bit
	"moderator": UserEntitlement(0 << 2), // Enable the 3rd bit
	"developer": UserEntitlement(0 << 3), // Enable the 4th bit
	"premium1":  UserEntitlement(0 << 4), // Enable the 5th bit
	"premium2":  UserEntitlement(0 << 5), // Enable the 6th bit
	"premium3":  UserEntitlement(0 << 6), // Enable the bit
	"wfm staff": UserEntitlement(0 << 7), // Enable the bit
}

// A struct to represent a user's VaporTrader account
type User struct {
	gorm.Model
	ID                string          `gorm:"primaryKey unique"` //
	Name              string          `gorm:"<-:update"`         //
	Entitlements      UserEntitlement `gorm:"<-:update"`         //
	Locale            sql.NullString  `gorm:"<-:update"`         //
	WFMID             sql.NullString  `gorm:"<-:update"`         //
	PreferredPlatform sql.NullString  `gorm:"<-:update"`         //
	FirstSeen         time.Time       `gorm:"autoCreateTime"`    // the first time this user was seen (either on the socket, or the bot)
	LastSeen          time.Time       `gorm:"<-:update"`         // the last time this user was seen (either on the socket, or the bot)
	UpdatedAt         time.Time       `gorm:"autoUpdateTime"`    // the last time this user was updated
	Awards            []Award         `gorm:"foreignkey:UserId"`
	Alerts            []Alert         `gorm:"foreignkey:UserId"`
}

type Badge struct {
	gorm.Model
	ID          int    `gorm:"primaryKey unique autoIncrement"` // The unique ID of this badge
	Name        string `gorm:"unique"`
	Description string `gorm:"unique"`
	Icon        string `gorm:"unique"`
	Achievable  bool
	AddedAt     time.Time `gorm:"autoCreateTime"` // The time this badge was added
	UpdatedAt   time.Time `gorm:"autoUpdateTime"` // The last time this badge was modified
}

type Award struct {
	gorm.Model
	ID        int       `gorm:"primaryKey unique autoIncrement"` // The unique ID of this award
	UserId    string    `gorm:"index"`
	User      User      `gorm:"references:ID"`
	BadgeId   int       `gorm:"index"`
	Badge     Badge     `gorm:"references:ID"`
	CreatedAt time.Time `gorm:"autoCreateTime"` // The time this award was created
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // The last time this award was modified
}

// A struct to represent a user's VaporTrader alerts
type Alert struct {
	gorm.Model
	ID            int       `gorm:"primaryKey unique autoIncrement"` // The unique ID of this alert
	UserId        string    `gorm:"index"`
	User          User      `gorm:"references:ID"`
	ItemId        string    `gorm:"index"`
	Item          Item      `gorm:"references:ID"`
	OrderType     string    // The order type of this alert ()
	PriceMode     string    // The price mode of this alert ()
	LowerPrice    uint32    // The minimum price of this alert (inclusive)
	UpperPrice    uint32    // The maximum price of this alert (inclusive)
	PriceVariance uint32    // The price variance to trigger the alert
	Platform      string    // The platform this alert applies to
	Hits          int       `gorm:"default:0"`
	Active        bool      `gorm:"default:true"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"` // The last time this alert was modified
	CreatedAt     time.Time `gorm:"autoCreateTime"` // The time this alert was created
}

// A struct to represent a trade
type Trade struct {
	gorm.Model
	ID        int       `gorm:"primaryKey unique autoIncrement"` // The unique ID of this trade
	UserId    string    `gorm:"index"`
	User      User      `gorm:"references:ID"`
	ItemId    string    `gorm:"index"`
	Item      Item      `gorm:"references:ID"`
	Price     uint32    // The price of this trade
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// A struct to represent an item from the API
type Item struct {
	gorm.Model
	ID           string `gorm:"primaryKey unique"` //
	Slug         string `gorm:"unique"`
	IsSet        bool
	PartOf       sql.NullString
	Icon         sql.NullString
	SubIcon      sql.NullString
	Thumbnail    sql.NullString
	IconFormat   sql.NullString
	NumberForSet sql.NullInt16
	MasteryLevel sql.NullInt64
	Ducats       int
	TradeTax     int
	Tags         sql.NullString
	Vaulted      sql.NullBool
	Translations []ItemTranslation `gorm:"foreignkey:ItemId"`
}

// A struct to represent a translation of an item
type ItemTranslation struct {
	gorm.Model
	ID          int `gorm:"primaryKey unique autoIncrement"`
	ItemId      string
	Item        Item `gorm:"references:ID"`
	Locale      string
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
	ID         int `gorm:"primaryKey autoIncrement"`
	LastSynced time.Time
}
