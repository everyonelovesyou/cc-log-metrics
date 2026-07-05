package extract

import (
	"strings"

	"cc-log-metrics/internal/transcript"
)

// Lexicon は correction 判定の語彙。
type Lexicon []string

// FromSession は1セッション分のエントリから全イベントを検出する。
func FromSession(entries []transcript.Entry, project string, lex Lexicon) []Event {
	var events []Event
	base := func(kind string, e transcript.Entry) Event {
		return Event{Kind: kind, Project: project, SessionID: e.SessionID, Timestamp: e.Timestamp}
	}

	// rewind: 同一 parentUuid を持つ user/assistant エントリの分岐 (2件目以降を数える)
	seenParent := map[string]bool{}

	// clear_boundary: 境界を積んでおき、後続の最初のプロンプト文字数を埋める
	type pendingBoundary struct{ index int }
	var pending []pendingBoundary

	addBoundary := func(e transcript.Entry, reason string) {
		ev := base(KindClearBoundary, e)
		ev.Detail = map[string]any{"reason": reason, "nextPromptChars": 0}
		events = append(events, ev)
		pending = append(pending, pendingBoundary{index: len(events) - 1})
	}

	for i, e := range entries {
		if i == 0 {
			addBoundary(e, "session_start")
		}

		if (e.Type == "user" || e.Type == "assistant") && e.ParentUUID != "" {
			if seenParent[e.ParentUUID] {
				events = append(events, base(KindRewind, e))
			}
			seenParent[e.ParentUUID] = true
		}

		if e.IsCompactSummary || e.Subtype == "compact_boundary" {
			events = append(events, base(KindCompact, e))
		}

		if e.Type == "user" && strings.Contains(e.UserText(), "<command-name>/clear</command-name>") {
			addBoundary(e, "clear")
		}

		switch {
		case isPrompt(e):
			text := e.UserText()
			chars := len([]rune(text))
			ev := base(KindUserPrompt, e)
			ev.Detail = map[string]any{"chars": chars}
			events = append(events, ev)
			if pattern, ok := lex.Match(text); ok {
				cv := base(KindCorrection, e)
				cv.Detail = map[string]any{"utterance": text, "pattern": pattern}
				events = append(events, cv)
			}
			// 未解決の境界すべてに、境界後最初のプロンプト長として記録する
			for _, pb := range pending {
				events[pb.index].Detail["nextPromptChars"] = chars
			}
			pending = pending[:0]
		case e.Type == "system" && e.Subtype == "turn_duration":
			ev := base(KindTurn, e)
			ev.Detail = map[string]any{"durationMs": e.DurationMs, "messageCount": e.MessageCount}
			events = append(events, ev)
		case e.Type == "user" && e.HasToolError():
			events = append(events, base(KindToolError, e))
		}
	}
	return events
}

// isPrompt は「人間が打った発話」かを判定する。
// メタ行・サイドチェーン・compact 要約・コマンド実行痕跡・tool_result のみの行を除外する。
func isPrompt(e transcript.Entry) bool {
	if e.Type != "user" || e.IsMeta || e.IsSidechain || e.IsCompactSummary {
		return false
	}
	text := e.UserText()
	if text == "" {
		return false
	}
	if strings.Contains(text, "<command-name>") || strings.Contains(text, "<local-command") {
		return false
	}
	return true
}
