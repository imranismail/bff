package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/imranismail/bff/config"
	"github.com/imranismail/bff/log"
	"github.com/imranismail/bff/proxy"
	"github.com/spf13/cobra"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

var cfgFile string
var cfgPath = path.Join(xdg.ConfigHome, "bff")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bff",
	Short: "Backend for frontend",
	Long: `
An API-aware proxy that is cabaple of routing, filtering, verifying
and modifing HTTP request and response.`,
	Run:     proxy.Serve,
	Version: config.Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Root: %v", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $XDG_CONFIG_HOME/bff/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("port", "p", "5000", "Port to run the server on")
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))

	rootCmd.Flags().BoolP("insecure", "i", false, "Skip TLS verify")
	viper.BindPFlag("insecure", rootCmd.Flags().Lookup("insecure"))

	rootCmd.Flags().IntP("verbosity", "v", 2, "Verbosity")
	viper.BindPFlag("verbosity", rootCmd.Flags().Lookup("verbosity"))

	rootCmd.Flags().BoolP("pretty", "r", false, "Pretty logs")
	viper.BindPFlag("pretty", rootCmd.Flags().Lookup("pretty"))

	rootCmd.Flags().StringP("url", "u", "", "Proxied URL")
	viper.BindPFlag("url", rootCmd.Flags().Lookup("url"))

	if hasPipedInput() {
		b, err := ioutil.ReadAll(os.Stdin)

		if err == nil {
			viper.Set("modifiers", b)
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(cfgPath)
		viper.AddConfigPath(".")
	}

	// read in environment variables that match
	viper.SetEnvPrefix("bff")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	// If a config file is found, read it in
	if err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}

	log.Configure()

	viper.OnConfigChange(func(evt fsnotify.Event) {
		log.Infof("Reconfiguring: %v", evt.Name)
		log.Configure()
		proxy.Configure()
	})

	viper.WatchConfig()
}

func hasPipedInput() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice == 0 && fi.Size() > 0
}
