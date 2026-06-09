package disgoplus

import "github.com/disgoorg/disgo/discord"

// Handler processes a command, component, or modal interaction.
type Handler interface {
	HandleCommand(ctx *Ctx)
}

// HandlerFunc wraps a plain function to implement Handler.
type HandlerFunc func(ctx *Ctx)

func (f HandlerFunc) HandleCommand(ctx *Ctx) { f(ctx) }

// Command represents a slash command or sub-command.
type Command struct {
	Name        string
	Description string
	// Options are the leaf-level ApplicationCommandOptions (not sub-commands).
	// Use SubCommands for nested commands.
	Options     []discord.ApplicationCommandOption
	Type        discord.ApplicationCommandType
	Handler     Handler
	GuildID     string
	Middlewares []Handler
	SubCommands *Router
}

// toSlashCommandCreate converts a top-level Command to a SlashCommandCreate suitable
// for bulk-overwrite registration. GuildID on the result is empty; the sync logic
// controls which guild (if any) commands are sent to.
func (cmd *Command) toSlashCommandCreate() discord.SlashCommandCreate {
	create := discord.SlashCommandCreate{
		Name:        cmd.Name,
		Description: cmd.Description,
	}
	if cmd.SubCommands != nil && cmd.SubCommands.Count() > 0 {
		for _, sub := range cmd.SubCommands.list() {
			create.Options = append(create.Options, sub.toApplicationCommandOption())
		}
	} else {
		create.Options = cmd.Options
	}
	return create
}

// toApplicationCommandOption converts a non-top-level Command to the appropriate
// option type (sub-command or sub-command group).
func (cmd *Command) toApplicationCommandOption() discord.ApplicationCommandOption {
	if cmd.SubCommands != nil && cmd.SubCommands.Count() > 0 {
		group := discord.ApplicationCommandOptionSubCommandGroup{
			Name:        cmd.Name,
			Description: cmd.Description,
		}
		for _, sub := range cmd.SubCommands.list() {
			if sc, ok := sub.toApplicationCommandOption().(discord.ApplicationCommandOptionSubCommand); ok {
				group.Options = append(group.Options, sc)
			}
		}
		return group
	}
	return discord.ApplicationCommandOptionSubCommand{
		Name:        cmd.Name,
		Description: cmd.Description,
		Options:     cmd.Options,
	}
}

// MessageComponent represents a message component (button, select menu) handler.
type MessageComponent struct {
	// CustomID supports slug routing: "LEADERBOARD/:page" matches "LEADERBOARD/3"
	// and stores {"page": "3"} in Ctx.MessageComponentOptions.
	CustomID    string
	Handler     Handler
	Middlewares []Handler
}

func (mc MessageComponent) try(customID string) (map[string]string, bool) {
	return trySlug(mc.CustomID, customID)
}

// Modal represents a modal submit handler.
type Modal struct {
	// CustomID supports slug routing: "RESET/:userID" matches "RESET/123".
	CustomID    string
	Handler     Handler
	Middlewares []Handler
}

func (m Modal) try(customID string) (map[string]string, bool) {
	return trySlug(m.CustomID, customID)
}
