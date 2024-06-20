package commands

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"
	"vaportrader/bot/src/constants"
	"vaportrader/bot/src/services"

	"github.com/bwmarrin/discordgo"
	"go.mills.io/bitcask/v2"
)

type CommandActionMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx ActionContext) (bool, error)
type CommandHandlerMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error)
type CommandPermissionsMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error)
type CommandAutocompleteMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error)
type CommandModalMethod func(s *discordgo.Session, m *discordgo.InteractionCreate, ctx ModalContext) (bool, error)
type CommandRegisterMethod func() Command

var Commands = map[string]Command{}

type CommandContext struct {
	User    *services.User
	Options map[string]*discordgo.ApplicationCommandInteractionDataOption
}

type ModalContext struct {
	User    *services.User
	Options map[string]string
}

type ActionContext struct {
	User   *services.User
	Action *discordgo.MessageComponentInteractionData
}

type Command struct {
	Name         string
	Description  string
	Usage        string
	Category     string
	Cooldown     time.Duration
	Aliases      []string
	Handler      CommandHandlerMethod
	Permissions  CommandPermissionsMethod
	Autocomplete CommandAutocompleteMethod
	Action       CommandActionMethod
	Modal        CommandModalMethod
	Options      []*discordgo.ApplicationCommandOption
}

func (c *Command) Register(s *discordgo.Session) {

	s.ApplicationCommandCreate(s.State.User.ID, os.Getenv("TEST_GUILD"), &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
		Options:     c.Options,
	})
}

type CommandHandler struct {
	icom  map[string]Command
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

	switch m.Type {
	case discordgo.InteractionType(discordgo.InteractionApplicationCommand):
		c.HandleApplicationCommand(s, m)
	case discordgo.InteractionMessageComponent:
		c.HandleMessageComponent(s, m)
	case discordgo.InteractionModalSubmit:
		c.HandleModalSubmit(s, m)
	case discordgo.InteractionPing:
		c.HandlePing(s, m)
	case discordgo.InteractionApplicationCommandAutocomplete:
		c.HandleAutocomplete(s, m)
	default:
		log.Printf("Unknown interaction type: %v", m.Type)
	}
}

