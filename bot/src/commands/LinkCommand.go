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
		Description: "Used to link your Warframe Market account to your Discord account.",
		Usage:       "link username: VaporTrader",
		Category:    "Utility",
		Cooldown:    5 * time.Second,
		Handler:     LinkHandler,
		Permissions: LinkPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "username",
				Description: "The username you of your Warframe Market account.",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}
}

func LinkHandler(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {

	code, err := services.GenerateOTP(ctx.User.ID)

	println("Generated code: " + code)

	if err != nil {
		return false, err
	}

	username := m.ApplicationCommandData().Options[0].StringValue()

	entry := services.KV.Set(code+":totp", ctx.User.ID+"*:*"+username, time.Minute*15)

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
							Value: "```\nlink " + code + "\n```",
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

	if !ctx.User.WfmID.Valid {
		return true, "", nil
	}

	if ctx.User.WfmID.Valid {
		return false, "You already have a Warframe Market account linked to this bot\nIf you wish to link another account, please use `/unlink` to invalidate the linked account.", nil
	}

	return false, "This error shouldn't happen, as it indicates a corrupted database.", nil
}
