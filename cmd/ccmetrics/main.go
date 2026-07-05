package main

import (
	"errors"
	"fmt"
	"os"
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
		return errors.New("extract: 未実装")
	case "classify":
		return errors.New("classify: 未実装")
	case "report":
		return errors.New("report: 未実装")
	default:
		return fmt.Errorf("不明なサブコマンド: %s", args[0])
	}
}
