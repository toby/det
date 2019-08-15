package command

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toby/det/server"
)

var rootCmd = &cobra.Command{
	Use:   "det",
	Short: "det - P2P search and discovery on the BitTorrent network",
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetDefault("ListenHost", "")
	viper.SetDefault("ListenPort", 42069)
	viper.SetDefault("PublicHost", "")
	viper.SetDefault("DisableUpnp", false)
	viper.SetDefault("HashQueueLength", 500)
	viper.SetDefault("SqlitePath", "./")
	viper.SetDefault("BoltDBPath", "./")
	viper.SetDefault("DownloadPath", "./")
	viper.SetDefault("Listen", true)
	viper.SetDefault("Seed", false)
	viper.SetDefault("NumResolvers", 5)
	viper.SetDefault("ResolverTimeout", time.Second*30)
	viper.SetDefault("ResolverWindow", time.Minute*10)
	viper.SetDefault("TorrentDebug", false)
}

func serverConfigFromDefaults() *server.Config {
	cfg := &server.Config{}
	cfg.ListenHost = viper.GetString("ListenHost")
	cfg.ListenPort = viper.GetInt("ListenPort")
	cfg.DisableUpnp = viper.GetBool("DisableUpnp")
	cfg.HashQueueLength = viper.GetInt("HashQueueLength")
	cfg.SqlitePath = viper.GetString("SqlitePath")
	cfg.BoltDBPath = viper.GetString("BoltDBPath")
	cfg.DownloadPath = viper.GetString("DownloadPath")
	cfg.Listen = viper.GetBool("Listen")
	cfg.Seed = viper.GetBool("Seed")
	cfg.NumResolvers = viper.GetInt("NumResolvers")
	cfg.ResolverTimeout = viper.GetDuration("ResolverTimeout")
	cfg.ResolverWindow = viper.GetDuration("ResolverWindow")
	cfg.TorrentDebug = viper.GetBool("TorrentDebug")
	cfg.PublicHost = viper.GetString("PublicHost")
	return cfg
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/det/")
	viper.AddConfigPath("$HOME/.det")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("det")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
