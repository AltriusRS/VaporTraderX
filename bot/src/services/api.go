package services

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type APIClient struct {
	client *http.Client
}

func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (a *APIClient) GetItems() ([]ApiItemPartial, error) {
	response, err := a.client.Get("https://api.warframe.market/v1/items")
	if err != nil {
		return nil, err
	}

	payload := ApiItemListResponse{}
	err = a.ReadJSON(response.Body, &payload)
	if err != nil {
		return nil, err
	}

	return payload.Payload.Items, nil
}

func (a *APIClient) GetItem(slug string) (*ApiItemGroup, error) {
	response, err := a.client.Get("https://api.warframe.market/v1/items/" + slug)
	if err != nil {
		return nil, err
	}

	payload := ApiItemResponse{}
	err = a.ReadJSON(response.Body, &payload)
	if err != nil {
		return nil, err
	}

	return &payload.Payload.Item, nil
}

func (a *APIClient) ReadBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (a *APIClient) ReadJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

var api = NewAPIClient()

// ItemTranslationFromAPI converts an API item translation to a database item translation
func ItemTranslationFromAPI(translation ApiItemTranslation) ItemTranslation {
	return ItemTranslation{
		Name:        translation.ItemName,
		Description: translation.Description,
		WikiLink:    translation.WikiLink,
		Icon:        translation.Icon,
		Thumbnail:   translation.Thumb,
		// Drop:        translation.Drop,
	}
}

// ItemFromApi converts an API item to a database item
func ItemFromApi(item ApiItem, isSet bool, rawSetId string) Item {
	var tags sql.NullString
	if item.Tags == nil {
		tags = sql.NullString{String: "", Valid: false}
	} else {
		tags = sql.NullString{String: strings.Join(item.Tags, ","), Valid: true}
	}

	setId := sql.NullString{String: rawSetId, Valid: isSet}
	icon := sql.NullString{String: item.Icon, Valid: item.Icon != ""}
	subIcon := sql.NullString{String: item.SubIcon, Valid: item.SubIcon != ""}
	thumbnail := sql.NullString{String: item.Thumb, Valid: item.Thumb != ""}
	iconFormat := sql.NullString{String: item.IconFormat, Valid: item.IconFormat != ""}
	numberForSet := sql.NullInt16{Int16: int16(item.QuantityForSet), Valid: item.QuantityForSet != 0}
	masteryLevel := sql.NullInt64{Int64: int64(item.MasteryLevel), Valid: item.MasteryLevel != 0}
	vaulted := sql.NullBool{Bool: item.Vaulted, Valid: !item.Vaulted}
	ducats := item.Ducats
	tradeTax := item.TradingTax

	return Item{
		ID:           item.ID,
		Slug:         item.URLName,
		IsSet:        isSet,
		PartOf:       setId,
		Icon:         icon,
		SubIcon:      subIcon,
		Thumbnail:    thumbnail,
		IconFormat:   iconFormat,
		NumberForSet: numberForSet,
		MasteryLevel: masteryLevel,
		Vaulted:      vaulted,
		Ducats:       ducats,
		TradeTax:     tradeTax,
		Tags:         tags,
		Translations: []ItemTranslation{ItemTranslationFromAPI(item.En), ItemTranslationFromAPI(item.Ru), ItemTranslationFromAPI(item.Ko), ItemTranslationFromAPI(item.Fr), ItemTranslationFromAPI(item.Sv), ItemTranslationFromAPI(item.De), ItemTranslationFromAPI(item.ZhHant), ItemTranslationFromAPI(item.ZhHans), ItemTranslationFromAPI(item.Pt), ItemTranslationFromAPI(item.Es), ItemTranslationFromAPI(item.Pl), ItemTranslationFromAPI(item.Cs), ItemTranslationFromAPI(item.Uk)},
	}
}
