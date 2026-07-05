package report

import (
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
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
}

// TestMainTrailingFlagsAfterPositional は、位置引数 (入力パス) の後に置かれた
// -o / --json フラグが無視されず反映されることを確認する。
// flag パッケージは最初の非フラグ引数でパースを止めるため、
// 事前に修正しないと -o/--json は無視されデフォルトの out/ 配下に書かれてしまう。
func TestMainTrailingFlagsAfterPositional(t *testing.T) {
	dir := t.TempDir()
	// out/ デフォルトへ書かれていないことを確認するため、カレントディレクトリを
	// テスト専用の一時ディレクトリに切り替える (他テストの残留物の影響を避ける)。
	t.Chdir(dir)
	in := filepath.Join(dir, "events.jsonl")
	writeJSONL(t, in, sampleEvents())

	wantOut := filepath.Join(dir, "report.md")
	wantJSON := filepath.Join(dir, "metrics.json")

	if err := Main([]string{in, "-o", wantOut, "--json", wantJSON}); err != nil {
		t.Fatalf("Main: %v", err)
	}

	if _, err := os.Stat(wantOut); err != nil {
		t.Errorf("report.md が指定パスに作成されていない: %v", err)
	}
	if _, err := os.Stat(wantJSON); err != nil {
		t.Errorf("metrics.json が指定パスに作成されていない: %v", err)
	}

	if _, err := os.Stat("out/report.md"); err == nil {
		t.Errorf("デフォルトの out/report.md が作られてしまっている (トレイリングフラグが無視されている)")
	}
}
