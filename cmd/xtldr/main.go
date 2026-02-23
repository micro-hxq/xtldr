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

type candidateGenerator interface {
	Generate(ctx context.Context, request, workingDir string) ([]model.Candidate, error)
}

var newGenerator = func() candidateGenerator { return generator.NewCopilot() }

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
		case "roadmap":
			printRoadmap(stdout)
			return 0
		case "version":
			printVersion(stdout)
			return 0
		}
	}

	fs := flag.NewFlagSet("xtldr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	hideExplanation := fs.Bool("e", false, "hide Command Explanation panel")
	nonInteractive := fs.Bool("non-interactive", false, "print generated commands without interactive UI")
	fs.BoolVar(nonInteractive, "n", false, "print generated commands without interactive UI")
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

	generatorClient := newGenerator()
	copier := clipboardutil.SystemCopier{}
	loader := func() ([]model.Candidate, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		return generatorClient.Generate(ctx, request, workingDir)
	}
	if *nonInteractive {
		candidates, err := loader()
		if err != nil {
			fmt.Fprintf(stderr, "❌ Failed to generate command candidates: %v\n", err)
			return 1
		}
		printCommands(stdout, candidates)
		return 0
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
	fmt.Fprintln(w, "  xtldr roadmap")
	fmt.Fprintln(w, "  xtldr version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -e          Hide Command Explanation panel")
	fmt.Fprintln(w, "  -n, --non-interactive   Print generated commands without interactive UI")
	fmt.Fprintln(w, "  -v          Print version information")
	fmt.Fprintln(w, "  --version   Print version information")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, `  xtldr "find large files in current directory"`)
	fmt.Fprintln(w, `  xtldr -e "show top 10 processes by memory on macOS"`)
	fmt.Fprintln(w, `  xtldr -n "show top 10 processes by memory on macOS"`)
	fmt.Fprintln(w, "  xtldr roadmap")
	fmt.Fprintln(w, "  xtldr version")
}

func printCommands(w io.Writer, candidates []model.Candidate) {
	for _, candidate := range candidates {
		command := strings.TrimSpace(candidate.Command)
		if command == "" {
			continue
		}
		fmt.Fprintln(w, command)
	}
}

func printRoadmap(w io.Writer) {
	fmt.Fprintln(w, "📌 Capability gaps compared with mature CLI tools:")
	fmt.Fprintln(w, "  [ ] Shell-aware output mode (bash/zsh/powershell safe command variants)")
	fmt.Fprintln(w, "  [ ] Non-interactive mode for CI/scripts (machine-readable output)")
	fmt.Fprintln(w, "  [ ] Risk guardrails (dangerous command detection + confirmation)")
	fmt.Fprintln(w, "  [ ] Execution preview (show expected impact before copy/execute)")
	fmt.Fprintln(w, "  [ ] Session history (store, search, and reuse previous requests)")
	fmt.Fprintln(w, "  [ ] Extensibility hooks (custom prompt/template and policy controls)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "✅ Please manually confirm priorities, then implement confirmed items in small increments.")
}
