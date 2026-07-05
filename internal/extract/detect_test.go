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

func TestRewindDetection(t *testing.T) {
	events := fixtureEvents(t)
	// u2 を親に持つ assistant が a2, a2b の2件 → rewind 1件
	if got := countKind(events, KindRewind); got != 1 {
		t.Errorf("rewind = %d, want 1", got)
	}
}

func TestBoundaryDetection(t *testing.T) {
	entries, _, err := transcript.ReadFile("../../testdata/session_small.jsonl")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	events := FromSession(entries, "/Users/sample/Dev/sample", DefaultLexicon())

	var boundaries []Event
	for _, ev := range events {
		if ev.Kind == KindClearBoundary {
			boundaries = append(boundaries, ev)
		}
	}
	// session_start + /clear の2件
	if len(boundaries) != 2 {
		t.Fatalf("clear_boundary = %d, want 2", len(boundaries))
	}
	if boundaries[0].Detail["reason"] != "session_start" {
		t.Errorf("boundaries[0].reason = %v", boundaries[0].Detail["reason"])
	}
	// session_start 直後の最初のプロンプトは u1 (14文字)
	if boundaries[0].Detail["nextPromptChars"] != 14 {
		t.Errorf("session_start の nextPromptChars = %v, want 14", boundaries[0].Detail["nextPromptChars"])
	}
	if boundaries[1].Detail["reason"] != "clear" {
		t.Errorf("boundaries[1].reason = %v", boundaries[1].Detail["reason"])
	}
	// /clear 直後の最初のプロンプトは u5 (29文字)
	if boundaries[1].Detail["nextPromptChars"] != 29 {
		t.Errorf("clear の nextPromptChars = %v, want 29", boundaries[1].Detail["nextPromptChars"])
	}
}

func TestCompactDetection(t *testing.T) {
	events := fixtureEvents(t)
	// u6 (isCompactSummary) の1件
	if got := countKind(events, KindCompact); got != 1 {
		t.Errorf("compact = %d, want 1", got)
	}
}
