package socket

import (
	"database/sql"
	"time"
	"vaportrader/src/commands"
	"vaportrader/src/services"
)

func LinkCommand() SocketCommand {
	return SocketCommand{
		Name:        services.LanguageManager.Get(nil, "commands.wfm.link.name", nil),
		Description: services.LanguageManager.Get(nil, "commands.wfm.link.description", nil),
		Usage:       services.LanguageManager.Get(nil, "commands.wfm.link.usage", nil),
		Category:    "General",
		Cooldown:    5,
		Aliases:     []string{},
		Handler:     LinkCommandHandler,
		Permissions: LinkCommandPermissions,
	}
}

func LinkCommandHandler(s *services.SocketClient, ctx *CommandContext) error {
	if len(ctx.Arguments) < 1 {
		_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.invalid_code", nil))
		return nil
	}

	code := ctx.GetArgument(0)

	if code == "" {
		_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.invalid_code", nil))
		return nil
	}

	codeEntry := services.KV.Get(code + ":totp")

	rawEntry := services.KV.Get(codeEntry.Value.(string))

	if rawEntry != nil {
		if rawEntry.Expiry.Before(time.Now()) {
			_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.expired_code", &map[string]interface{}{
				"CommandName": services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.name", nil),
			}))
		} else {
			entry := rawEntry.Value.(commands.AccountLinkStatus)

			user, err := services.DB.GetUserByID(entry.ID)

			if err != nil {
				_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.error", nil))
				return err
			}

			if entry.Profile.ID != ctx.Author {
				_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.not_owner", &map[string]interface{}{
					"AccountName": entry.Profile.IngameName,
				}))
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
				_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.error", nil))

				//_, _ = s.Session.InteractionResponseEdit(entry.Interaction, &discordgo.WebhookEdit{
				//	Embeds: &[]*discordgo.MessageEmbed{
				//		{
				//			Title:       "An error occured while saving your account information.",
				//			Description: "Please try again later.",
				//			Color:       constants.ThemeColor,
				//			Fields: []*discordgo.MessageEmbedField{
				//				{
				//					Name:   "Error",
				//					Value:  err.Error(),
				//					Inline: true,
				//				},
				//			},
				//		},
				//	},
				//})

				return err
			}

			_, _ = ctx.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.wfm.link.dialog.success", &map[string]interface{}{
				"UserName": user.WfmUsername.String,
			}))

			services.KV.Delete(entry.Code + ":totp")
			services.KV.Delete(codeEntry.Value.(string))
		}
	} else {
		_, _ = ctx.Reply("This link does not exist, please request a new one. You can do this using the `/link` interaction in Discord.")
	}

	return nil
}

func LinkCommandPermissions(s *services.SocketClient, ctx *CommandContext) (bool, string, error) {
	return true, "", nil
}
