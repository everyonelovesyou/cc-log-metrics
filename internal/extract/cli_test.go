package extract

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractMain(t *testing.T) {
	// フィクスチャをプロジェクトディレクトリ構造に見立てて配置
	root := t.TempDir()
	projDir := filepath.Join(root, "-Users-sample-Dev-sample")
	os.MkdirAll(projDir, 0o755)
	src, _ := os.ReadFile("../../testdata/session_small.jsonl")
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), src, 0o644)

	out := filepath.Join(t.TempDir(), "events.jsonl")
	err := Main([]string{"--projects", filepath.Join(root, "*"), "-o", out})
	if err != nil {
		t.Fatalf("Main: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("出力がない: %v", err)
	}
	defer f.Close()
	var kinds []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var ev Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			t.Fatalf("出力行が壊れている: %v", err)
		}
		kinds = append(kinds, ev.Kind)
		if ev.Project != "/Users/sample/Dev/sample" {
			t.Errorf("project = %q (cwd から取得するはず)", ev.Project)
		}
	}
	// user_prompt 3 + correction 1 + rewind 1 + boundary 2 + compact 1 + turn 1 + tool_error 1 = 10
	if len(kinds) != 10 {
		t.Errorf("イベント数 = %d, want 10 (%v)", len(kinds), kinds)
	}
}

// TestExtractMainExtraPositionalRejected は、extract サブコマンドに余剰の
// 位置引数を渡した場合に無言で無視せずエラーを返すことを確認する。
func TestExtractMainExtraPositionalRejected(t *testing.T) {
	root := t.TempDir()
	projDir := filepath.Join(root, "-Users-sample-Dev-sample")
	os.MkdirAll(projDir, 0o755)
	src, _ := os.ReadFile("../../testdata/session_small.jsonl")
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), src, 0o644)

	out := filepath.Join(t.TempDir(), "events.jsonl")
	err := Main([]string{"--projects", filepath.Join(root, "*"), "-o", out, "extra-arg"})
	if err == nil {
		t.Fatal("Main: エラーを期待したが nil だった (余剰の位置引数が無視されている)")
	}
	if !strings.Contains(err.Error(), "想定外の引数") {
		t.Errorf("エラーメッセージ = %q, want 想定外の引数 を含む", err.Error())
	}
}

func TestExtractSinceFilter(t *testing.T) {
	root := t.TempDir()
	projDir := filepath.Join(root, "-Users-sample-Dev-sample")
	os.MkdirAll(projDir, 0o755)
	src, _ := os.ReadFile("../../testdata/session_small.jsonl")
	os.WriteFile(filepath.Join(projDir, "s1.jsonl"), src, 0o644)

	out := filepath.Join(t.TempDir(), "events.jsonl")
	// フィクスチャは全行 2026-07-01。未来日付で全部落ちる
	err := Main([]string{"--projects", filepath.Join(root, "*"), "--since", "2026-08-01", "-o", out})
	if err != nil {
		t.Fatalf("Main: %v", err)
	}
	b, _ := os.ReadFile(out)
	if len(b) != 0 {
		t.Errorf("since フィルタ後は空のはず: %q", b)
	}
}
