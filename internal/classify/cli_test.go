package classify

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cc-log-metrics/internal/extract"
)

// writeJSONL は Event 列を JSONL 形式で書き出す (テスト用ヘルパー)。
func writeJSONL(t *testing.T, path string, events []extract.Event) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
}

func readJSONL(t *testing.T, path string) []extract.Event {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	var events []extract.Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var ev extract.Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		events = append(events, ev)
	}
	return events
}

// TestMainFallsBackWhenClaudeMissing は claude CLI が PATH にない場合、
// フォールバックせず未分類のまま入力を出力へコピーして正常終了することを確認する。
func TestMainFallsBackWhenClaudeMissing(t *testing.T) {
	// PATH を空ディレクトリのみにして claude を「見つからない」状態にする。
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	in := filepath.Join(dir, "events.jsonl")
	out := filepath.Join(dir, "events.enriched.jsonl")
	writeJSONL(t, in, []extract.Event{corr("そうじゃなくて A", "")})

	if err := Main([]string{"-o", out, in}); err != nil {
		t.Fatalf("Main: %v", err)
	}

	got := readJSONL(t, out)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Classification != "" {
		t.Errorf("classification = %q, want 未分類のまま", got[0].Classification)
	}
}

// TestMainTrailingFlagAfterPositional は、位置引数 (入力パス) の後に置かれた
// -o フラグが無視されず反映されることを確認する。
// flag パッケージは最初の非フラグ引数でパースを止めるため、
// 事前に修正しないと -o は無視され out/events.enriched.jsonl (デフォルト) に書かれてしまう。
func TestMainTrailingFlagAfterPositional(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	in := filepath.Join(dir, "events.jsonl")
	out := filepath.Join(dir, "events.enriched.jsonl")
	writeJSONL(t, in, []extract.Event{corr("そうじゃなくて A", "")})

	if err := Main([]string{in, "-o", out}); err != nil {
		t.Fatalf("Main: %v", err)
	}

	got := readJSONL(t, out)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (トレイリング -o が反映されていない)", len(got))
	}
}

// TestMainTrailingBatchZeroRejected は、位置引数の後に置かれた --batch 0 でも
// 従来どおりバリデーションエラーになることを確認する。
func TestMainTrailingBatchZeroRejected(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	in := filepath.Join(dir, "events.jsonl")
	writeJSONL(t, in, []extract.Event{corr("そうじゃなくて A", "")})

	err := Main([]string{in, "--batch", "0"})
	if err == nil {
		t.Fatal("Main: エラーを期待したが nil だった (トレイリング --batch 0 が検証されていない)")
	}
}
