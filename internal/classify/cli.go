package classify

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"cc-log-metrics/internal/extract"
)

// Main は classify サブコマンド。events.jsonl の correction に分類を追記する。
func Main(args []string) error {
	fs := flag.NewFlagSet("classify", flag.ContinueOnError)
	out := fs.String("o", "out/events.enriched.jsonl", "出力先")
	model := fs.String("model", "haiku", "分類に使うモデル")
	batch := fs.Int("batch", 20, "1 回の呼び出しで分類する件数")
	if err := fs.Parse(args); err != nil {
		return err
	}
	in := "out/events.jsonl"
	// flag は最初の非フラグ引数でパースを止めてしまうため、位置引数 (入力パス) を
	// 取り出したうえで残りを再パースし、位置引数の後ろに置かれたフラグも反映する。
	if fs.NArg() > 0 {
		in = fs.Arg(0)
		if err := fs.Parse(fs.Args()[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 {
			return fmt.Errorf("想定外の引数です: %v", fs.Args())
		}
	}
	if *batch <= 0 {
		return fmt.Errorf("--batch は正の整数を指定してください: %d", *batch)
	}

	events, err := readEvents(in)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Fprintln(os.Stderr, "claude CLI が見つかりません。分類せずそのまま出力します (設計どおりフォールバックはしません)")
		return write(events, *out)
	}

	n, applyErr := Apply(events, ClaudeCLI{Model: *model}, *batch)
	fmt.Fprintf(os.Stderr, "%d 件を分類しました\n", n)
	if writeErr := write(events, *out); writeErr != nil {
		return writeErr
	}
	// 途中失敗でも分類済み分は保存されている (再実行で残りだけ分類される)
	return applyErr
}

func readEvents(path string) ([]extract.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var events []extract.Event
	for sc.Scan() {
		var ev extract.Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("%s の行を解釈できません: %w", path, err)
		}
		events = append(events, ev)
	}
	return events, sc.Err()
}

func write(events []extract.Event, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return err
		}
	}
	return nil
}
