package main

import (
	"errors"
	"fmt"
	"os"

	"cc-log-metrics/internal/classify"
	"cc-log-metrics/internal/extract"
	"cc-log-metrics/internal/report"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ccmetrics:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("サブコマンドを指定してください: extract | classify | report")
	}
	switch args[0] {
	case "extract":
		return extract.Main(args[1:])
	case "classify":
		return classify.Main(args[1:])
	case "report":
		return report.Main(args[1:])
	default:
		return fmt.Errorf("不明なサブコマンド: %s", args[0])
	}
}
