package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func Sync() error {
	apiItems, err := api.GetItems()

	if err != nil {
		return err
	}

	for index, apiItem := range apiItems {
		fmt.Printf("Syncing item %d of %d - %s\n", index+1, len(apiItems), apiItem.Name)
		manifest, err := api.GetItem(apiItem.Slug)

		if err != nil {
			return err
		}

		for _, apiItemInSet := range manifest.ItemsInSet {
			highlight := "\x1b[32m"

			tx := DB.db.Select("id").Where("id = ?", apiItemInSet.ID).First(&Item{})

			if tx.Error != nil {
				if tx.Error == gorm.ErrRecordNotFound {

					highlight = "\x1b[31m"

					item := ItemFromApi(apiItemInSet, len(manifest.ItemsInSet) > 1, manifest.ID)

					err = DB.InsertItem(&item)

					if err != nil {
						log.Printf("Error inserting item: %s", err)
					}
				}
			}

			fmt.Println(highlight, " ->", apiItemInSet.ID, "|", apiItemInSet.En.ItemName, "\x1b[0m")

		}

		time.Sleep(time.Millisecond * 330)
	}

	fmt.Println("Sync complete")

	err = DB.SetLastSynced(time.Now())

	if err != nil {
		return err
	}

	return nil
}
