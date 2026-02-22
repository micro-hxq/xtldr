package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"xtldr/internal/clipboardutil"
	"xtldr/internal/generator"
	"xtldr/internal/model"
	"xtldr/internal/ui"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "help":
			printHelp(stdout)
			return 0
		case "version":
			printVersion(stdout)
			return 0
		}
	}

	fs := flag.NewFlagSet("xtldr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	hideExplanation := fs.Bool("e", false, "hide Command Explanation panel")
	showVersion := fs.Bool("version", false, "print version information")
	fs.BoolVar(showVersion, "v", false, "print version information")
	fs.Usage = func() { printHelp(stderr) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHelp(stdout)
			return 0
		}
		fmt.Fprintf(stderr, "❌ Invalid arguments: %v\n", err)
		return 2
	}

	if *showVersion {
		printVersion(stdout)
		return 0
	}

	request := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if request == "" {
		printHelp(stdout)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "❌ Failed to get working directory: %v\n", err)
		return 1
	}

	generatorClient := generator.NewCopilot()
	copier := clipboardutil.SystemCopier{}
	loader := func() ([]model.Candidate, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		return generatorClient.Generate(ctx, request, workingDir)
	}

	program := tea.NewProgram(ui.NewLoadingModel(loader, copier, !*hideExplanation))
	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(stderr, "❌ Failed to run interactive UI: %v\n", err)
		return 1
	}

	if selected := selectedCommand(finalModel); selected != "" {
		fmt.Fprintln(stdout, selected)
		if err := copier.Copy(selected); err != nil {
			fmt.Fprintf(stderr, "❌ Failed to copy selected command: %v\n", err)
			return 1
		}
		fmt.Fprintln(stderr, "📋 Command copied to clipboard.")
	}

	return 0
}

func selectedCommand(finalModel tea.Model) string {
	switch m := finalModel.(type) {
	case ui.Model:
		return strings.TrimSpace(m.SelectedCommand())
	case *ui.Model:
		return strings.TrimSpace(m.SelectedCommand())
	default:
		return ""
	}
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "🚀 xtldr version %s\n", Version)
	fmt.Fprintf(w, "🔖 commit: %s\n", Commit)
	fmt.Fprintf(w, "🕒 build date: %s\n", BuildDate)
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "✨ xtldr - Copilot-powered command suggestion tool")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  xtldr [flags] <request>")
	fmt.Fprintln(w, "  xtldr help")
	fmt.Fprintln(w, "  xtldr version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -e          Hide Command Explanation panel")
	fmt.Fprintln(w, "  -v          Print version information")
	fmt.Fprintln(w, "  --version   Print version information")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, `  xtldr "find large files in current directory"`)
	fmt.Fprintln(w, `  xtldr -e "show top 10 processes by memory on macOS"`)
	fmt.Fprintln(w, "  xtldr version")
}
