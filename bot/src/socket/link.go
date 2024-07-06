package socket

import (
	"database/sql"
	"time"
	"vaportrader/src/commands"
	"vaportrader/src/constants"
	"vaportrader/src/services"

	"github.com/bwmarrin/discordgo"
)

func LinkCommand() SocketCommand {
	return SocketCommand{
		Name:        "link",
		Description: "Used to link your Warframe Market account to your Discord account.",
		Usage:       "link [code]",
		Category:    "General",
		Cooldown:    5,
		Aliases:     []string{},
		Handler:     LinkCommandHandler,
		Permissions: LinkCommandPermissions,
	}
}

func LinkCommandHandler(s *services.SocketClient, ctx *CommandContext) error {
	if len(ctx.Arguments) < 1 {
		ctx.Reply("Please provide a valid code.")
		return nil
	}

	code := ctx.GetArgument(0)

	if code == "" {
		ctx.Reply("Please provide a valid code.")
		return nil
	}

	codeEntry := services.KV.Get(code + ":totp")

	rawEntry := services.KV.Get(codeEntry.Value.(string))

	if rawEntry != nil {
		if rawEntry.Expiry.Before(time.Now()) {
			ctx.Reply("This link has expired, please request a new one. You can do this using the `/link` interaction in Discord.")
		} else {
			entry := rawEntry.Value.(commands.AccountLinkStatus)

			user, err := services.DB.GetUserByID(entry.ID)

			if err != nil {
				ctx.Reply("An error occured while fetching your account information. Please try again later.")
				return err
			}

			if entry.Profile.ID != ctx.Author {
				ctx.Reply("You are not the owner of the account '" + entry.Profile.IngameName + "', please make sure that you are the owner of the account you are trying to link.")
				return nil
			}

			user.WfmID = sql.NullString{String: ctx.Author, Valid: true}
			user.WfmUsername = sql.NullString{String: entry.Profile.IngameName, Valid: true}
			user.Locale = sql.NullString{String: entry.Profile.Locale, Valid: true}
			user.PreferredPlatform = sql.NullString{String: entry.Profile.Platform, Valid: true}
			user.LastSeen = entry.Profile.LastSeen
			user.Awards = append(user.Awards, services.Award{
				BadgeId: 4,
				UserId:  user.ID,
			})

			err = services.DB.Save(user)

			if err != nil {
				ctx.Reply("An error occured while saving your account information. Please try again later.")

				s.Session.InteractionResponseEdit(entry.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{
						{
							Title:       "An error occured while saving your account information.",
							Description: "Please try again later.",
							Color:       constants.ThemeColor,
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:   "Error",
									Value:  err.Error(),
									Inline: true,
								},
							},
						},
					},
				})

				return err
			}

			ctx.Reply("Congratulations, " + entry.Interaction.ID + ", you have successfully linked your Warframe Market account to your Discord profile!")

			// var banText string = "No"

			// if entry.Profile.Banned {
			// 	banText = "Yes"
			// }

			// thumbnail := &discordgo.MessageEmbedThumbnail{}

			// if entry.Profile.Avatar != nil {
			// 	thumbnail = &discordgo.MessageEmbedThumbnail{
			// 		URL: "https://warframe.market/static/assets/" + *entry.Profile.Avatar,
			// 	}
			// }

			// channel, err := s.Session.UserChannelCreate(entry.ID)

			// if err != nil {
			// 	return err
			// }

			// s.Session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
			// 	Embeds: []*discordgo.MessageEmbed{
			// 		{
			// 			Title:       "Congratulations!",
			// 			Description: "You have successfully linked your Warframe Market account to your Discord profile!",
			// 			Color:       constants.ThemeColor,
			// 			Fields: []*discordgo.MessageEmbedField{
			// 				{
			// 					Name:   "Username",
			// 					Value:  entry.Profile.IngameName,
			// 					Inline: true,
			// 				},
			// 				{
			// 					Name:   "Status",
			// 					Value:  fmt.Sprintf("**%s**", entry.Profile.Status),
			// 					Inline: true,
			// 				},
			// 				{
			// 					Name:   "Banned",
			// 					Value:  banText,
			// 					Inline: true,
			// 				},
			// 			},
			// 			Thumbnail: thumbnail,
			// 		},
			// 	},
			// 	Flags: discordgo.MessageFlagsEphemeral,
			// })

			services.KV.Delete(entry.Code + ":totp")
			services.KV.Delete(codeEntry.Value.(string))
		}
	} else {
		ctx.Reply("This link does not exist, please request a new one. You can do this using the `/link` interaction in Discord.")
	}

	return nil
}

func LinkCommandPermissions(s *services.SocketClient, ctx *CommandContext) (bool, string, error) {
	return true, "", nil
}
