// Package transcript は Claude Code トランスクリプト (jsonl) の読み取りを担う。
// ログ形式は非公開仕様のため、生ログの解釈はこのパッケージに閉じ込める。
package transcript

import (
	"encoding/json"
	"strings"
)

// Entry はトランスクリプトの1行。未知のフィールドは無視される。
type Entry struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	UUID             string          `json:"uuid"`
	ParentUUID       string          `json:"parentUuid"`
	SessionID        string          `json:"sessionId"`
	Timestamp        string          `json:"timestamp"`
	CWD              string          `json:"cwd"`
	IsMeta           bool            `json:"isMeta"`
	IsSidechain      bool            `json:"isSidechain"`
	IsCompactSummary bool            `json:"isCompactSummary"`
	DurationMs       int64           `json:"durationMs"`
	MessageCount     int             `json:"messageCount"`
	Message          json.RawMessage `json:"message"`
}

type messagePayload struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	IsError *bool           `json:"is_error"`
	Content json.RawMessage `json:"content"`
}

func (e Entry) payload() (messagePayload, bool) {
	var p messagePayload
	if len(e.Message) == 0 || json.Unmarshal(e.Message, &p) != nil {
		return p, false
	}
	return p, true
}

// UserText は type=user エントリの発話テキストを返す。
// content が文字列ならそのまま、ブロック配列なら text ブロックを連結する。
// tool_result のみの行や user 以外のエントリは空文字。
func (e Entry) UserText() string {
	if e.Type != "user" {
		return ""
	}
	p, ok := e.payload()
	if !ok {
		return ""
	}
	var s string
	if json.Unmarshal(p.Content, &s) == nil {
		return s
	}
	var blocks []contentBlock
	if json.Unmarshal(p.Content, &blocks) != nil {
		return ""
	}
	var text string
	for _, b := range blocks {
		if b.Type == "text" {
			text += b.Text
		}
	}
	return text
}

// HasToolError は message.content に is_error=true の tool_result があるかを返す。
func (e Entry) HasToolError() bool {
	p, ok := e.payload()
	if !ok {
		return false
	}
	var blocks []contentBlock
	if json.Unmarshal(p.Content, &blocks) != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "tool_result" && b.IsError != nil && *b.IsError {
			return true
		}
	}
	return false
}

// 権限拒否の tool_result を識別する文言。Claude Code 側の文言変更に追従が必要。
const (
	denialPrefix = "The user doesn't want to proceed with this tool use."
	// 添付メッセージの開始位置。定型注意書きは長いプレフィックスで切る
	// (「\n\nNote:」だけを目印にすると、ユーザー本文に偶然含まれる場合に過剰に切り落とすため)。
	denialSaidMarker = "the user said:\n"
	denialNoteMarker = "\n\nNote: The user's next message"
)

// PermissionDenial は「ユーザーが権限プロンプトで No を選んだ」tool_result を検出する。
// message は No に添付されたテキスト指示 (なければ空文字列)。
// 拒否文言を引用しただけの tool_result (grep 出力の転記など) を弾くため、
// is_error=true のブロックのみ判定する。複数該当時は最初の1件のみ報告する。
func (e Entry) PermissionDenial() (message string, ok bool) {
	if e.Type != "user" {
		return "", false
	}
	p, pok := e.payload()
	if !pok {
		return "", false
	}
	var blocks []contentBlock
	if json.Unmarshal(p.Content, &blocks) != nil {
		return "", false
	}
	for _, b := range blocks {
		if b.Type != "tool_result" || b.IsError == nil || !*b.IsError {
			continue
		}
		var s string
		if json.Unmarshal(b.Content, &s) != nil {
			continue // 配列形の content は対象外
		}
		if !strings.HasPrefix(s, denialPrefix) {
			continue
		}
		if i := strings.Index(s, denialSaidMarker); i >= 0 {
			msg := s[i+len(denialSaidMarker):]
			if j := strings.Index(msg, denialNoteMarker); j >= 0 {
				msg = msg[:j]
			}
			message = strings.TrimSpace(msg)
		}
		return message, true
	}
	return "", false
}
