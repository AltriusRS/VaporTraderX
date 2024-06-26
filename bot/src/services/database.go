package services

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database is a wrapper around sql.DB that provides some ORM focused methods for ease of use
type Database struct {
	Inner *gorm.DB
}

func NewDatabase() *Database {

	db, err := gorm.Open(postgres.Open(os.Getenv("DB_STRING")), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	sqldb, err := db.DB()

	if err != nil {
		log.Fatal(err)
	}

	sqldb.SetMaxOpenConns(20)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxLifetime(time.Hour)

	db.AutoMigrate(&User{}, &Badge{}, &Award{}, &Alert{}, &Trade{}, &Item{}, &ItemTranslation{}, &StateInfo{}, &TradeInfo{})

	return &Database{Inner: db}
}

func (db *Database) GetUserByID(id string) (*User, error) {
	var user User

	err := db.Inner.Find(&user, "id = ?", id).Limit(1).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &user, nil
}

func (db *Database) GetUserByWFMID(wfmid string) (*User, error) {
	var user User

	err := db.Inner.Find(&user, "wfm_id = ?", wfmid).Limit(1).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &user, nil
}

func (db *Database) InsertItem(item *Item) error {

	tx := db.Inner.Begin()

	translations := item.Translations

	item.Translations = []ItemTranslation{}

	err := tx.Save(item).Error

	if err != nil {
		_ = tx.Rollback().Error
		return err
	}

	for _, translation := range translations {
		err = tx.Save(&translation).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit().Error

	if err != nil {
		_ = tx.Rollback().Error
		return err
	}

	return nil
}

func (db *Database) InsertOrder(order *SubscriptionsNewOrder) error {
	kind := 0

	switch order.OrderType {
	case "buy":
		kind = 0
	case "sell":
		kind = 1
	}

	trade := Trade{
		ID:     order.ID,
		UserId: order.User.ID,
		ItemId: order.Item.ID,
		Kind:   uint8(kind),
		Price:  uint32(order.Price),
	}

	return db.Inner.Create(&trade).Error
}

func (db *Database) GetLastSynced() (time.Time, time.Time, error) {
	var stateInfo StateInfo

	err := db.Inner.First(&stateInfo).Error

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return stateInfo.LastSynced, stateInfo.DeepSynced, nil
}

func (db *Database) SetLastSynced(lastSynced time.Time, deep bool) error {
	if deep {
		stateInfo := StateInfo{
			ID:         1,
			LastSynced: lastSynced,
			DeepSynced: lastSynced,
		}
		return db.Inner.Save(&stateInfo).Error
	} else {
		stateInfo := StateInfo{
			ID:         1,
			LastSynced: lastSynced,
		}
		return db.Inner.Save(&stateInfo).Error
	}

}

var DB *Database

func (db *Database) Create(data interface{}) error {
	return db.Inner.Create(data).Error
}

func (db *Database) Save(data interface{}) error {
	return db.Inner.Save(data).Error
}

func (u *User) HasPermission(entitlement string) bool {

	entitlement = strings.ToLower(entitlement)

	mask, ok := Entitlements[entitlement]

	if !ok {
		return false
	}

	return u.Entitlements&mask != 0
}

func (u *User) GrantPermission(entitlement string) {

	entitlement = strings.ToLower(entitlement)

	mask, ok := Entitlements[entitlement]

	if !ok {
		return
	}

	u.Entitlements |= mask
}

func (u *User) RevokePermission(entitlement string) {

	entitlement = strings.ToLower(entitlement)

	mask, ok := Entitlements[entitlement]

	if !ok {
		return
	}

	u.Entitlements &= ^mask
}

func (i *Trade) AfterSave(tx *gorm.DB) (err error) {
	tx.Create(&TradeInfo{
		Time:        i.CreatedAt,
		ItemId:      i.ItemId,
		Price:       i.Price,
		IsSellOrder: i.Kind == 1,
	})

	return nil
}

func (i *Item) BeforeDelete(tx *gorm.DB) (err error) {
	translations := []ItemTranslation{}
	err = tx.Find(&translations).Error

	if err != nil {
		return err
	}

	for _, translation := range translations {
		err = tx.Delete(&translation).Error

		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func InitDatabase() {
	DB = NewDatabase()
}
