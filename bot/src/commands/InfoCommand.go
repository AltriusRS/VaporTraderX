package commands

import (
	"time"
	"vaportrader/bot/src/constants"

	"github.com/bwmarrin/discordgo"
)

func InfoCommand() Command {
	return Command{
		Name:        "info",
		Description: "Get some basic information about the bot.",
		Usage:       "info",
		Category:    "Utility",
		Cooldown:    5 * time.Second,
		Handler:     InfoHandler,
		Permissions: InfoPermissions,
		Options:     []*discordgo.ApplicationCommandOption{},
	}
}

func InfoHandler(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {
	err := s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Bot Information",
					Description: "Vapor Trader is a bot which allows you to get information about the market value of Warframe items, as well as set up alerts for new orders which match your criteria.",
					Color:       constants.ThemeColor,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Name",
							Value:  constants.BotName,
							Inline: true,
						},
						{
							Name:   "Version",
							Value:  constants.Version,
							Inline: true,
						},
						{
							Name:   "Author",
							Value:  constants.AuthorString,
							Inline: false,
						},
					},
					Footer: constants.Footer,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	return true, err
}

func InfoPermissions(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error) {
	return true, "", nil
}
