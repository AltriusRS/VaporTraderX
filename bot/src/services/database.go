package services

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Database is a wrapper around sql.DB that provides some ORM focused methods for ease of use
type Database struct {
	db *gorm.DB
}

func NewDatabase() *Database {

	db, err := gorm.Open(sqlite.Open("./vp.db"), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Badge{}, &Award{}, &Alert{}, &Trade{}, &Item{}, &ItemTranslation{}, &StateInfo{})

	return &Database{db: db}
}

func (db *Database) GetUserByID(id string) (*User, error) {
	var user User

	err := db.db.First(&user, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *Database) GetUserByWFMID(wfmid string) (*User, error) {
	var user User

	err := db.db.First(&user, "wfmid = ?", wfmid).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *Database) InsertItem(item *Item) error {

	tx := db.db.Begin()

	translations, err := tx.Select("*").Where("item_id = ?", item.ID).Find(&ItemTranslation{}).Rows()

	if err != nil {
		tx.Rollback()
		return err
	}

	for translations.Next() {
		var translation ItemTranslation
		translations.Scan(&translation)
		err = tx.Unscoped().Delete(&translation, clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{
					Column: clause.Column{
						Name: "item_id",
					},
					Value: item.ID,
				},
			},
		}).Error

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, translation := range item.Translations {
		err = tx.Create(&translation).Error

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: []clause.Assignment{
			{Column: clause.Column{Name: "slug"}, Value: item.Slug},
			{Column: clause.Column{Name: "is_set"}, Value: item.IsSet},
			{Column: clause.Column{Name: "part_of"}, Value: item.PartOf},
			{Column: clause.Column{Name: "icon"}, Value: item.Icon},
			{Column: clause.Column{Name: "sub_icon"}, Value: item.SubIcon},
			{Column: clause.Column{Name: "thumbnail"}, Value: item.Thumbnail},
			{Column: clause.Column{Name: "icon_format"}, Value: item.IconFormat},
			{Column: clause.Column{Name: "mastery_level"}, Value: item.MasteryLevel},
			{Column: clause.Column{Name: "ducats"}, Value: item.Ducats},
			{Column: clause.Column{Name: "trade_tax"}, Value: item.TradeTax},
			{Column: clause.Column{Name: "tags"}, Value: item.Tags},
			{Column: clause.Column{Name: "vaulted"}, Value: item.Vaulted},
		},
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: clause.Column{Name: "id"}, Value: item.ID},
		}},
	}).Create(item).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error

	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (db *Database) GetLastSynced() (time.Time, error) {
	var stateInfo StateInfo

	err := db.db.First(&stateInfo).Error

	if err != nil {
		return time.Time{}, err
	}

	return stateInfo.LastSynced, nil
}

func (db *Database) SetLastSynced(lastSynced time.Time) error {
	stateInfo := StateInfo{
		ID:         1,
		LastSynced: lastSynced,
	}

	return db.db.Save(&stateInfo).Error
}

var DB = NewDatabase()

func (db *Database) Create(data interface{}) error {
	return db.db.Create(data).Error
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

func (e *UserEntitlement) Scan(value interface{}) error {
	var s uint16
	err := value.(sql.Scanner).Scan(&s)
	if err != nil {
		return err
	}

	*e = UserEntitlement(s)

	return nil
}

func (e *UserEntitlement) Value() (driver.Value, error) {
	return uint16(*e), nil
}
