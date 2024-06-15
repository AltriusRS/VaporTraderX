package socketCommands

import (
	"fmt"
	"strings"
	"vaportrader/bot/src/services"
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

type SocketCommandHandlerMethod func(s *services.SocketClient, msg *services.SocketPrivateMessage) error
type SocketCommandPermissionsMethod func(s *services.SocketClient, msg *services.SocketPrivateMessage) (bool, string, error)

var SocketCommands = map[string]SocketCommand{}

type SocketCommandHandler struct {
	index map[string]SocketCommand
}

func (c *SocketCommandHandler) Register(s *services.SocketClient, cmd SocketCommand) {
	c.index[cmd.Name] = cmd
}

func (c *SocketCommandHandler) HandleCommand(s *services.SocketClient, msg *services.SocketPrivateMessage) {
	words := strings.Split(msg.Text, " ")

	if len(words) < 1 {
		return
	}

	cmdName := words[0]

	cmd, ok := c.index[cmdName]

	if !ok {
		return
	}

	permitted, reason, err := cmd.Permissions(s, msg)

	if err != nil {
		msg.Reply(fmt.Sprintf("An error occured while checking permissions for this command.\nError: '%s'", err.Error()))
		return
	}

	if !permitted {
		msg.Reply(fmt.Sprintf("You do not have permission to use this command.\nReason: '%s'", reason))
		return
	}

	err = cmd.Handler(s, msg)

	if err != nil {
		msg.Reply(fmt.Sprintf("An error occured while executing this command.\nError: '%s'", err.Error()))
		return
	}
}
