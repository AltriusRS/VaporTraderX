package services

import (
	"github.com/titanous/json5"
	"io"
	"log"
	"os"
)

type I18n struct {
	Languages map[string]Language
}

func NewI18n() *I18n {
	return &I18n{
		Languages: map[string]Language{},
	}
}

func (i *I18n) Load(path string) error {
	files, err := os.ReadDir(path)

	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name() == "contributing.md" {
			continue
		}

		handle, err := os.Open(path + "/" + file.Name())

		if err != nil {
			return err
		}

		data, err := io.ReadAll(handle)

		if err != nil {
			return err
		}

		j5 := &RawLanguage{}

		err = json5.Unmarshal(data, j5)

		if err != nil {
			return err
		}

		final := j5.Finalize()

		i.Languages[final.ISO] = final
	}

	return nil
}

type RawLanguage map[string]interface{}

type Param struct {
	Start int
	End   int
	Name  string
}

type Snippet struct {
	RawText string
	Params  []Param
}

type Language struct {
	ISO          string
	Name         string
	Maintainer   string
	Bindings     map[string]Snippet
	Measurements LanguageMeasurements
	Time         LanguageTime
}

type LanguageMeasurements struct {
	Meter      LanguageMeasurement
	Centimeter LanguageMeasurement
	Millimeter LanguageMeasurement
}

type LanguageTime struct {
	Second        string
	SecondsPlural string
	Minute        string
	MinutesPlural string
	Hour          string
	HoursPlural   string
	Day           string
	DaysPlural    string
	Week          string
	WeeksPlural   string
	Month         string
	MonthsPlural  string
	Year          string
	YearsPlural   string
}

// LanguageMeasurement is a measurement that can be used to convert between different units.
type LanguageMeasurement struct {
	// The name of the measurement.
	Name string

	// The multiplier relative to the base unit (meters).
	Multiplier float64
}

func SnippetFromString(s string) Snippet {
	RawText := s
	var Params []Param

	for index, char := range RawText {
		if char == '%' {
			start := index
			end := index + 1
			for end < len(RawText) && RawText[end] != '%' {
				end++
			}
			name := RawText[start+1 : end]
			Params = append(Params, Param{
				Start: start,
				End:   end,
				Name:  name,
			})
		}
	}

	return Snippet{
		RawText: RawText,
		Params:  Params,
	}
}

func (l RawLanguage) Finalize() Language {

	language := Language{
		ISO:        l["meta.iso"].(string),
		Name:       l["name"].(string),
		Maintainer: l["maintainer"].(string),
		Bindings:   map[string]Snippet{},
		Measurements: LanguageMeasurements{
			Meter: LanguageMeasurement{
				Name:       l["units.distance.name.meter"].(string),
				Multiplier: l["units.distance.multiplier.meter"].(float64),
			},
			Centimeter: LanguageMeasurement{
				Name:       l["units.distance.name.centimeter"].(string),
				Multiplier: l["units.distance.multiplier.centimeter"].(float64),
			},
			Millimeter: LanguageMeasurement{
				Name:       l["units.distance.name.millimeter"].(string),
				Multiplier: l["units.distance.multiplier.millimeter"].(float64),
			},
		},
		Time: LanguageTime{
			Second:        l["units.time.second"].(string),
			SecondsPlural: l["units.time.seconds"].(string),
			Minute:        l["units.time.minute"].(string),
			MinutesPlural: l["units.time.minutes"].(string),
			Hour:          l["units.time.hour"].(string),
			HoursPlural:   l["units.time.hours"].(string),
			Day:           l["units.time.day"].(string),
			DaysPlural:    l["units.time.days"].(string),
			Week:          l["units.time.week"].(string),
			WeeksPlural:   l["units.time.weeks"].(string),
			Month:         l["units.time.month"].(string),
			MonthsPlural:  l["units.time.months"].(string),
			Year:          l["units.time.year"].(string),
			YearsPlural:   l["units.time.years"].(string),
		},
	}

	for name, value := range l {
		switch name {
		case "meta.iso":
		case "meta.name":
		case "meta.maintainer":
		case "units.distance.name.meter":
		case "units.distance.multiplier.meter":
		case "units.distance.name.centimeter":
		case "units.distance.multiplier.centimeter":
		case "units.distance.name.millimeter":
		case "units.distance.multiplier.millimeter":
		case "units.time.second":
		case "units.time.seconds":
		case "units.time.minute":
		case "units.time.minutes":
		case "units.time.hour":
		case "units.time.hours":
		case "units.time.day":
		case "units.time.days":
		case "units.time.week":
		case "units.time.weeks":
		case "units.time.month":
		case "units.time.months":
		case "units.time.year":
		case "units.time.years":
			continue
		default:
			switch v := value.(type) {
			case string:
				language.Bindings[name] = SnippetFromString(v)
			default:
				log.Printf("Value for key is not valid - %s", name)
			}
		}

	}

	return language
}

// Get Returns the value of the given key in the language, or an empty string if the key is not found.
func (i *I18n) Get(iso *string, key string, params *map[string]interface{}) string {
	// Check if the language is present (or select the default)
	var lang Language
	if iso != nil {
		if l, ok := i.Languages[*iso]; ok {
			lang = l
		} else {
			lang = i.Languages["en-US"]
		}
	} else {
		lang = i.Languages["en-US"]
	}

	snippet, ok := lang.Bindings[key]

	if !ok {
		return ""
	}

	result := snippet.RawText

	if params == nil {
		params = &map[string]interface{}{}
	}

	// Replace params with values where possible or leave as is if value is not provided.
	for _, param := range snippet.Params {
		if (*params)[param.Name] == nil {
			continue
		} else {
			result = result[:param.Start] + (*params)[param.Name].(string) + result[param.End:]
		}
	}

	return result
}

var LanguageManager *I18n

func InitI18n() {
	LanguageManager = NewI18n()
	err := LanguageManager.Load("i18n")

	if err != nil {
		log.Fatal(err)
	}
}
