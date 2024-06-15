package services

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

type ApiItemResponse struct {
	Payload ApiItemResponseItem `json:"payload"`
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
	Ducats         int                `json:"ducats"`
	TradingTax     int                `json:"trading_tax"`
	MasteryLevel   int                `json:"mastery_level"`
	Vaulted        bool               `json:"vaulted"`
	SubIcon        string             `json:"sub_icon"`
	URLName        string             `json:"url_name"`
	QuantityForSet int                `json:"quantity_for_set,omitempty"`
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
	ID          int    `json:"id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	WikiLink    string `json:"wiki_link"`
	Icon        string `json:"icon"`
	Thumb       string `json:"thumb"`
	Drop        []any  `json:"drop"`
}
