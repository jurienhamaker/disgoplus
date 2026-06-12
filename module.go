package disgoplus

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// CommandRegistration pairs an application command create with an optional
// target guild. A zero GuildID means the command syncs globally.
type CommandRegistration struct {
	Create  discord.ApplicationCommandCreate
	GuildID snowflake.ID
}

// Global wraps a command for global registration.
func Global(cmd discord.ApplicationCommandCreate) CommandRegistration {
	return CommandRegistration{Create: cmd}
}

// InGuild wraps a command for registration in a specific guild.
func InGuild(guildID snowflake.ID, cmd discord.ApplicationCommandCreate) CommandRegistration {
	return CommandRegistration{Create: cmd, GuildID: guildID}
}

// RoutableModule is the single interface top-level slash command modules must implement.
type RoutableModule interface {
	// Register wires all slash commands, components, and modals into the router.
	Register(r handler.Router)
	// Commands returns the command registrations used for sync.
	Commands() []CommandRegistration
}

// RegisterCommandModules registers all modules with the bot's router.
func RegisterCommandModules(bot *Bot, modules []RoutableModule) {
	for _, m := range modules {
		m.Register(bot.Router)
	}
}

// SyncCommands aggregates module commands and submits them via disgo's
// handler.SyncCommands.
//
// When devOverride is non-zero, every command — regardless of its own
// GuildID — syncs to that single guild (dev-mode override). Otherwise
// commands are grouped: those with GuildID == 0 sync globally, the rest
// per their guild.
func SyncCommands(bot *Bot, modules []RoutableModule, devOverride snowflake.ID, opts ...rest.RequestOpt) error {
	if devOverride != 0 {
		var all []discord.ApplicationCommandCreate
		for _, m := range modules {
			for _, reg := range m.Commands() {
				all = append(all, reg.Create)
			}
		}
		return handler.SyncCommands(bot.Client(), all, []snowflake.ID{devOverride}, opts...)
	}

	var globals []discord.ApplicationCommandCreate
	perGuild := map[snowflake.ID][]discord.ApplicationCommandCreate{}
	for _, m := range modules {
		for _, reg := range m.Commands() {
			if reg.GuildID == 0 {
				globals = append(globals, reg.Create)
				continue
			}
			perGuild[reg.GuildID] = append(perGuild[reg.GuildID], reg.Create)
		}
	}

	if err := handler.SyncCommands(bot.Client(), globals, nil, opts...); err != nil {
		return err
	}

	for gid, cmds := range perGuild {
		if err := handler.SyncCommands(bot.Client(), cmds, []snowflake.ID{gid}, opts...); err != nil {
			return err
		}
	}

	return nil
}
