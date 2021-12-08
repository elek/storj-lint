package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/elek/golangci-lint/pkg/commands"
	"github.com/elek/storj-lint/align"
	"github.com/elek/storj-lint/copyright"
	"github.com/elek/storj-lint/imports"
	"github.com/elek/storj-lint/largefiles"
	"github.com/elek/storj-lint/monitoring"
	"github.com/elek/storj-lint/peer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
)

type check func() error
type argsAdjuster func([]string) []string
type Check struct {
	Name    string
	Checker check
	Args    argsAdjuster
}

var rootCmd cobra.Command

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func init() {
	checkers := []Check{
		{
			Name:    "check-copyright",
			Checker: copyright.CheckCopyright,
		},
		{
			Name: "check-imports",
			Checker: func() error {
				exitCode := imports.Run()
				if exitCode != 0 {
					return errs.New("import check is failed")
				}
				return nil
			},
		},
		{
			Name: "check-peer-constraints",
			Checker: func() error {
				return peer.Run()
			},
		},
		{
			Name: "check-monitoring",
			Checker: func() error {
				return monitoring.Run()
			},
		},
		{
			Name: "check-large-files",
			Checker: func() error {
				return largefiles.Run()
			},
		},
		{
			Name: "check-atomic-align",
			Checker: func() error {
				exitCode := align.Run()
				if exitCode != 0 {
					return errs.New("align check is failed")
				}
				return nil
			},
		},
		{
			Name: "golangci-lint",
			Checker: func() error {
				e := commands.NewExecutor(version, commit, date)
				if err := e.Execute(); err != nil {
					fmt.Fprintf(os.Stderr, "failed executing command with error %v\n", err)
					return errs.New("golangcilint error: %v", err)
				}
				if e.ExitCode() != 0 {
					return errs.New("there are reported golangci-lint failures")
				}
				return nil
			},
			Args: func(original []string) []string {
				return []string{original[0], "run"}

			},
		},
	}

	rootCmd = cobra.Command{
		Use: "storj-lint <subcheck>",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "lint",
		Short: "RUN ALL LINTING CHECKS",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunChecks(checkers)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Write the default .golangci.yaml descriptor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ioutil.WriteFile(".golangci.yml", commands.StorjDefaults, 0644)
		},
	})

	for _, c := range checkers {
		execute := func() error {
			args := []string{os.Args[0]}
			if len(os.Args) > 2 {
				args = append(args, os.Args[2:]...)
			}
			os.Args = args
			return c.Checker()
		}
		checkerCommand := &cobra.Command{
			Use:   c.Name,
			Short: fmt.Sprintf("Execute only the '%s' with parameters", c.Name),
			RunE: func(cmd *cobra.Command, args []string) error {
				return execute()
			},
		}
		checkerCommand.SetHelpFunc(func(command *cobra.Command, strings []string) {
			_ = execute()
		})
		rootCmd.AddCommand(checkerCommand)
	}
}

func RunChecks(checkers []Check) error {
	origArgs := os.Args
	failed := false
	for _, c := range checkers {
		fmt.Println("---------------------------------------------------------------")
		color.Yellow("Running check: %s", c.Name)
		if c.Args != nil {
			os.Args = c.Args(origArgs)
		} else {
			args := []string{os.Args[0]}
			if len(os.Args) > 2 {
				args = append(args, os.Args[2:]...)
			}
			os.Args = args
		}
		err := c.Checker()
		if err != nil {
			color.Red("FAILED: %s", err)
			failed = true
		} else {
			color.Green("PASSED")
		}
		os.Args = origArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	if failed {
		return errs.New("Check is failed")
	}
	return nil
}

var (
	// Populated by goreleaser during build
	version = "master"
	commit  = "?"
	date    = ""
)