func (c *CommandHandler) HandleMessageComponent(s *discordgo.Session, m *discordgo.InteractionCreate) {
	log.Printf("Message component interaction: %v", m.Type)

	cmdData := m.MessageComponentData()

	var cmdName string = ""

	switch {
	case strings.HasPrefix(cmdData.CustomID, "link_account_wfm_"):
		cmdName = "link"
	case strings.HasPrefix(cmdData.CustomID, "unlink_account_wfm_"):
		cmdName = "unlink"
	default:
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
		return
	}

	cmd, ok := c.index[cmdName]

	if !ok {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
		return
	}

	var DUser *discordgo.User

	if m.User != nil {
		DUser = m.User
	} else {
		DUser = m.Member.User
	}

	user, err := services.DB.GetUserByID(DUser.ID)

	if err != nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error whilst executing command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		log.Fatalf("Error while fetching user: %v", err)
		return
	} else if user.ID == "" {
		user = &services.User{
			ID:                DUser.ID,
			Name:              DUser.GlobalName,
			Entitlements:      uint32(0),
			Locale:            sql.NullString{Valid: false},
			WfmID:             sql.NullString{Valid: false},
			PreferredPlatform: sql.NullString{Valid: false},
			FirstSeen:         time.Now(),
			LastSeen:          time.Now(),
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

	ctx := ActionContext{
		User:   user,
		Action: &cmdData,
	}

	if cmd.Action == nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
		return
	}

	success, err := cmd.Action(s, m, ctx)

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

func (c *CommandHandler) HandleModalSubmit(s *discordgo.Session, m *discordgo.InteractionCreate) {
	log.Printf("Modal submit interaction: %v", m.Type)
	cmdData := m.ModalSubmitData()

	var cmdName string = ""

	switch {
	case strings.HasPrefix(cmdData.CustomID, "modals_link_account_wfm_"):
		cmdName = "link"
	case strings.HasPrefix(cmdData.CustomID, "modals_unlink_account_wfm_"):
		cmdName = "unlink"
	default:
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
		return
	}

	cmd, ok := c.index[cmdName]

	var DUser *discordgo.User

	if m.User != nil {
		DUser = m.User
	} else {
		DUser = m.Member.User
	}

	var id string = "" + DUser.ID
	var username string = "" + DUser.GlobalName

	user, err := services.DB.GetUserByID(id)

	if err != nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error whilst executing command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		log.Fatalf("Error while fetching user: %v", err)
		return
	} else if user.ID == "" {
		user = &services.User{
			ID:                id,
			Name:              username,
			Entitlements:      uint32(0),
			Locale:            sql.NullString{Valid: false},
			WfmID:             sql.NullString{Valid: false},
			PreferredPlatform: sql.NullString{Valid: false},
			FirstSeen:         time.Now(),
			LastSeen:          time.Now(),
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

	ctx := ModalContext{
		User:    user,
		Options: map[string]string{},
	}

	for _, option := range cmdData.Components {

		if option.Type() == discordgo.ActionsRowComponent {
			comp := option.(*discordgo.ActionsRow)

			for _, component := range comp.Components {
				switch component.Type() {
				case discordgo.TextInputComponent:
					cp := component.(*discordgo.TextInput)
					ctx.Options[cp.CustomID] = cp.Value
				}
			}
		} else {
			continue
		}
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

	if cmd.Modal == nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No modal handler",
			},
		})
		return
	}

	success, err := cmd.Modal(s, m, ctx)

	if err != nil {
		log.Fatal(err.Error())
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

func (c *CommandHandler) HandlePing(s *discordgo.Session, m *discordgo.InteractionCreate) {
	log.Printf("Ping interaction: %v", m.Type)
}

func (c *CommandHandler) HandleAutocomplete(s *discordgo.Session, m *discordgo.InteractionCreate) {
	cmdData := m.ApplicationCommandData()

	cmd, ok := c.index[cmdData.Name]

	ctx := CommandContext{
		User:    nil,
		Options: map[string]*discordgo.ApplicationCommandInteractionDataOption{},
	}

	for _, option := range cmdData.Options {
		ctx.Options[option.Name] = option
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

	if cmd.Autocomplete == nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No autocomplete handler",
			},
		})
		return
	}

	success, err := cmd.Autocomplete(s, m, ctx)

	if err != nil {
		log.Fatal(err.Error())
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

func (c *CommandHandler) HandleApplicationCommand(s *discordgo.Session, m *discordgo.InteractionCreate) {
	cmdData := m.ApplicationCommandData()

	cmd, ok := c.index[cmdData.Name]

	var DUser *discordgo.User

	if m.User != nil {
		DUser = m.User
	} else {
		DUser = m.Member.User
	}

	var id string = "" + DUser.ID
	var username string = "" + DUser.GlobalName

	user, err := services.DB.GetUserByID(id)

	if err != nil {
		_ = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error whilst executing command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		log.Fatalf("Error while fetching user: %v", err)
		return
	} else if user.ID == "" {
		user = &services.User{
			ID:                id,
			Name:              username,
			Entitlements:      uint32(0),
			Locale:            sql.NullString{Valid: false},
			WfmID:             sql.NullString{Valid: false},
			PreferredPlatform: sql.NullString{Valid: false},
			FirstSeen:         time.Now(),
			LastSeen:          time.Now(),
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

	ctx := CommandContext{
		User:    user,
		Options: map[string]*discordgo.ApplicationCommandInteractionDataOption{},
	}

	for _, option := range cmdData.Options {
		ctx.Options[option.Name] = option
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
	icom:  map[string]Command{},
	index: Commands,
	kv:    bk,
}

func Load(s *discordgo.Session) {
	// commands, err := s.ApplicationCommands(s.State.User.ID, os.Getenv("TEST_GUILD"))

	// if err != nil {
	// 	log.Fatal(err.Error())
	// 	return
	// }

	// for _, command := range commands {
	// 	s.ApplicationCommandDelete(s.State.User.ID, os.Getenv("TEST_GUILD"), command.ID)
	// }

	CMDHandler.Register(s, InfoCommand)
	CMDHandler.Register(s, LinkCommand)
	CMDHandler.Register(s, ItemCommand)
}
