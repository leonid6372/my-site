// Command api is the backend of the personal site: it imports content into
// PostgreSQL and serves the read-only JSON API (see specs/backend/spec.md).
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// errNotImplemented is returned by subcommands that are not built yet.
var errNotImplemented = errors.New("not implemented")

type command struct {
	name  string
	usage string
	run   func(args []string) error
}

var commands = []command{
	{"serve", "run migrations and start the HTTP server", runServe},
	{"seed", "validate content/ and replace all content in the database", runSeed},
	{"validate", "validate content/ only (no database access)", runValidate},
	{"migrate", "run database migrations only", runMigrate},
	{"start", "seed, then serve (container default command)", runStart},
	{"healthcheck", "check the local /api/v1/health endpoint (exit 0 if healthy)", runHealthcheck},
}

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run dispatches to a subcommand and returns the process exit code:
// 0 on success, 1 on command failure, 2 on usage errors.
func run(args []string, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	for _, c := range commands {
		if c.name == args[0] {
			if err := c.run(args[1:]); err != nil {
				fmt.Fprintf(stderr, "api %s: %v\n", c.name, err)
				return 1
			}
			return 0
		}
	}
	fmt.Fprintf(stderr, "api: unknown command %q\n\n", args[0])
	printUsage(stderr)
	return 2
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: api <command>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	for _, c := range commands {
		fmt.Fprintf(w, "  %-12s %s\n", c.name, c.usage)
	}
}

func runServe(_ []string) error       { return errNotImplemented }
func runSeed(_ []string) error        { return errNotImplemented }
func runValidate(_ []string) error    { return errNotImplemented }
func runMigrate(_ []string) error     { return errNotImplemented }
func runStart(_ []string) error       { return errNotImplemented }
func runHealthcheck(_ []string) error { return errNotImplemented }
