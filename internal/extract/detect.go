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
	for _, e := range entries {
		switch {
		case isPrompt(e):
			text := e.UserText()
			ev := base(KindUserPrompt, e)
			ev.Detail = map[string]any{"chars": len([]rune(text))}
			events = append(events, ev)
			if pattern, ok := lex.Match(text); ok {
				cv := base(KindCorrection, e)
				cv.Detail = map[string]any{"utterance": text, "pattern": pattern}
				events = append(events, cv)
			}
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
