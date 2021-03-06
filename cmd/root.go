package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwalton/gchalk"
	"github.com/minepkg/minepkg/cmd/bump"
	"github.com/minepkg/minepkg/cmd/config"
	"github.com/minepkg/minepkg/cmd/dev"
	"github.com/minepkg/minepkg/cmd/initCmd"
	"github.com/minepkg/minepkg/internals/commands"
	"github.com/minepkg/minepkg/internals/credentials"
	"github.com/minepkg/minepkg/internals/globals"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: this logger is not so great – also: it should not be global
var logger = globals.Logger

var (
	cfgFile   string
	globalDir = "/tmp"

	// Version is the current version. it should be set by goreleaser
	Version string
	// commit is also set by goreleaser (in main.go)
	Commit string
	// nextVersion is a placeholder version. only used for local dev
	nextVersion string = "0.1.0-dev-local"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	// Version gets set dynamically
	Use:   "minepkg",
	Short: "Minepkg at your service.",
	Long:  "Manage Minecraft mods with ease",

	Example: `
  minepkg init -l fabric
  minepkg install modmenu@latest
  minepkg join demo.minepkg.host`,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	initRoot()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(commands.ErrorBox(err.Error(), ""))
		os.Exit(1)
	}
}

func initRoot() {
	// include commit if this is next version
	if strings.HasSuffix(Version, "-next") {
		Version = Version + "+" + Commit
	}
	rootCmd.Version = Version
	if rootCmd.Version == "" {
		rootCmd.Version = nextVersion
	}
	homeConfigs, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	apiKey := os.Getenv("MINEPKG_API_KEY")
	globalDir = filepath.Join(homeConfigs, "minepkg")
	credStore, err := credentials.New(globalDir, globals.ApiClient.APIUrl)
	if err != nil {
		if apiKey != "" {
			logger.Warn("Could not initialize credential store: " + err.Error())
		} else {
			logger.Fail("Could not initialize credential store: " + err.Error())
		}
	}
	globals.CredStore = credStore

	if credStore.MinepkgAuth != nil {
		globals.ApiClient.JWT = credStore.MinepkgAuth.AccessToken
	}

	if apiKey != "" {
		globals.ApiClient.APIKey = apiKey
		fmt.Println("Using MINEPKG_API_KEY for authentication")
	}

	cobra.OnInitialize(initConfig)

	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configPath := filepath.Join(configDir, "minepkg")

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s/config.toml)", configPath))
	rootCmd.PersistentFlags().BoolP("accept-minecraft-eula", "a", false, "Accept Minecraft's eula. See https://www.minecraft.net/en-us/eula/")
	// rootCmd.PersistentFlags().BoolP("system-java", "", false, "Use system java instead of internal installation for launching Minecraft server or client")
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "More verbose logging. Not really implemented yet")
	rootCmd.PersistentFlags().BoolP("non-interactive", "", false, "Do not prompt for anything (use defaults instead)")

	viper.BindPFlag("useSystemJava", rootCmd.PersistentFlags().Lookup("system-java"))
	viper.BindPFlag("acceptMinecraftEula", rootCmd.PersistentFlags().Lookup("accept-minecraft-eula"))
	viper.BindPFlag("verboseLogging", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("nonInteractive", rootCmd.PersistentFlags().Lookup("non-interactive"))

	// viper.SetDefault("init.defaultSource", "https://github.com/")

	// subcommands
	rootCmd.AddCommand(dev.SubCmd)
	rootCmd.AddCommand(config.SubCmd)
	rootCmd.AddCommand(initCmd.New())
	rootCmd.AddCommand(bump.New())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if viper.GetBool("noColor") || os.Getenv("CI") != "" {
		gchalk.ForceLevel(gchalk.LevelNone)
		viper.Set("nonInteractive", true)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath("/etc/minepkg/")                     // path to look for the config file in
		viper.AddConfigPath(filepath.Join(configDir, "minepkg")) // call multiple times to add many search paths
		viper.AddConfigPath(".")                                 // optionally look for config in the working directory
	}

	viper.SetEnvPrefix("MINEPKG")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verboseLogging") {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if viper.GetString("apiUrl") != "" {
		logger.Warn("NOT using default minepkg API URL: " + viper.GetString("apiUrl"))
		globals.ApiClient.APIUrl = viper.GetString("apiUrl")
	}
}
