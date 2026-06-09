package disgoplus

import (
	"testing"
)

func TestCtxNext_RunsHandlersInOrder(t *testing.T) {
	order := []int{}

	makeHandler := func(n int) Handler {
		return HandlerFunc(func(ctx *Ctx) {
			order = append(order, n)
			ctx.Next()
		})
	}

	ctx := &Ctx{
		remaining: []Handler{
			makeHandler(1),
			makeHandler(2),
			makeHandler(3),
		},
	}

	ctx.Next()

	if len(order) != 3 {
		t.Fatalf("expected 3 handlers called, got %d", len(order))
	}
	for i, v := range order {
		if v != i+1 {
			t.Fatalf("expected order[%d]=%d, got %d", i, i+1, v)
		}
	}
}

func TestCtxNext_StopsIfNotAdvanced(t *testing.T) {
	called := 0

	h1 := HandlerFunc(func(ctx *Ctx) {
		called++
		// deliberately does NOT call ctx.Next()
	})
	h2 := HandlerFunc(func(ctx *Ctx) {
		called++
	})

	ctx := &Ctx{remaining: []Handler{h1, h2}}
	ctx.Next()

	if called != 1 {
		t.Fatalf("expected 1 handler called, got %d", called)
	}
}

func TestCtxNext_EmptyChain(t *testing.T) {
	ctx := &Ctx{}
	// should not panic
	ctx.Next()
}

func TestHandlerFunc(t *testing.T) {
	called := false
	h := HandlerFunc(func(ctx *Ctx) { called = true })
	h.HandleCommand(&Ctx{})
	if !called {
		t.Fatal("HandlerFunc.HandleCommand should have called the function")
	}
}
