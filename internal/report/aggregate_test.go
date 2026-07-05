package report

import (
	"math"
	"testing"

	"cc-log-metrics/internal/extract"
)

func ev(kind, project, session, ts string, detail map[string]any) extract.Event {
	return extract.Event{Kind: kind, Project: project, SessionID: session, Timestamp: ts, Detail: detail}
}

func sampleEvents() []extract.Event {
	return []extract.Event{
		ev(extract.KindUserPrompt, "/p/alpha", "s1", "2026-06-01T10:00:00.000Z", map[string]any{"chars": 10}),
		ev(extract.KindUserPrompt, "/p/alpha", "s1", "2026-06-01T10:05:00.000Z", map[string]any{"chars": 20}),
		ev(extract.KindCorrection, "/p/alpha", "s1", "2026-06-01T10:05:00.000Z", map[string]any{"utterance": "そうじゃなくて", "pattern": "そうじゃなくて"}),
		ev(extract.KindRewind, "/p/alpha", "s1", "2026-06-01T10:06:00.000Z", nil),
		ev(extract.KindTurn, "/p/alpha", "s1", "2026-06-01T10:07:00.000Z", map[string]any{"durationMs": float64(30000), "messageCount": float64(5)}),
		ev(extract.KindUserPrompt, "/p/beta", "s2", "2026-07-01T10:00:00.000Z", map[string]any{"chars": 30}),
		ev(extract.KindClearBoundary, "/p/beta", "s2", "2026-07-01T10:00:00.000Z", map[string]any{"reason": "session_start", "nextPromptChars": float64(30)}),
	}
}

func TestAggregateBySession(t *testing.T) {
	groups := Aggregate(sampleEvents(), "session")
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}
	s1 := groups[0]
	if s1.Key != "s1" {
		t.Fatalf("Key 昇順のはず: %+v", groups)
	}
	if s1.Prompts != 2 || s1.Rewinds != 1 || s1.Corrections != 1 {
		t.Errorf("s1 カウント不正: %+v", s1)
	}
	// 介入率 = (1+1)/2 = 1.0
	if math.Abs(s1.InterventionRate-1.0) > 1e-9 {
		t.Errorf("InterventionRate = %f, want 1.0", s1.InterventionRate)
	}
	if math.Abs(s1.AvgTurnSeconds-30.0) > 1e-9 {
		t.Errorf("AvgTurnSeconds = %f, want 30.0", s1.AvgTurnSeconds)
	}
}

func TestAggregateByMonth(t *testing.T) {
	groups := Aggregate(sampleEvents(), "month")
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2 (2026-06 と 2026-07)", len(groups))
	}
	if groups[0].Key != "2026-06" || groups[1].Key != "2026-07" {
		t.Errorf("月キー不正: %q, %q", groups[0].Key, groups[1].Key)
	}
	if math.Abs(groups[1].AvgReexplainChars-30.0) > 1e-9 {
		t.Errorf("AvgReexplainChars = %f, want 30.0", groups[1].AvgReexplainChars)
	}
}
