package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fiws/minepkg/internals/cmdlog"
	"github.com/fiws/minepkg/pkg/api"
	"github.com/gookit/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MinepkgVersion is a constant of the current minepkg version
const MinepkgVersion = "0.0.26"

// TODO: this logger is not so great – also: it should not be global
var logger *cmdlog.Logger = cmdlog.New()

var (
	cfgFile       string
	globalDir     = "/tmp"
	apiClient     = api.New()
	loginData     = &api.AuthResponse{}
	disableColors bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: MinepkgVersion,
	Use:     "minepkg",
	Short:   "Minepkg at your service.",
	Long:    "Manage Minecraft mods with ease",

	Example: `
  minepkg init -l fabric
  minepkg install modmenu@latest
  minepkg install https://minepkg.io/projects/desire-paths`,
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Args:  cobra.MaximumNArgs(1),
	Short: "Output shell completion code for bash",
	Long: `To load completion run

. <(minepkg completion)

You can add that line to your ~/.bashrc or ~/.profile to
persist completion in your shell.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	globalDir = filepath.Join(home, ".minepkg")

	// check if user is logged in
	if rawCreds, err := ioutil.ReadFile(filepath.Join(globalDir, "credentials.json")); err == nil {
		if err := json.Unmarshal(rawCreds, &loginData); err == nil && loginData.Token != "" {
			apiClient.JWT = loginData.Token
			apiClient.User = loginData.User
		}
	}

	token := os.Getenv("MINEPKG_API_TOKEN")
	if token != "" {
		apiClient.JWT = token
		fmt.Println("Using MINEPKG_API_TOKEN for authentication")
	}

	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&disableColors, "no-color", "", false, "disable color output")
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.minepkg/config.toml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if disableColors == true {
		color.Disable()
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".minepkg" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".minepkg")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
