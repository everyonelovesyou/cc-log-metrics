package transcript

import (
	"encoding/json"
	"testing"
)

func loadFixture(t *testing.T) []Entry {
	t.Helper()
	entries, _, err := ReadFile("../../testdata/session_small.jsonl")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return entries
}

func TestUserText(t *testing.T) {
	entries := loadFixture(t)
	if got := entries[0].UserText(); got != "トップページを作ってください" {
		t.Errorf("UserText = %q", got)
	}
	// tool_result のみの行 (u3) は発話ではない
	if got := entries[6].UserText(); got != "" {
		t.Errorf("tool_result 行の UserText = %q, want \"\"", got)
	}
	// assistant 行は対象外
	if got := entries[1].UserText(); got != "" {
		t.Errorf("assistant 行の UserText = %q, want \"\"", got)
	}
}

func TestHasToolError(t *testing.T) {
	entries := loadFixture(t)
	if !entries[6].HasToolError() {
		t.Error("u3 は tool_error のはず")
	}
	if entries[0].HasToolError() {
		t.Error("u1 は tool_error ではないはず")
	}
}

func entryFromJSON(t *testing.T, line string) Entry {
	t.Helper()
	var e Entry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return e
}

func TestPermissionDenial(t *testing.T) {
	// 現行 Claude Code の拒否文言 (前方一致で判定される部分)
	const deny = "The user doesn't want to proceed with this tool use. The tool use was rejected (eg. if it was a file edit, the new_string was NOT written to the file)."

	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantMsg string
	}{
		{
			name:   "メッセージなしの拒否",
			line:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":"` + deny + `"}]}}`,
			wantOK: true,
		},
		{
			name:    "メッセージあり + Note 注意書きの除去",
			line:    `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":"` + deny + ` To tell you how to proceed, the user said:\n私がサーバーを起動しますね\n\nNote: The user's next message may contain a correction or preference."}]}}`,
			wantOK:  true,
			wantMsg: "私がサーバーを起動しますね",
		},
		{
			name:    "メッセージ本文中の Note: を過剰に切り落とさない",
			line:    `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":"` + deny + ` To tell you how to proceed, the user said:\nポートは3000で。\n\nNote: これは私のメモです"}]}}`,
			wantOK:  true,
			wantMsg: "ポートは3000で。\n\nNote: これは私のメモです",
		},
		{
			name:   "自動拒否 (権限ルール) は検出しない",
			line:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":"Permission to use Bash with command npm -v has been denied."}]}}`,
			wantOK: false,
		},
		{
			name:   "通常の tool_error は検出しない",
			line:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":"command failed"}]}}`,
			wantOK: false,
		},
		{
			name:   "is_error=false の前方一致 (引用) は検出しない",
			line:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":false,"content":"` + deny + `"}]}}`,
			wantOK: false,
		},
		{
			name:   "content が配列形ならクラッシュせず非該当",
			line:   `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","is_error":true,"content":[{"type":"text","text":"` + deny + `"}]}]}}`,
			wantOK: false,
		},
		{
			name:   "user 以外のエントリは対象外",
			line:   `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_result","is_error":true,"content":"` + deny + `"}]}}`,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entryFromJSON(t, tt.line)
			msg, ok := e.PermissionDenial()
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if msg != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}
