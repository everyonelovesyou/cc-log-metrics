package extract

import (
	_ "embed"
	"encoding/json"
	"os"
	"strings"
)

//go:embed lexicon.json
var defaultLexiconJSON []byte

type lexiconFile struct {
	Patterns []string `json:"patterns"`
}

// DefaultLexicon は埋め込みの既定語彙を返す。
func DefaultLexicon() Lexicon {
	var f lexiconFile
	// 埋め込みファイルはビルド時に検証済みのため失敗しない
	json.Unmarshal(defaultLexiconJSON, &f)
	return Lexicon(f.Patterns)
}

// LoadLexicon は外部 JSON ({"patterns": [...]}) から語彙を読む。
func LoadLexicon(path string) (Lexicon, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f lexiconFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return Lexicon(f.Patterns), nil
}

// Match は最初に部分一致した語を返す (first-match-wins)。
// この性質上、語彙の並び順には「より具体的な語を、それを含む/より汎用的な
// 部分文字列より先に置く」という不変条件がある
// (例: 「そうではなくて」は「ではなく」より前に置く。逆順だと後者が先に
// マッチしてしまい、前者には決してマッチしなくなる)。
func (l Lexicon) Match(text string) (string, bool) {
	for _, p := range l {
		if strings.Contains(text, p) {
			return p, true
		}
	}
	return "", false
}
