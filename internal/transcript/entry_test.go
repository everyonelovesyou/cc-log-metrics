package transcript

import "testing"

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
