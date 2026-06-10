package disgoplus

import (
	"testing"
)

func TestNewRouter_Empty(t *testing.T) {
	r := NewRouter(nil)
	if r.Count() != 0 {
		t.Fatalf("expected 0 commands, got %d", r.Count())
	}
}

func TestNewRouter_PrePopulated(t *testing.T) {
	cmds := []*Command{
		{Name: "ping", Description: "ping pong"},
		{Name: "help", Description: "show help"},
	}

	r := NewRouter(cmds)
	if r.Count() != 2 {
		t.Fatalf("expected 2 commands, got %d", r.Count())
	}
}

func TestRouter_RegisterDuplicate(t *testing.T) {
	r := newRouter()
	cmd := &Command{Name: "test", Handler: HandlerFunc(func(*Ctx) {})}
	r.Register(cmd)
	r.Register(&Command{Name: "test", Handler: HandlerFunc(func(*Ctx) {})})

	if r.Count() != 1 {
		t.Fatalf(
			"duplicate register should be a no-op, got %d commands",
			r.Count(),
		)
	}
}

func TestRouter_GetComponent_ExactMatch(t *testing.T) {
	r := newRouter()
	r.RegisterMessageComponent(&MessageComponent{
		CustomID: "BUTTON",
		Handler:  HandlerFunc(func(*Ctx) {}),
	})

	mc, params := r.getComponent("BUTTON")
	if mc == nil {
		t.Fatal("expected component to be found")
	}

	if len(params) != 0 {
		t.Fatalf("expected no params, got %v", params)
	}
}

func TestRouter_GetComponent_SlugMatch(t *testing.T) {
	r := newRouter()
	r.RegisterMessageComponent(&MessageComponent{
		CustomID: "LEADERBOARD/:page",
		Handler:  HandlerFunc(func(*Ctx) {}),
	})

	mc, params := r.getComponent("LEADERBOARD/5")
	if mc == nil {
		t.Fatal("expected slug component to match")
	}

	if params["page"] != "5" {
		t.Fatalf("expected page=5, got %q", params["page"])
	}
}

func TestRouter_GetComponent_NoMatch(t *testing.T) {
	r := newRouter()
	r.RegisterMessageComponent(&MessageComponent{
		CustomID: "LEADERBOARD/:page",
		Handler:  HandlerFunc(func(*Ctx) {}),
	})

	mc, _ := r.getComponent("OTHER/5")
	if mc != nil {
		t.Fatal("expected no match")
	}
}

func TestRouter_GetModal_SlugMatch(t *testing.T) {
	r := newRouter()
	r.RegisterModal(&Modal{
		CustomID: "RESET_LEADERBOARD/:userID",
		Handler:  HandlerFunc(func(*Ctx) {}),
	})

	m, params := r.getModal("RESET_LEADERBOARD/789")
	if m == nil {
		t.Fatal("expected modal to match")
	}

	if params["userID"] != "789" {
		t.Fatalf("expected userID=789, got %q", params["userID"])
	}
}

func TestCommand_ToSlashCommandCreate_TopLevel(t *testing.T) {
	cmd := &Command{
		Name:        "ping",
		Description: "ping pong",
		Handler:     HandlerFunc(func(*Ctx) {}),
	}

	create := cmd.toSlashCommandCreate()
	if create.Name != "ping" {
		t.Fatalf("expected name ping, got %q", create.Name)
	}

	if create.Description != "ping pong" {
		t.Fatalf("unexpected description: %q", create.Description)
	}

	if len(create.Options) != 0 {
		t.Fatalf(
			"expected no options for leaf command, got %d",
			len(create.Options),
		)
	}
}

func TestCommand_ToSlashCommandCreate_WithSubCommands(t *testing.T) {
	subRouter := NewRouter([]*Command{
		{
			Name:        "show",
			Description: "show settings",
			Handler:     HandlerFunc(func(*Ctx) {}),
		},
		{
			Name:        "reset",
			Description: "reset settings",
			Handler:     HandlerFunc(func(*Ctx) {}),
		},
	})
	cmd := &Command{
		Name:        "settings",
		Description: "manage settings",
		SubCommands: subRouter,
	}

	create := cmd.toSlashCommandCreate()
	if len(create.Options) != 2 {
		t.Fatalf("expected 2 sub-command options, got %d", len(create.Options))
	}
}
