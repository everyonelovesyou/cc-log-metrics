// Package extract はトランスクリプトを正規化イベントに変換する。
package extract

// イベント種別。設計書「イベントモデル」に対応する。
const (
	KindUserPrompt    = "user_prompt"
	KindRewind        = "rewind"
	KindCorrection    = "correction"
	KindClearBoundary = "clear_boundary"
	KindCompact       = "compact"
	KindTurn          = "turn"
	KindToolError     = "tool_error"
)

// Event は正規化イベント。events.jsonl の1行に対応する。
type Event struct {
	Kind           string         `json:"kind"`
	Project        string         `json:"project"`
	SessionID      string         `json:"sessionId"`
	Timestamp      string         `json:"timestamp"`
	Detail         map[string]any `json:"detail,omitempty"`
	Classification string         `json:"classification,omitempty"`
}
