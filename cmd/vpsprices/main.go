package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lesichkovm/vps-prices/internal/generator"
	"github.com/lesichkovm/vps-prices/internal/runner"
	"github.com/lesichkovm/vps-prices/internal/validator"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", []string(*i))
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch subcommand {
	case "update":
		cmd := flag.NewFlagSet("update", flag.ExitOnError)
		dryRun := cmd.Bool("dry-run", false, "Print diff summary and report without writing changes")
		timeout := cmd.Duration("timeout", 3*time.Minute, "Operation timeout")
		verbose := cmd.Bool("verbose", false, "Enable verbose logging")
		reportPath := cmd.String("report", "", "Path to write machine-readable Markdown report")
		var selectedProviders arrayFlags
		cmd.Var(&selectedProviders, "provider", "Filter by provider name (can be repeated)")

		_ = cmd.Parse(os.Args[2:])

		logLevel := slog.LevelInfo
		if *verbose {
			logLevel = slog.LevelDebug
		}
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
		slog.SetDefault(logger)

		opts := runner.UpdateOptions{
			DryRun:     *dryRun,
			Timeout:    *timeout,
			ReportPath: *reportPath,
			Providers:  selectedProviders,
		}

		if err := runner.RunUpdate(ctx, opts); err != nil {
			slog.Error("Update failed", "error", err)
			os.Exit(1)
		}

	case "generate":
		cmd := flag.NewFlagSet("generate", flag.ExitOnError)
		dataPath := cmd.String("data", "data.json", "Path to data.json")
		indexPath := cmd.String("index", "index.html", "Path to index.html")
		sitemapPath := cmd.String("sitemap", "sitemap.xml", "Path to sitemap.xml")
		_ = cmd.Parse(os.Args[2:])

		gen := generator.NewGenerator(*dataPath, *indexPath, *sitemapPath)
		if err := gen.GenerateAll(nil); err != nil {
			fmt.Printf("Generation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Pre-rendered index.html and sitemap.xml successfully.")

	case "validate":
		cmd := flag.NewFlagSet("validate", flag.ExitOnError)
		_ = cmd.Parse(os.Args[2:])

		if err := validator.ValidateDataJSON("data.json"); err != nil {
			fmt.Printf("Validation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("data.json validation passed successfully.")

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: vpsprices <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  update    Fetch, convert, and update data.json and research notes")
	fmt.Println("  generate  Pre-render index.html table rows and generate sitemap.xml")
	fmt.Println("  validate  Validate current data.json without network access")
}
