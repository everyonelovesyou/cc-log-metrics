package transcript

import (
	"bufio"
	"encoding/json"
	"os"
)

// Stats は読み取りの成否集計。Skipped は JSON として解釈できず捨てた行数。
type Stats struct {
	Lines   int
	Skipped int
}

// ReadFile は1セッションの jsonl を読み取る。壊れた行はスキップして数える。
func ReadFile(path string) ([]Entry, Stats, error) {
	var stats Stats
	f, err := os.Open(path)
	if err != nil {
		return nil, stats, err
	}
	defer f.Close()

	// 実ログで1行 300KB 程度を観測済み。余裕を持って上限 16MB。
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var entries []Entry
	for sc.Scan() {
		stats.Lines++
		var e Entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			stats.Skipped++
			continue
		}
		entries = append(entries, e)
	}
	return entries, stats, sc.Err()
}
