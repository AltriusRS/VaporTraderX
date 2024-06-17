package socket

import (
	"fmt"
	"vaportrader/bot/src/services"
)

func HelpCommand() SocketCommand {
	return SocketCommand{
		Name:        "help",
		Description: "Shows a list of actions or help for a specific action",
		Usage:       "help [action]",
		Category:    "General",
		Cooldown:    5,
		Aliases:     []string{"commands"},
		Handler:     HelpCommandHandler,
		Permissions: HelpCommandPermissions,
	}
}

func HelpCommandHandler(s *services.SocketClient, ctx *CommandContext) error {
	if len(ctx.Arguments) < 1 {

		var response string = "Hi there! This account is operated by an automated system.\nThese are all of the things I can do here:\n"

		for _, cmd := range CMDHandler.Index {
			response += fmt.Sprintf("\n- `%s` - %s", cmd.Name, cmd.Description)
		}

		response += "\n\nThis is all for now."

		ctx.Reply(response)
		return nil
	}

	commandName := ctx.GetArgument(0)

	if _, ok := CMDHandler.Index[commandName]; !ok {
		ctx.Reply(fmt.Sprintf("Action '%s' not found.", commandName))
		return nil
	}

	cmd := CMDHandler.Index[commandName]

	response := fmt.Sprintf("**%s**\n\n%s\n\n", cmd.Name, cmd.Description)

	for _, alias := range cmd.Aliases {
		response += fmt.Sprintf("Alias: `%s`\n", alias)
	}

	ctx.Reply(response)
	return nil
}

func HelpCommandPermissions(s *services.SocketClient, ctx *CommandContext) (bool, string, error) {
	return true, "", nil
}
