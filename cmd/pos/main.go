package main

import (
	"flag"
	"fmt"
	"os"

	"pos/internal/app"
	"pos/internal/di"
)

func main() {
	interactive := flag.Bool("interactive", false, "Открыть консольное меню")
	demo := flag.Bool("demo", false, "Выполнить демонстрационный сценарий (по умолчанию)")
	policy := flag.String("policy", "zero", "Правило итога: zero или graded")
	exportPath := flag.String("export", "results.csv", "Путь к CSV для демонстрации")
	flag.Parse()
	if *interactive && *demo {
		fmt.Fprintln(os.Stderr, "Выберите один режим: --interactive или --demo")
		os.Exit(1)
	}
	container, err := di.NewDefaultContainer(*policy)
	if err == nil {
		cli := app.NewCLI(container.Service, os.Stdin, os.Stdout)
		if *interactive {
			err = cli.RunInteractive()
		} else {
			err = cli.RunDemo(*exportPath)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
