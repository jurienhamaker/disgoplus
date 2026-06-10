package disgoplus

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// ErrCommandNotExists is returned when a command lookup fails.
var ErrCommandNotExists = errCommandNotExists("command not exists")

type errCommandNotExists string

func (e errCommandNotExists) Error() string { return string(e) }

// Router stores registered commands, components, and modals, and dispatches
// incoming interactions. It implements bot.EventListener.
type Router struct {
	commands          map[string]*Command
	messageComponents map[string]*MessageComponent
	modals            map[string]*Modal
}

var _ bot.EventListener = (*Router)(nil)

func newRouter() *Router {
	return &Router{
		commands:          make(map[string]*Command),
		messageComponents: make(map[string]*MessageComponent),
		modals:            make(map[string]*Modal),
	}
}

// NewRouter creates a Router pre-populated with the given commands.
func NewRouter(cmds []*Command) *Router {
	r := newRouter()

	for _, cmd := range cmds {
		r.Register(cmd)
	}

	return r
}

// Register registers a command.
func (r *Router) Register(cmd *Command) {
	if _, ok := r.commands[cmd.Name]; !ok {
		r.commands[cmd.Name] = cmd
	}
}

// RegisterMessageComponent registers a message component handler.
func (r *Router) RegisterMessageComponent(mc *MessageComponent) {
	if _, ok := r.messageComponents[mc.CustomID]; !ok {
		r.messageComponents[mc.CustomID] = mc
	}
}

// RegisterModal registers a modal submit handler.
func (r *Router) RegisterModal(m *Modal) {
	if _, ok := r.modals[m.CustomID]; !ok {
		r.modals[m.CustomID] = m
	}
}

// Count returns the number of registered top-level commands.
func (r *Router) Count() int {
	if r == nil {
		return 0
	}

	return len(r.commands)
}

// list returns all registered commands as a slice.
func (r *Router) list() []*Command {
	if r == nil {
		return nil
	}

	cmds := make([]*Command, 0, len(r.commands))
	for _, c := range r.commands {
		cmds = append(cmds, c)
	}

	return cmds
}

// get returns the named command or nil.
func (r *Router) get(name string) *Command {
	if r == nil {
		return nil
	}

	return r.commands[name]
}

// Commands returns all top-level commands (for metrics/introspection).
func (r *Router) Commands() []*Command {
	return r.list()
}

// Sync bulk-overwrites application commands with Discord.
//
// If guildID is non-zero, commands scoped to that guild ID are overwritten for
// that guild; all other commands are written globally. If guildID is zero,
// all commands are written globally.
func (r *Router) Sync(
	client *bot.Client,
	appID snowflake.ID,
	guildID snowflake.ID,
) error {
	global := make([]discord.ApplicationCommandCreate, 0)
	perGuild := make(map[snowflake.ID][]discord.ApplicationCommandCreate)

	for _, cmd := range r.commands {
		create := cmd.toSlashCommandCreate()
		if cmd.GuildID != "" {
			id, err := snowflake.Parse(cmd.GuildID)
			if err != nil {
				return err
			}

			perGuild[id] = append(perGuild[id], create)

			continue
		}

		if guildID != 0 {
			perGuild[guildID] = append(perGuild[guildID], create)
		} else {
			global = append(global, create)
		}
	}

	if len(global) > 0 {
		if _, err := client.Rest.SetGlobalCommands(appID, global); err != nil {
			return err
		}
	}

	for gid, cmds := range perGuild {
		if _, err := client.Rest.SetGuildCommands(
			appID,
			gid,
			cmds,
		); err != nil {
			return err
		}
	}

	return nil
}

// OnEvent implements bot.EventListener.
func (r *Router) OnEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.ApplicationCommandInteractionCreate:
		r.dispatchCommand(e)
	case *events.ComponentInteractionCreate:
		r.dispatchComponent(e)
	case *events.ModalSubmitInteractionCreate:
		r.dispatchModal(e)
	}
}

