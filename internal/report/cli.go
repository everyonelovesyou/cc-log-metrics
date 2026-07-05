package report

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cc-log-metrics/internal/extract"
)

// Main は report サブコマンド。events(.enriched).jsonl から report.md / metrics.json を出す。
func Main(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	by := fs.String("by", "", "出力する粒度の絞り込み (session|project|month)。省略時は3粒度すべて")
	out := fs.String("o", "out/report.md", "Markdown レポートの出力先")
	jsonOut := fs.String("json", "out/metrics.json", "機械可読メトリクスの出力先")
	if err := fs.Parse(args); err != nil {
		return err
	}

	in := ""
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
	} else {
		// enriched があれば優先、なければ素の events
		in = "out/events.enriched.jsonl"
		if _, err := os.Stat(in); err != nil {
			in = "out/events.jsonl"
		}
	}

	events, err := readEvents(in)
	if err != nil {
		return err
	}

	grans := []string{"month", "project", "session"}
	if *by != "" {
		if *by != "session" && *by != "project" && *by != "month" {
			return fmt.Errorf("--by は session|project|month のいずれか: %s", *by)
		}
		grans = []string{*by}
	}
	groups := map[string][]Group{}
	for _, g := range grans {
		groups[g] = Aggregate(events, g)
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	mf, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer mf.Close()
	if err := RenderMarkdown(mf, groups); err != nil {
		return err
	}
	jf, err := os.Create(*jsonOut)
	if err != nil {
		return err
	}
	defer jf.Close()
	if err := RenderJSON(jf, groups); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d イベント (%s) から %s と %s を生成しました\n", len(events), in, *out, *jsonOut)
	return nil
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
