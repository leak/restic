package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/restic/restic/internal/debug"
	"github.com/restic/restic/internal/options"
	"github.com/restic/restic/internal/restic"

	"github.com/spf13/cobra"

	"github.com/restic/restic/internal/config"
	"github.com/restic/restic/internal/errors"
)

var cmdRoot = &config.RootCmd

var logBuffer = bytes.NewBuffer(nil)

func init() {
	// install custom global logger into a buffer, if an error occurs
	// we can show the logs
	log.SetOutput(logBuffer)

	cmdRoot.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		// set verbosity, default is one
		globalOptions.verbosity = 1
		if globalOptions.Quiet && (globalOptions.Verbose > 1) {
			return errors.Fatal("--quiet and --verbose cannot be specified at the same time")
		}

		switch {
		case globalOptions.Verbose >= 2:
			globalOptions.verbosity = 3
		case globalOptions.Verbose > 0:
			globalOptions.verbosity = 2
		case globalOptions.Quiet:
			globalOptions.verbosity = 0
		}

		// parse extended options
		opts, err := options.Parse(globalOptions.Options)
		if err != nil {
			return err
		}
		globalOptions.extended = opts
		if c.Name() == "version" {
			return nil
		}

		// run the debug functions for all subcommands (if build tag "debug" is
		// enabled)
		if err := runDebug(); err != nil {
			return err
		}

		return nil
	}
}

func main() {
	debug.Log("main %#v", os.Args)
	debug.Log("restic %s compiled with %v on %v/%v",
		version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	err := cmdRoot.Execute()

	switch {
	case restic.IsAlreadyLocked(errors.Cause(err)):
		fmt.Fprintf(os.Stderr, "%v\nthe `unlock` command can be used to remove stale locks\n", err)
	case errors.IsFatal(errors.Cause(err)):
		fmt.Fprintf(os.Stderr, "%v\n", err)
	case err != nil:
		fmt.Fprintf(os.Stderr, "%+v\n", err)

		if logBuffer.Len() > 0 {
			fmt.Fprintf(os.Stderr, "also, the following messages were logged by a library:\n")
			sc := bufio.NewScanner(logBuffer)
			for sc.Scan() {
				fmt.Fprintln(os.Stderr, sc.Text())
			}
		}
	}

	var exitCode int
	if err != nil {
		exitCode = 1
	}

	Exit(exitCode)
}
