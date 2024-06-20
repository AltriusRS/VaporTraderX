package services

// Get all active price alerts
func GetActivePriceAlerts() ([]*Alert, error) {
	var alerts []*Alert

	err := DB.Inner.Where("active = ?", true).Find(&alerts).Error

	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// Get all price alerts for a given item
func GetActivePriceAlertsForItem(itemId string) ([]*Alert, error) {
	var alerts []*Alert

	err := DB.Inner.Where("item_id = ? AND active = ?", itemId, true).Find(&alerts).Error

	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// Get all price alerts for a given user
func GetActivePriceAlertsForUser(userId string) ([]*Alert, error) {
	var alerts []*Alert

	err := DB.Inner.Where("user_id = ?", userId).Find(&alerts).Error

	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// Add a new price alert
func AddPriceAlert(alert *Alert) error {
	return DB.Inner.Create(alert).Error
}

// Update a price alert
func UpdatePriceAlert(alert *Alert) error {
	return DB.Inner.Save(alert).Error
}

// Delete a price alert
func DeletePriceAlert(alert *Alert) error {
	return DB.Inner.Delete(alert).Error
}
