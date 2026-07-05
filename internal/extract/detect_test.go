package extract

import (
	"testing"

	"cc-log-metrics/internal/transcript"
)

func fixtureEvents(t *testing.T) []Event {
	t.Helper()
	entries, _, err := transcript.ReadFile("../../testdata/session_small.jsonl")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return FromSession(entries, "/Users/sample/Dev/sample", nil)
}

func countKind(events []Event, kind string) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

func TestUserPromptDetection(t *testing.T) {
	events := fixtureEvents(t)
	// u1, u2, u5 の3件。u3 (tool_result のみ)・u4 (/clear コマンド)・u6 (compact 要約) は発話ではない
	if got := countKind(events, KindUserPrompt); got != 3 {
		t.Errorf("user_prompt = %d, want 3", got)
	}
	for _, ev := range events {
		if ev.Kind == KindUserPrompt {
			if ev.Project != "/Users/sample/Dev/sample" || ev.SessionID != "s1" {
				t.Errorf("共通フィールドが不正: %+v", ev)
			}
			if _, ok := ev.Detail["chars"].(int); !ok {
				t.Errorf("detail.chars がない: %+v", ev)
			}
		}
	}
}

func TestTurnDetection(t *testing.T) {
	events := fixtureEvents(t)
	if got := countKind(events, KindTurn); got != 1 {
		t.Fatalf("turn = %d, want 1", got)
	}
	for _, ev := range events {
		if ev.Kind == KindTurn && ev.Detail["durationMs"].(int64) != 30000 {
			t.Errorf("durationMs = %v, want 30000", ev.Detail["durationMs"])
		}
	}
}

func TestToolErrorDetection(t *testing.T) {
	events := fixtureEvents(t)
	if got := countKind(events, KindToolError); got != 1 {
		t.Errorf("tool_error = %d, want 1", got)
	}
}
