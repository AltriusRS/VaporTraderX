package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

var Syncing bool = false

func Sync(deep bool) error {

	Syncing = true

	defer func() {
		Syncing = false
	}()

	apiItems, err := API.GetItems()

	if err != nil {
		return err
	}

	existingItems := make(map[string]bool)

	// If we are doing a deep sync, we need to sync __everything__
	// So, efficiency gains from skipping items that already exist
	// are lost
	if !deep {
		// Check if we have the items in our database already
		// If we do, we can skip them
		for _, apiItem := range apiItems {
			if existingItems[apiItem.ID] {
				continue
			}

			err := DB.Inner.Select("id").Where("id = ?", apiItem.ID).First(&Item{}).Error

			if err != nil {
				if err != gorm.ErrRecordNotFound {
					log.Println(err)
					continue
				} else {
					continue
				}
			}

			existingItems[apiItem.ID] = true
		}
	}

	// Sync the items from the API to the database
	for index, apiItem := range apiItems {
		if existingItems[apiItem.ID] {
			fmt.Printf("\x1b[33mSkipping item %d of %d - %s\x1b[0m\n", index+1, len(apiItems), apiItem.Name)
			continue
		}
		fmt.Printf("Syncing item %d of %d - %s\n", index+1, len(apiItems), apiItem.Name)
		manifest, err := API.GetItem(apiItem.Slug)

		if err != nil {
			return err
		}

		for _, apiItemInSet := range manifest.ItemsInSet {

			// If we are doing a deep sync, this condition will always be false
			if existingItems[apiItemInSet.ID] {
				fmt.Println("\x1b[33m ->", apiItemInSet.ID, "|", apiItemInSet.En.ItemName, "\x1b[0m")
				continue
			}

			highlight := "\x1b[32m"

			tx := DB.Inner.Select("id").Where("id = ?", apiItemInSet.ID).First(&Item{})

			// If there is an error which is not a record not found error, log it and continue
			// Otherwise, we can just call InsertItem to save the new item to the database
			if tx.Error != nil {
				if tx.Error == gorm.ErrRecordNotFound {

					highlight = "\x1b[31m"

					item := ItemFromApi(apiItemInSet, len(manifest.ItemsInSet) > 1, manifest.ID)

					err = DB.InsertItem(&item)

					if err != nil {
						log.Fatalf("Error inserting item %s: %s", item.ID, err)
					}
				} else {
					log.Println(tx.Error)
				}
			}

			// If this passes, we are doing a deep sync of an existing item
			if tx.Error == nil && deep {
				item := ItemFromApi(apiItemInSet, len(manifest.ItemsInSet) > 1, manifest.ID)
				err = DB.Inner.Save(&item).Error
				if err != nil {
					log.Fatalf("Error saving item %s: %s", item.ID, err)
				}
			}

			fmt.Println(highlight, " ->", apiItemInSet.ID, "|", apiItemInSet.En.ItemName, "\x1b[0m")

		}

		// Wait just long enough to not upset Cloudflare's anti-DDoS system
		time.Sleep(time.Millisecond * 330)
	}

	fmt.Println("Sync complete")

	err = DB.SetLastSynced(time.Now(), deep)

	if err != nil {
		return err
	}

	return nil
}
