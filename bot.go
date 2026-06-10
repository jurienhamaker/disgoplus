// Package disgoplus wraps github.com/disgoorg/disgo with a discordgoplus-compatible
// API surface: a Router for commands/components/modals, a Ctx handler context,
// and convenience response helpers.
package disgoplus

import (
	"context"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/snowflake/v2"
)

// Bot wraps a disgo *bot.Client and owns the Router.
type Bot struct {
	client  *bot.Client
	Router  *Router
	Sharded bool
}

// New creates a Bot. When sharded is true the shard manager is used with
// auto-scaling enabled; otherwise a single gateway connection is opened.
// Extra bot.ConfigOpt values are forwarded to disgo.New.
func New(token string, sharded bool, opts ...bot.ConfigOpt) (*Bot, error) {
	r := newRouter()

	buildOpts := []bot.ConfigOpt{bot.WithEventListeners(r)}

	if sharded {
		buildOpts = append(buildOpts,
			bot.WithShardManagerConfigOpts(
				sharding.WithAutoScaling(true),
				sharding.WithGatewayConfigOpts(
					gateway.WithAutoReconnect(true),
				),
			),
		)
	} else {
		buildOpts = append(buildOpts,
			bot.WithGatewayConfigOpts(
				gateway.WithAutoReconnect(true),
			),
		)
	}

	buildOpts = append(buildOpts, opts...)

	client, err := disgo.New(token, buildOpts...)
	if err != nil {
		return nil, err
	}

	return &Bot{
		client:  client,
		Router:  r,
		Sharded: sharded,
	}, nil
}

// Client returns the underlying disgo bot.Client.
func (b *Bot) Client() *bot.Client {
	return b.client
}

// ApplicationID returns the bot application ID.
func (b *Bot) ApplicationID() snowflake.ID {
	return b.client.ApplicationID
}

// Open connects to Discord. Uses the shard manager when Sharded is true.
func (b *Bot) Open(ctx context.Context) error {
	if b.Sharded {
		return b.client.OpenShardManager(ctx)
	}

	return b.client.OpenGateway(ctx)
}

// Close disconnects from Discord.
func (b *Bot) Close(ctx context.Context) {
	b.client.Close(ctx)
}

// AddEventListeners adds extra event listeners alongside the router.
func (b *Bot) AddEventListeners(listeners ...bot.EventListener) {
	b.client.EventManager.AddEventListeners(listeners...)
}
