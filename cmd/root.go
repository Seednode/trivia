/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "0.1.0"
)

var (
	bind        string
	exitOnError bool
	files       []string
	port        uint16
	profile     bool
	verbose     bool
	version     bool

	requiredArgs = []string{}

	rootCmd = &cobra.Command{
		Use:   "trivia",
		Short: "Serves a basic trivia web frontend.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := servePage()

			return err
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().BoolVar(&exitOnError, "exit-on-error", false, "shut down webserver on error, instead of just printing the error")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&profile, "profile", false, "register net/http/pprof handlers")
	rootCmd.Flags().StringSliceVarP(&files, "question-file", "f", []string{}, "path to file containing trivia questions (can be supplied multiple times)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "log requests to stdout")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "display version and exit")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.MarkFlagsOneRequired(requiredArgs...)

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("trivia v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
