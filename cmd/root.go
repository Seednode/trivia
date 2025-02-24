/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ReleaseVersion string = "2.9.4"
)

var (
	bind           string
	colorsFile     string
	exitOnError    bool
	export         bool
	extension      string
	html           bool
	port           uint16
	profile        bool
	recursive      bool
	reload         bool
	reloadInterval string
	verbose        bool
	version        bool
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "trivia",
		Short: "Serves a basic trivia web frontend.",
		Args:  cobra.MinimumNArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return servePage(args)
		},
	}

	rootCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().StringVarP(&colorsFile, "colors", "c", "", "file from which to load color schemes")
	rootCmd.Flags().BoolVar(&exitOnError, "exit-on-error", false, "shut down webserver on error, instead of just printing the error")
	rootCmd.Flags().BoolVar(&export, "export", false, "allow exporting of trivia database")
	rootCmd.Flags().StringVar(&extension, "extension", ".trivia", "only process files ending in this extension (leave empty to match all files)")
	rootCmd.Flags().BoolVar(&html, "html", false, "allow arbitrary html tags in input")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&profile, "profile", false, "register net/http/pprof handlers")
	rootCmd.Flags().BoolVar(&reload, "reload", false, "allow live-reload of questions")
	rootCmd.Flags().StringVar(&reloadInterval, "reload-interval", "", "interval at which to rebuild question list (e.g. \"5m\" or \"1h\")")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse into directories")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "log requests to stdout")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "display version and exit")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("trivia v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) {
	v := viper.New()

	v.SetEnvPrefix("trivia")

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.AutomaticEnv()

	bindFlags(cmd, v)
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "_")

		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
