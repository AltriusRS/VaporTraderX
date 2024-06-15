package commands

import (
	"database/sql"
	"log"
	"os"
	"time"
	"vaportrader/bot/src/constants"
	"vaportrader/bot/src/services"

	"github.com/bwmarrin/discordgo"
	"go.mills.io/bitcask/v2"
	"gorm.io/gorm"
)

type CommandHandlerMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error)
type CommandPermissionsMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error)
type CommandRegisterMethod func() Command

var Commands = map[string]Command{}

type CommandContext struct {
	User *services.User
}

type Command struct {
	Name        string
	Description string
	Usage       string
	Category    string
	Cooldown    time.Duration
	Aliases     []string
	Handler     CommandHandlerMethod
	Permissions CommandPermissionsMethod
	Options     []*discordgo.ApplicationCommandOption
}

func (c *Command) Register(s *discordgo.Session) {
	s.ApplicationCommandCreate(s.State.User.ID, os.Getenv("TEST_GUILD"), &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
		Options:     c.Options,
	})
}

type CommandHandler struct {
	index map[string]Command
	kv    *bitcask.Bitcask
}

func (c *CommandHandler) Register(s *discordgo.Session, cmd CommandRegisterMethod) {
	command := cmd()

	log.Printf("Registering command %v", command.Name)

	c.index[command.Name] = command
	command.Register(s)
}

func (c *CommandHandler) HandleCommand(s *discordgo.Session, m *discordgo.InteractionCreate) {
	cmdData := m.ApplicationCommandData()

	cmd, ok := c.index[cmdData.Name]

	var DUser *discordgo.User

	if m.User != nil {
		DUser = m.User
	} else {
		DUser = m.Member.User
	}

	user, err := services.DB.GetUserByID(DUser.ID)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user = &services.User{
				ID:                DUser.ID,
				Name:              DUser.GlobalName,
				Entitlements:      services.UserEntitlement(0),
				Locale:            sql.NullString{Valid: false},
				WFMID:             sql.NullString{Valid: false},
				PreferredPlatform: sql.NullString{Valid: false},
				FirstSeen:         time.Now(),
				LastSeen:          time.Now(),
				UpdatedAt:         time.Now(),
			}

			err = services.DB.Create(user)
			if err != nil {
				_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error whilst executing command",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				println(err)
				return
			}
		}
	}

	ctx := CommandContext{
		User: user,
	}

	if !ok {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
		return
	}

	permitted, reason, err := cmd.Permissions(s, m, ctx)

	if err != nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Error Checking Permissions",
						Description: "An error was encountered while we tried to verify your access to this command.",
						Color:       constants.ThemeColor,
						Footer:      constants.Footer,
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if !permitted {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Not Authorized",
						Description: "You are unable to use this command",
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Reason",
								Value:  reason,
								Inline: false,
							},
						},
						Color:  constants.ThemeColor,
						Footer: constants.Footer,
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	success, err := cmd.Handler(s, m, ctx)

	if err != nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error running command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if !success {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Command failed", Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

}

var bk, _ = bitcask.Open("vp.bk")

var CMDHandler = &CommandHandler{
	index: Commands,
	kv:    bk,
}

func Load(s *discordgo.Session) {
	CMDHandler.Register(s, InfoCommand)
	CMDHandler.Register(s, LinkCommand)
}
