package app

import (
	"fmt"
	"os"
	srv "webtool/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string
	opt     *Option
	rootCmd = &cobra.Command{
		Use:   "webtool",
		Short: "echo和waf test功能",
		Long: `echo: 用fasthttp实现的http echo功能
waf: 用tcp实现的nginx配套waf测试功能`,
	}
)

type Option struct {
	ConfigPath string
	Codes      srv.ResponseCodes
}

func (o *Option) Config() *Config {
	cfg := &Config{}
	return cfg
}

func NewOption() *Option {
	opt := new(Option)
	opt.ConfigPath = os.Getenv(PROGRAM_NAME + "_CONFIG")
	if opt.ConfigPath == "" {
		opt.ConfigPath = "config/webtool.yaml"
	}
	return opt
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	opt = NewOption()

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", opt.ConfigPath, "config file (default is config/webtool.yaml)")
	rootCmd.PersistentFlags().IntVarP(&opt.Codes.RequestCode, "requestcode", "r", 200, "request's response code")
	rootCmd.PersistentFlags().IntVarP(&opt.Codes.ResponseCode, "responsecode", "w", 200, "response's response code")
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")

	rootCmd.AddCommand(NewHttpCmd())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("webtool")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Open file:", cfgFile, err.Error())
		os.Exit(0)
	}
}
