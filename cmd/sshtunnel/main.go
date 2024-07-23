/*
Copyright © 2024 yong
*/
package main

import (
	"os"
	
	"github.com/go-puzzles/pflags"
	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/plog/level"
	"github.com/go-puzzles/ssh-tunnel/internal/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	debug bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ssh-tunnel",
	Short: "A simple ssh tunneling tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			plog.Enable(level.LevelDebug)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflags.OverrideDefaultConfigFile(os.Getenv("HOME") + "/.ssh-tunnel.yaml")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")
	
	rootCmd.AddCommand(cmd.ForwardCmd)
	rootCmd.AddCommand(cmd.ReverseCmd)
}

func main() {
	Execute()
}
