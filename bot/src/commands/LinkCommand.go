package commands

import (
	"time"
	"vaportrader/bot/src/constants"
	"vaportrader/bot/src/services"

	"github.com/bwmarrin/discordgo"
)

func LinkCommand() Command {
	return Command{
		Name:        "link",
		Description: "Link your Warframe Market account to the bot",
		Usage:       "link",
		Category:    "Utility",
		Cooldown:    5 * time.Second,
		Handler:     LinkHandler,
		Permissions: LinkPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "username",
				Description: "The username you of your Warframe Market account",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func LinkHandler(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {

	println("Fetching existing codes...")
	entry := services.KV.Get(ctx.User.ID + ":totp")
	println("Fetched existing codes...")

	// The user has a code being stored in the KV
	if entry != nil {
		// resend the existing code
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "You already have an open link request. Here's your code!",
						Color: constants.ThemeColor,
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Code",
								Value:  entry.Value.(string),
								Inline: true,
							},
							{
								Name: "Expires",
								// Show the expiry time as a relative time
								Value: entry.Expiry.Format("in 2 minutes"),
							},
						},
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})

		return true, nil
	}

	println("No existing code found, generating a new one...")

	code, err := services.GenerateOTP(ctx.User.ID)

	println("Generated code: " + code)

	if err != nil {
		return false, err
	}

	entry = services.KV.Set(ctx.User.ID+":totp", code, time.Minute*15)

	_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Started a new link request. Here's your code!",
					Description: "Send this code to the [Vapor Trader](https://warframe.market/profile/VaporTrader) account as a private message. This will automatically link your Discord, and Warframe Market accounts",
					Color:       constants.ThemeColor,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Code",
							Value:  code,
							Inline: true,
						},
						{
							Name: "Expires",
							// Show the expiry time as a relative time
							Value:  entry.Expiry.Format("in 2 minutes"),
							Inline: true,
						},
						{
							Name:  "Usage",
							Value: "```\n!link " + code + "\n```",
						},
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	return true, nil
}

func LinkPermissions(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error) {

	if !ctx.User.WFMID.Valid {
		return true, "", nil
	}

	if ctx.User.WFMID.Valid {
		return false, "You already have a Warframe Market account linked to this bot", nil
	}

	return false, "This error shouldn't happen, as it indicates a corrupted database.", nil
}
