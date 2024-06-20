package services

import "time"

type ApiItemListResponse struct {
	Payload ApiItemList `json:"payload"`
}

type ApiItemList struct {
	Items []ApiItemPartial `json:"items"`
}

type ApiItemPartial struct {
	ID        string `json:"id"`
	Slug      string `json:"url_name"`
	Name      string `json:"item_name"`
	Thumbnail string `json:"thumb"`
}

type ApiCoreResponse[T any] struct {
	Payload T `json:"payload"`
}

type ApiItemResponseItem struct {
	Item ApiItemGroup `json:"item"`
}

type ApiItemGroup struct {
	ID         string    `json:"id"`
	ItemsInSet []ApiItem `json:"items_in_set"`
}

type ApiItem struct {
	Tags           []string           `json:"tags"`
	IconFormat     string             `json:"icon_format"`
	SetRoot        bool               `json:"set_root"`
	Ducats         uint32             `json:"ducats"`
	TradingTax     uint32             `json:"trading_tax"`
	MasteryLevel   uint8              `json:"mastery_level"`
	Vaulted        bool               `json:"vaulted"`
	SubIcon        string             `json:"sub_icon"`
	URLName        string             `json:"url_name"`
	QuantityForSet uint8              `json:"quantity_for_set,omitempty"`
	Icon           string             `json:"icon"`
	Thumb          string             `json:"thumb"`
	ID             string             `json:"id"`
	En             ApiItemTranslation `json:"en"`
	Ru             ApiItemTranslation `json:"ru"`
	Ko             ApiItemTranslation `json:"ko"`
	Fr             ApiItemTranslation `json:"fr"`
	Sv             ApiItemTranslation `json:"sv"`
	De             ApiItemTranslation `json:"de"`
	ZhHant         ApiItemTranslation `json:"zh-hant"`
	ZhHans         ApiItemTranslation `json:"zh-hans"`
	Pt             ApiItemTranslation `json:"pt"`
	Es             ApiItemTranslation `json:"es"`
	Pl             ApiItemTranslation `json:"pl"`
	Cs             ApiItemTranslation `json:"cs"`
	Uk             ApiItemTranslation `json:"uk"`
}

type ApiItemTranslation struct {
	ID          int32  `json:"id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	WikiLink    string `json:"wiki_link"`
	Icon        string `json:"icon"`
	Thumb       string `json:"thumb"`
	Drop        []any  `json:"drop"`
}

type ApiAchievements struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Exposed     bool   `json:"exposed"`
	Type        string `json:"type"`
}

type ApiProfile struct {
	Reputation   int               `json:"reputation"`
	IngameName   string            `json:"ingame_name"`
	Achievements []ApiAchievements `json:"achievements"`
	Platform     string            `json:"platform"`
	Status       string            `json:"status"`
	Region       string            `json:"region"`
	Locale       string            `json:"locale"`
	OwnProfile   bool              `json:"own_profile"`
	ID           string            `json:"id"`
	LastSeen     time.Time         `json:"last_seen"`
	Banned       bool              `json:"banned"`
	About        string            `json:"about"`
	Avatar       *string           `json:"avatar"`
	Background   any               `json:"background"`
	AboutRaw     string            `json:"about_raw"`
}

type ApiProfilePayload struct {
	Profile ApiProfile `json:"profile"`
}
