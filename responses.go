package disgoplus

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
)

// Defer sends a deferred "thinking" response.
// Pass ephemeral=true to make the response visible only to the invoker.
func Defer(ctx *Ctx, ephemeral ...bool) error {
	ep := len(ephemeral) > 0 && ephemeral[0]
	return ctx.DeferCreateMessage(ep)
}

// Respond responds to the interaction with a visible message.
func Respond(
	ctx *Ctx,
	msg discord.MessageCreate,
	opts ...rest.RequestOpt,
) error {
	return ctx.CreateMessage(msg, opts...)
}

// FollowUp sends a followup message after a deferred response.
func FollowUp(
	ctx *Ctx,
	msg discord.MessageCreate,
	opts ...rest.RequestOpt,
) (*discord.Message, error) {
	return ctx.CreateFollowupMessage(msg, opts...)
}

// Update updates the message the component interaction is from.
func Update(
	ctx *Ctx,
	msg discord.MessageUpdate,
	opts ...rest.RequestOpt,
) error {
	return ctx.UpdateMessage(msg, opts...)
}

// ModalRespond responds to the interaction with a modal.
func ModalRespond(
	ctx *Ctx,
	modal discord.ModalCreate,
	opts ...rest.RequestOpt,
) error {
	return ctx.Modal(modal, opts...)
}

// ErrorResponse sends an ephemeral "something went wrong" embed response.
func ErrorResponse(ctx *Ctx, ephemeral ...bool) error {
	embed := discord.NewEmbed().
		WithColor(0xff0000).
		WithTitle("Something went wrong").
		WithDescription("Sorry, something broke along the way! My developer has been informed. Sorry for the inconvenience!")

	ep := len(ephemeral) > 0 && ephemeral[0]

	msg := discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	}
	if ep {
		msg.Flags = discord.MessageFlagEphemeral
	}

	return ctx.CreateMessage(msg)
}

// InteractionError sends an ephemeral error message. Use after Defer.
func InteractionError(ctx *Ctx, isFollowup bool) {
	const content = "Something went wrong, try again later."

	if isFollowup {
		_, _ = ctx.CreateFollowupMessage(discord.MessageCreate{
			Content: content,
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	_ = ctx.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
}

// MessageComponentError sends an ephemeral error update for a component interaction.
func MessageComponentError(ctx *Ctx) {
	empty := []discord.LayoutComponent{}
	emptyEmbeds := []discord.Embed{}
	content := "Something went wrong, try again later."
	_ = ctx.UpdateMessage(discord.MessageUpdate{
		Content:    &content,
		Embeds:     &emptyEmbeds,
		Components: &empty,
	})
}

// ForbiddenResponse sends an ephemeral "forbidden" embed response.
func ForbiddenResponse(ctx *Ctx) error {
	embed := discord.NewEmbed().
		WithColor(0xff0000).
		WithTitle("Forbidden").
		WithDescription("Sorry, you can't use this interaction!")

	return ctx.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
}
