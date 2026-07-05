package extract

import (
	"os"
	"path/filepath"
	"testing"

	"cc-log-metrics/internal/transcript"
)

func TestDefaultLexiconMatch(t *testing.T) {
	lex := DefaultLexicon()
	tests := []struct {
		text        string
		wantPattern string
		wantOK      bool
	}{
		{"そうじゃなくて、一覧ページが先です", "そうじゃなくて", true},
		{"やっぱり赤にしてください", "やっぱり", true},
		{"トップページを作ってください", "", false},
	}
	for _, tt := range tests {
		pattern, ok := lex.Match(tt.text)
		if ok != tt.wantOK || pattern != tt.wantPattern {
			t.Errorf("Match(%q) = (%q, %v), want (%q, %v)", tt.text, pattern, ok, tt.wantPattern, tt.wantOK)
		}
	}
}

func TestLoadLexicon(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lex.json")
	os.WriteFile(path, []byte(`{"patterns":["ちがう"]}`), 0o644)
	lex, err := LoadLexicon(path)
	if err != nil {
		t.Fatalf("LoadLexicon: %v", err)
	}
	if _, ok := lex.Match("ちがうよ"); !ok {
		t.Error("外部語彙でマッチしない")
	}
}

func TestCorrectionDetection(t *testing.T) {
	entries, _, err := transcript.ReadFile("../../testdata/session_small.jsonl")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	events := FromSession(entries, "/Users/sample/Dev/sample", DefaultLexicon())
	// u2「そうじゃなくて、一覧ページが先です」の1件のみ
	var corrections []Event
	for _, ev := range events {
		if ev.Kind == KindCorrection {
			corrections = append(corrections, ev)
		}
	}
	if len(corrections) != 1 {
		t.Fatalf("correction = %d, want 1", len(corrections))
	}
	c := corrections[0]
	if c.Detail["pattern"] != "そうじゃなくて" {
		t.Errorf("pattern = %v", c.Detail["pattern"])
	}
	if c.Detail["utterance"] != "そうじゃなくて、一覧ページが先です" {
		t.Errorf("utterance = %v", c.Detail["utterance"])
	}
}
