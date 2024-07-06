package socket

import (
	"strings"
	"vaportrader/src/services"
)

type SocketCommand struct {
	Name        string
	Description string
	Usage       string
	Category    string
	Cooldown    int
	Aliases     []string
	Handler     SocketCommandHandlerMethod
	Permissions SocketCommandPermissionsMethod
}

type CommandContext struct {
	Command   string
	message   *services.NewMessage
	Arguments []string
	Author    string
	User      *services.User
}

func BuildCommandContext(s *services.SocketClient, msg *services.NewMessage) *CommandContext {
	var user *services.User
	user, _ = services.DB.GetUserByWFMID(msg.MessageFrom)

	return &CommandContext{
		Command:   strings.Split(msg.RawMessage, " ")[0],
		message:   msg,
		Arguments: strings.Split(msg.RawMessage, " ")[1:],
		Author:    msg.MessageFrom,
		User:      user,
	}
}

func (c *CommandContext) GetArgument(index int) string {
	if index < len(c.Arguments) {
		return c.Arguments[index]
	}
	return ""
}

func (c *CommandContext) Reply(text string) (*services.MessageAcknowledgement, error) {
	return c.message.Reply(text)
}

func (c *CommandContext) Acknowledge() {
	c.message.Acknowledge()
}

type SocketCommandHandlerMethod func(s *services.SocketClient, ctx *CommandContext) error
type SocketCommandPermissionsMethod func(s *services.SocketClient, ctx *CommandContext) (bool, string, error)

var CMDHandler = SocketCommandHandler{
	Index: map[string]SocketCommand{},
}

type SocketCommandHandler struct {
	Index map[string]SocketCommand
}

func (c *SocketCommandHandler) Register(cmd SocketCommand) {
	c.Index[cmd.Name] = cmd
}

func (c *SocketCommandHandler) HandleCommand(s *services.SocketClient, msg *services.NewMessage) {
	words := strings.Split(msg.RawMessage, " ")
	msg.Acknowledge()

	if len(words) < 1 {
		return
	}

	cmdName := words[0]

	cmd, ok := c.Index[cmdName]

	if !ok {
		return
	}

	ctx := BuildCommandContext(s, msg)

	permitted, reason, err := cmd.Permissions(s, ctx)

	if err != nil {
		_, _ = msg.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.handler.errors.perms.failed", &map[string]interface{}{
			"Error": err.Error(),
		}))
		return
	}

	if !permitted {
		_, _ = msg.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.handler.errors.perms.unauthorized", &map[string]interface{}{
			"Reason": reason,
		}))
		return
	}

	err = cmd.Handler(s, ctx)

	if err != nil {
		_, _ = msg.Reply(services.LanguageManager.Get(&ctx.User.Locale.String, "commands.handler.errors.generic.failed", &map[string]interface{}{
			"Error": err.Error(),
		}))
		return
	}
}

func Load() {
	CMDHandler.Register(HelpCommand())
	CMDHandler.Register(LinkCommand())
}
