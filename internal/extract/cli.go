package extract

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cc-log-metrics/internal/transcript"
)

// Main は extract サブコマンド。生ログを走査して events.jsonl を出力する。
func Main(args []string) error {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	home, _ := os.UserHomeDir()
	projects := fs.String("projects", filepath.Join(home, ".claude", "projects", "*"), "プロジェクトディレクトリの glob")
	since := fs.String("since", "", "この日付以降のみ対象 (YYYY-MM-DD)")
	lexPath := fs.String("lexicon", "", "correction 語彙 JSON のパス (省略時は埋め込み既定)")
	out := fs.String("o", "out/events.jsonl", "出力先")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("想定外の引数です: %v", fs.Args())
	}

	lex := DefaultLexicon()
	if *lexPath != "" {
		var err error
		if lex, err = LoadLexicon(*lexPath); err != nil {
			return fmt.Errorf("語彙の読み込みに失敗: %w", err)
		}
	}

	dirs, err := filepath.Glob(*projects)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("プロジェクトが見つかりません: %s", *projects)
	}

	var events []Event
	var total transcript.Stats
	files := 0
	for _, dir := range dirs {
		sessions, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		for _, path := range sessions {
			entries, stats, err := transcript.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: %s を読めません: %v\n", path, err)
				continue
			}
			files++
			total.Lines += stats.Lines
			total.Skipped += stats.Skipped
			events = append(events, FromSession(entries, projectOf(entries, dir), lex)...)
		}
	}

	if *since != "" {
		cutoff := *since + "T00:00:00.000Z"
		var kept []Event
		for _, ev := range events {
			if ev.Timestamp >= cutoff {
				kept = append(kept, ev)
			}
		}
		events = kept
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := writeEvents(f, events); err != nil {
		return err
	}

	rate := 0.0
	if total.Lines > 0 {
		rate = float64(total.Skipped) / float64(total.Lines) * 100
	}
	fmt.Fprintf(os.Stderr, "%d ファイル / %d 行を処理、%d 行をスキップ (%.1f%%)、%d イベントを %s へ出力\n",
		files, total.Lines, total.Skipped, rate, len(events), *out)
	if rate > 5.0 {
		fmt.Fprintln(os.Stderr, "⚠ スキップ率が 5% を超えています。ログのスキーマが変わった可能性があります")
	}
	return nil
}

// projectOf はエントリの cwd からプロジェクトパスを決める。cwd がなければディレクトリ名。
func projectOf(entries []transcript.Entry, dir string) string {
	for _, e := range entries {
		if e.CWD != "" {
			return e.CWD
		}
	}
	return filepath.Base(dir)
}

func writeEvents(w io.Writer, events []Event) error {
	enc := json.NewEncoder(w)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return err
		}
	}
	return nil
}

// ReadEvents は events(.enriched).jsonl を読み込み Event 列を返す。
// classify / report 両サブコマンドで共通に使う。
func ReadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var events []Event
	for sc.Scan() {
		var ev Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("%s の行を解釈できません: %w", path, err)
		}
		events = append(events, ev)
	}
	return events, sc.Err()
}
