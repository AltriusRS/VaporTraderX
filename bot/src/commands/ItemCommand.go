package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func ItemCommand() Command {
	return Command{
		Name:         "market",
		Description:  "Get information about a specific item or set.",
		Usage:        "market [item: name] [set: name]",
		Category:     "Utility",
		Cooldown:     5 * time.Second,
		Handler:      ItemHandler,
		Permissions:  ItemPermissions,
		Autocomplete: ItemAutocomplete,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "item",
				Description:  "The name of the item to get information about.",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     false,
				Autocomplete: true,
			},
			{
				Name:         "set",
				Description:  "The name of the set to get information about.",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     false,
				Autocomplete: true,
			},
			{
				Name:        "platform",
				Description: "The platform to get information about.",
				Type:        discordgo.ApplicationCommandOptionString,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "PC",
						Value: "pc",
					},
					{
						Name:  "Xbox (One, Series X/S)",
						Value: "xbox",
					},
					{
						Name:  "Playstation 4/5",
						Value: "ps4",
					},
					{
						Name:  "Nintendo Switch",
						Value: "switch",
					},
				},
				Required: false,
			},
		},
	}
}

func ItemHandler(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {
	var item string = ""
	var set string = ""
	var platform string = "pc"

	if itemOption := ctx.Options["item"]; itemOption != nil {
		item = itemOption.StringValue()
	}

	if setOption := ctx.Options["set"]; setOption != nil {
		set = setOption.StringValue()

		if item != "" {
			return false, fmt.Errorf("You may not specify an item and a set at the same time.")
		}
	}

	if platformOption := ctx.Options["platform"]; platformOption != nil {
		platform = platformOption.StringValue()
	}

	if item == "" && set == "" {
		return false, fmt.Errorf("You must specify an item to get information about.")
	}

	if platform != "pc" {
		return false, fmt.Errorf("Platform not supported")
	}

	// Fetch item stats from the api

	// tx := services.DB.Inner.Find(&item, "", itemId)

	// services.API.GetItemStats(, platform)

	return true, nil
}

func ItemPermissions(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error) {
	return true, "", nil
}

func ItemAutocomplete(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {

	switch {
	case ctx.Options["item"].Focused:

	case ctx.Options["set"].Focused:

	case ctx.Options["platform"].Focused:

	}

	log.Printf("Autocomplete: %v", ctx.Options["item"].StringValue())

	if ctx.Options["item"].StringValue() == "" {
		return false, nil
	}

	return true, nil
}