func (r *Router) dispatchCommand(
	e *events.ApplicationCommandInteractionCreate,
) {
	if e.Data.Type() != discord.ApplicationCommandTypeSlash {
		return
	}

	data := e.Data.(discord.SlashCommandInteractionData)

	cmd := r.get(data.CommandName())
	if cmd == nil {
		return
	}

	handlers := make([]Handler, 0, len(cmd.Middlewares)+4)
	handlers = append(handlers, cmd.Middlewares...)
	finalCmd := cmd

	switch {
	case data.SubCommandGroupName != nil:
		groupCmd := cmd.SubCommands.get(*data.SubCommandGroupName)
		if groupCmd == nil {
			return
		}

		handlers = append(handlers, groupCmd.Middlewares...)
		if data.SubCommandName != nil {
			subCmd := groupCmd.SubCommands.get(*data.SubCommandName)
			if subCmd == nil {
				return
			}

			handlers = append(handlers, subCmd.Middlewares...)
			handlers = append(handlers, subCmd.Handler)
			finalCmd = subCmd
		}

	case data.SubCommandName != nil:
		subCmd := cmd.SubCommands.get(*data.SubCommandName)
		if subCmd == nil {
			return
		}

		handlers = append(handlers, subCmd.Middlewares...)
		handlers = append(handlers, subCmd.Handler)
		finalCmd = subCmd

	default:
		if cmd.Handler != nil {
			handlers = append(handlers, cmd.Handler)
		}
	}

	ctx := newCommandCtx(e, data, finalCmd, handlers)
	ctx.Next()
}

func (r *Router) getComponent(
	customID string,
) (*MessageComponent, map[string]string) {
	if mc, ok := r.messageComponents[customID]; ok {
		return mc, nil
	}

	for _, mc := range r.messageComponents {
		if params, ok := mc.try(customID); ok {
			return mc, params
		}
	}

	return nil, nil
}

func (r *Router) getModal(customID string) (*Modal, map[string]string) {
	if m, ok := r.modals[customID]; ok {
		return m, nil
	}

	for _, m := range r.modals {
		if params, ok := m.try(customID); ok {
			return m, params
		}
	}

	return nil, nil
}

func (r *Router) dispatchComponent(e *events.ComponentInteractionCreate) {
	customID := e.Data.CustomID()

	mc, params := r.getComponent(customID)
	if mc == nil {
		return
	}

	handlers := append(slicesClone(mc.Middlewares), mc.Handler)
	ctx := newComponentCtx(e, mc, params, handlers)
	ctx.Next()
}

func (r *Router) dispatchModal(e *events.ModalSubmitInteractionCreate) {
	customID := e.Data.CustomID

	m, params := r.getModal(customID)
	if m == nil {
		return
	}

	handlers := append(slicesClone(m.Middlewares), m.Handler)
	ctx := newModalCtx(e, m, params, handlers)
	ctx.Next()
}

// SetGuildCommands is a helper that overwrites commands for a single guild.
func SetGuildCommands(
	client *bot.Client,
	appID snowflake.ID,
	guildID snowflake.ID,
	cmds []discord.ApplicationCommandCreate,
	opts ...rest.RequestOpt,
) ([]discord.ApplicationCommand, error) {
	return client.Rest.SetGuildCommands(appID, guildID, cmds, opts...)
}

// SetGlobalCommands is a helper that overwrites global commands.
func SetGlobalCommands(
	client *bot.Client,
	appID snowflake.ID,
	cmds []discord.ApplicationCommandCreate,
	opts ...rest.RequestOpt,
) ([]discord.ApplicationCommand, error) {
	return client.Rest.SetGlobalCommands(appID, cmds, opts...)
}

func slicesClone[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}

	c := make(S, len(s))
	copy(c, s)

	return c
}
