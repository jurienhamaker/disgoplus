# disgoplus

A thin wrapper around [disgo](https://github.com/disgoorg/disgo) that provides a command router, middleware chain, interaction context, and response helpers for building Discord bots.

## Features

- `Router` — registers slash commands, message components, and modals; dispatches incoming interactions
  - Slug routing — component and modal custom IDs support `:param` patterns (`"LEADERBOARD/:page"`)
- `Ctx` — unified context passed to every handler with guild/channel/user data and response methods
- Middleware chain — handlers run in order via `ctx.Next()`; attach per-command or per-component
- Sharding — pass `sharded: true` to `New` to use the disgo shard manager with auto-scaling
- Response helpers — `Defer`, `Respond`, `FollowUp`, `Update`, `ModalRespond`, `ErrorResponse`, etc.

## Requirements

- Go 1.26+

## Installation

```bash
go get github.com/jurienhamaker/disgoplus
```

## Clone, build & test

```bash
git clone https://github.com/jurienhamaker/disgoplus.git
cd disgoplus

# Build
go build ./...

# Test
go test ./...

# Test with race detector
go test -race ./...
```

## Quick start

```go
package main

import (
    "context"
    "os"
    "os/signal"

    "github.com/disgoorg/disgo/discord"
    "github.com/jurienhamaker/disgoplus"
)

func main() {
    b, err := disgoplus.New(os.Getenv("TOKEN"), false)
    if err != nil {
        panic(err)
    }

    b.Router.Register(&disgoplus.Command{
        Name:        "ping",
        Description: "Replies with pong",
        Handler: disgoplus.HandlerFunc(func(ctx *disgoplus.Ctx) {
            _ = disgoplus.Respond(ctx, discord.MessageCreate{Content: "Pong!"})
        }),
    })

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    if err := b.Open(ctx); err != nil {
        panic(err)
    }
    <-ctx.Done()
    b.Close(context.Background())
}
```

## Used by

**[Yugen Bots](https://github.com/jurienhamaker/yugen)** — a collection of Discord bots:

| Bot    | Description    |
| ------ | -------------- |
| Koto   | Wordle bot     |
| Kusari | Word Chain bot |
| Kazu   | Counting bot   |
| Hoshi  | Starboard bot  |

## License

MIT
