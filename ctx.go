package disgoplus

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// Ctx is the context passed to every command, component, and modal handler.
type Ctx struct {
	// Client is the underlying disgo bot client.
	Client *bot.Client

	// GuildID is the guild the interaction occurred in; 0 for DMs.
	GuildID snowflake.ID

	// ChannelID is the channel the interaction occurred in.
	ChannelID snowflake.ID

	// Member is the guild member who triggered the interaction; nil for DMs.
	Member *discord.ResolvedMember

	// User is the user who triggered the interaction (always set).
	User discord.User

	// CallerName is the name of the matched command/component/modal.
	CallerName string

	// Caller is the matched Command (nil for components/modals).
	Caller *Command

	// CommandData holds slash-command option data (nil for components/modals).
	CommandData *discord.SlashCommandInteractionData

	// MessageComponentOptions holds custom-id slug params (nil for commands/modals).
	MessageComponentOptions map[string]string

	// ModalData holds modal submit data (nil for commands/components).
	ModalData *discord.ModalSubmitInteractionData

	// internal response state
	applicationID snowflake.ID
	token         string
	respondFn     events.InteractionResponderFunc

	remaining []Handler
}

// Next advances the middleware chain.
func (ctx *Ctx) Next() {
	if len(ctx.remaining) == 0 {
		return
	}

	h := ctx.remaining[0]
	ctx.remaining = ctx.remaining[1:]
	h.HandleCommand(ctx)
}

// CreateMessage responds to the interaction with a new message.
func (ctx *Ctx) CreateMessage(
	msg discord.MessageCreate,
	opts ...rest.RequestOpt,
) error {
	return ctx.respondFn(
		discord.InteractionResponseTypeCreateMessage,
		msg,
		opts...)
}

// DeferCreateMessage sends a deferred "thinking" response.
func (ctx *Ctx) DeferCreateMessage(
	ephemeral bool,
	opts ...rest.RequestOpt,
) error {
	var data discord.InteractionResponseData
	if ephemeral {
		data = discord.MessageCreate{Flags: discord.MessageFlagEphemeral}
	}

	return ctx.respondFn(
		discord.InteractionResponseTypeDeferredCreateMessage,
		data,
		opts...)
}

// UpdateMessage updates the message a component interaction is from.
func (ctx *Ctx) UpdateMessage(
	msg discord.MessageUpdate,
	opts ...rest.RequestOpt,
) error {
	return ctx.respondFn(
		discord.InteractionResponseTypeUpdateMessage,
		msg,
		opts...)
}

// DeferUpdateMessage sends a deferred update acknowledgement.
func (ctx *Ctx) DeferUpdateMessage(opts ...rest.RequestOpt) error {
	return ctx.respondFn(
		discord.InteractionResponseTypeDeferredUpdateMessage,
		nil,
		opts...)
}

// Modal responds to the interaction with a modal.
func (ctx *Ctx) Modal(
	modal discord.ModalCreate,
	opts ...rest.RequestOpt,
) error {
	return ctx.respondFn(discord.InteractionResponseTypeModal, modal, opts...)
}

// CreateFollowupMessage creates a followup message after a deferred response.
func (ctx *Ctx) CreateFollowupMessage(
	msg discord.MessageCreate,
	opts ...rest.RequestOpt,
) (*discord.Message, error) {
	return ctx.Client.Rest.CreateFollowupMessage(
		ctx.applicationID,
		ctx.token,
		msg,
		opts...)
}

// DeleteFollowupMessage deletes a followup message.
func (ctx *Ctx) DeleteFollowupMessage(
	messageID snowflake.ID,
	opts ...rest.RequestOpt,
) error {
	return ctx.Client.Rest.DeleteFollowupMessage(
		ctx.applicationID,
		ctx.token,
		messageID,
		opts...)
}

// ApplicationID returns the bot application ID for this interaction.
func (ctx *Ctx) ApplicationID() snowflake.ID {
	return ctx.applicationID
}

// Token returns the interaction token (valid 15 minutes).
func (ctx *Ctx) Token() string {
	return ctx.token
}

// newCommandCtx builds a Ctx from an application command event.
func newCommandCtx(
	e *events.ApplicationCommandInteractionCreate,
	data discord.SlashCommandInteractionData,
	cmd *Command,
	handlers []Handler,
) *Ctx {
	ctx := &Ctx{
		Client:        e.Client(),
		CallerName:    data.CommandName(),
		Caller:        cmd,
		CommandData:   &data,
		applicationID: e.ApplicationID(),
		token:         e.Token(),
		respondFn:     e.Respond,
		remaining:     handlers,
	}
	if gid := e.GuildID(); gid != nil {
		ctx.GuildID = *gid
	}

	ctx.ChannelID = e.Channel().ID()
	ctx.Member = e.Member()
	ctx.User = e.User()

	return ctx
}

// newComponentCtx builds a Ctx from a component interaction event.
func newComponentCtx(
	e *events.ComponentInteractionCreate,
	mc *MessageComponent,
	options map[string]string,
	handlers []Handler,
) *Ctx {
	ctx := &Ctx{
		Client:                  e.Client(),
		CallerName:              e.Data.CustomID(),
		MessageComponentOptions: options,
		applicationID:           e.ApplicationID(),
		token:                   e.Token(),
		respondFn:               e.Respond,
		remaining:               handlers,
	}
	_ = mc

	if gid := e.GuildID(); gid != nil {
		ctx.GuildID = *gid
	}

	ctx.ChannelID = e.Channel().ID()
	ctx.Member = e.Member()
	ctx.User = e.User()

	return ctx
}

// newModalCtx builds a Ctx from a modal submit event.
func newModalCtx(
	e *events.ModalSubmitInteractionCreate,
	m *Modal,
	options map[string]string,
	handlers []Handler,
) *Ctx {
	data := e.Data
	ctx := &Ctx{
		Client:                  e.Client(),
		CallerName:              data.CustomID,
		ModalData:               &data,
		MessageComponentOptions: options,
		applicationID:           e.ApplicationID(),
		token:                   e.Token(),
		respondFn:               e.Respond,
		remaining:               handlers,
	}
	_ = m

	if gid := e.GuildID(); gid != nil {
		ctx.GuildID = *gid
	}

	ctx.ChannelID = e.Channel().ID()
	ctx.Member = e.Member()
	ctx.User = e.User()

	return ctx
}
