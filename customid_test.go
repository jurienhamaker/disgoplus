package disgoplus

import (
	"testing"
)

func TestTrySlug_ExactMatch(t *testing.T) {
	params, ok := trySlug("LEADERBOARD", "LEADERBOARD")
	if !ok {
		t.Fatal("expected match")
	}

	if len(params) != 0 {
		t.Fatalf("expected no params, got %v", params)
	}
}

func TestTrySlug_SingleParam(t *testing.T) {
	params, ok := trySlug("LEADERBOARD/:page", "LEADERBOARD/3")
	if !ok {
		t.Fatal("expected match")
	}

	if params["page"] != "3" {
		t.Fatalf("expected page=3, got %q", params["page"])
	}
}

func TestTrySlug_MultiParam(t *testing.T) {
	params, ok := trySlug(
		"RESET_LEADERBOARD/:reset/:userID",
		"RESET_LEADERBOARD/true/123456",
	)
	if !ok {
		t.Fatal("expected match")
	}

	if params["reset"] != "true" {
		t.Fatalf("expected reset=true, got %q", params["reset"])
	}

	if params["userID"] != "123456" {
		t.Fatalf("expected userID=123456, got %q", params["userID"])
	}
}

func TestTrySlug_NoMatch(t *testing.T) {
	_, ok := trySlug("LEADERBOARD/:page", "OTHER/3")
	if ok {
		t.Fatal("expected no match")
	}
}

func TestTrySlug_EmptyTrailingParam_NoMatch(t *testing.T) {
	// Slug parser (ported from pat) requires at least one char per param.
	// Callers that encode an optional ID must handle the empty case by using
	// a different custom-id (e.g. "RESET_LEADERBOARD/") and checking it first.
	_, ok := trySlug("RESET_LEADERBOARD/:userID", "RESET_LEADERBOARD/")
	if ok {
		t.Fatal("empty trailing param segment should not match")
	}
}

func TestTrySlug_LiteralSlash(t *testing.T) {
	_, ok := trySlug("A/B", "A/C")
	if ok {
		t.Fatal("expected no match for mismatched literal segment")
	}
}
