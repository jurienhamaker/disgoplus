package disgoplus

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

// RoutableModule is the single interface top-level slash command modules must implement.
type RoutableModule interface {
	// Register wires all slash commands, components, and modals into the router.
	Register(r handler.Router)
	// Commands returns Discord API definitions used for command sync.
	Commands() []discord.ApplicationCommandCreate
}

// RegisterCommandModules registers all modules with the bot's router.
func RegisterCommandModules(bot *Bot, modules []RoutableModule) {
	for _, m := range modules {
		m.Register(bot.Router)
	}
}

// SyncCommands submits all module commands to Discord.
// Pass a non-zero guildID to scope commands to a single guild (dev mode).
func SyncCommands(bot *Bot, modules []RoutableModule, guildID snowflake.ID) error {
	var cmds []discord.ApplicationCommandCreate
	for _, m := range modules {
		cmds = append(cmds, m.Commands()...)
	}

	appID := bot.ApplicationID()

	if guildID != 0 {
		_, err := bot.Client().Rest.SetGuildCommands(appID, guildID, cmds)
		return err
	}

	_, err := bot.Client().Rest.SetGlobalCommands(appID, cmds)

	return err
}
