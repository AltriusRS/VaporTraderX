package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	payload := ApiCoreResponse[ApiItemResponseItem]{}
	err = a.ReadJSON(response.Body, &payload)
	if err != nil {
		return nil, err
	}

	return &payload.Payload.Item, nil
}

func (a *APIClient) GetUser(username string) (*ApiProfile, error) {
	response, err := a.client.Get("https://api.warframe.market/v1/profile/" + username)
	if err != nil {
		return nil, err
	}

	payload := ApiCoreResponse[ApiProfilePayload]{}
	err = a.ReadJSON(response.Body, &payload)
	if err != nil {
		return nil, err
	}

	return &payload.Payload.Profile, nil
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

var API = NewAPIClient()

// ItemTranslationFromAPI converts an API item translation to a database item translation
func ItemTranslationFromAPI(translation ApiItemTranslation, id string, locale string) ItemTranslation {
	return ItemTranslation{
		ID:          fmt.Sprintf("%s-%s", id, locale),
		Name:        translation.ItemName,
		Locale:      locale,
		ItemId:      id,
		Description: translation.Description,
		WikiLink:    translation.WikiLink,
		Icon:        translation.Icon,
		Thumbnail:   translation.Thumb,
		// Drop:        translation.Drop,
	}
}

// ItemFromApi converts an API item to a database item
func ItemFromApi(item ApiItem, isSet bool, rawSetId string) Item {
	setId := sql.NullString{String: rawSetId, Valid: isSet}
	icon := sql.NullString{String: item.Icon, Valid: item.Icon != ""}
	subIcon := sql.NullString{String: item.SubIcon, Valid: item.SubIcon != ""}
	thumbnail := sql.NullString{String: item.Thumb, Valid: item.Thumb != ""}
	iconFormat := sql.NullString{String: item.IconFormat, Valid: item.IconFormat != ""}

	return Item{
		ID:           item.ID,
		Slug:         item.URLName,
		IsSet:        isSet,
		PartOf:       setId,
		Icon:         icon,
		SubIcon:      subIcon,
		Thumbnail:    thumbnail,
		IconFormat:   iconFormat,
		NumberForSet: item.QuantityForSet,
		MasteryLevel: item.MasteryLevel,
		Vaulted:      item.Vaulted,
		Ducats:       item.Ducats,
		TradeTax:     item.TradingTax,
		Tags:         item.Tags,
		Translations: []ItemTranslation{
			ItemTranslationFromAPI(item.En, item.ID, "en"),
			ItemTranslationFromAPI(item.Ru, item.ID, "ru"),
			ItemTranslationFromAPI(item.Ko, item.ID, "ko"),
			ItemTranslationFromAPI(item.Fr, item.ID, "fr"),
			ItemTranslationFromAPI(item.Sv, item.ID, "sv"),
			ItemTranslationFromAPI(item.De, item.ID, "de"),
			ItemTranslationFromAPI(item.ZhHant, item.ID, "zh-hant"),
			ItemTranslationFromAPI(item.ZhHans, item.ID, "zh-hans"),
			ItemTranslationFromAPI(item.Pt, item.ID, "pt"),
			ItemTranslationFromAPI(item.Es, item.ID, "es"),
			ItemTranslationFromAPI(item.Pl, item.ID, "pl"),
			ItemTranslationFromAPI(item.Cs, item.ID, "cs"),
			ItemTranslationFromAPI(item.Uk, item.ID, "uk"),
		},
	}
}
