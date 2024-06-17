package socket

import (
	"database/sql"
	"strings"
	"time"
	"vaportrader/bot/src/services"
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

	entry := services.KV.Get(code + ":totp")

	if entry != nil {
		if entry.Expiry.Before(time.Now()) {
			ctx.Reply("This link has expired, please request a new one. You can do this using the `/link` interaction in Discord.")
		} else {
			user, err := services.DB.GetUserByID(entry.Value.(string))

			if err != nil {
				ctx.Reply("An error occured while fetching your account information. Please try again later.")
				return err
			}

			println(entry.Value.(string))

			username := strings.Split(entry.Value.(string), "*:*")[1]

			profile, err := services.API.GetUser(username)

			if err != nil {
				ctx.Reply("An error occured while fetching your account information. Please try again later. Note: This is case sensetive, so please make sure you spelled the username correctly on Discord.")
				return err
			}

			if profile.ID != ctx.Author {
				ctx.Reply("You are not the owner of the account '" + profile.IngameName + "', please make sure that you are the owner of the account you are trying to link.")
				return nil
			}

			user.WfmID = sql.NullString{String: ctx.Author, Valid: true}
			user.WfmUsername = sql.NullString{String: username, Valid: true}
			user.Locale = sql.NullString{String: profile.Locale, Valid: true}
			user.PreferredPlatform = sql.NullString{String: profile.Platform, Valid: true}
			user.LastSeen = profile.LastSeen
			user.Awards = append(user.Awards, services.Award{
				BadgeId: 4,
				UserId:  user.ID,
			})

			err = services.DB.Save(user)

			if err != nil {
				ctx.Reply("An error occured while saving your account information. Please try again later.")
				return err
			}

			ctx.Reply("Congratulations, you have successfully linked your Warframe Market account to your Discord profile!")
		}
	} else {
		ctx.Reply("This link does not exist, please request a new one. You can do this using the `/link` interaction in Discord.")
	}

	return nil
}

func LinkCommandPermissions(s *services.SocketClient, ctx *CommandContext) (bool, string, error) {
	return true, "", nil
}
