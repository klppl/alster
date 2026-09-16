package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alex/alster/internal/build"
	"github.com/alex/alster/internal/config"
	"github.com/alex/alster/internal/preview"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "alster:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Println("Usage: alster build [-config alster.yaml] [-base-url URL]\n       alster serve [-config alster.yaml] [-addr 127.0.0.1:8080]")
		return nil
	}
	command := args[0]
	if command != "build" && command != "serve" {
		return fmt.Errorf("unknown command %q (use alster help)", command)
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	filename := flags.String("config", "alster.yaml", "Site config file")
	var addr, base string
	if command == "serve" {
		flags.StringVar(&addr, "addr", "127.0.0.1:8080", "Preview listen address")
	} else {
		flags.StringVar(&base, "base-url", "", "Override canonical base URL")
	}
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if command == "serve" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return preview.Run(ctx, *filename, addr)
	}
	c, err := config.Load(*filename)
	if err != nil {
		return err
	}
	if base != "" {
		c.BaseURL = base
	}
	if err := build.Run(c); err != nil {
		return err
	}
	fmt.Printf("Built %s\n", c.Output)
	return nil
}
