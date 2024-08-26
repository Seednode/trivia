/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "0.23.0"
)

var (
	bind           string
	exitOnError    bool
	export         bool
	extension      string
	files          []string
	paths          []string
	port           uint16
	profile        bool
	recursive      bool
	reload         bool
	reloadInterval string
	verbose        bool
	version        bool

	ErrIncompatibleFlags = errors.New("--question-file and --question-path are mutually exclusive")

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
	rootCmd.Flags().BoolVar(&export, "export", false, "allow exporting of trivia database")
	rootCmd.Flags().StringVar(&extension, "extension", ".trivia", "only process files ending in this extension")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&profile, "profile", false, "register net/http/pprof handlers")
	rootCmd.Flags().BoolVar(&reload, "reload", false, "allow live-reload of questions")
	rootCmd.Flags().StringVar(&reloadInterval, "reload-interval", "", "interval at which to rebuild question list (e.g. \"5m\" or \"1h\")")
	rootCmd.Flags().StringSliceVarP(&files, "question-file", "f", []string{}, "path to file containing trivia questions (can be supplied multiple times)")
	rootCmd.Flags().StringSliceVar(&paths, "question-path", []string{}, "path containing trivia question files (can be supplied multiple times)")
	rootCmd.Flags().BoolVar(&recursive, "recursive", false, "recurse into directories when supplying --question-path")
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
