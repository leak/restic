package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

var RootCmd = cobra.Command{
	Use:   "restic",
	Short: "Backup and restore files",
	Long: `
restic is a backup program which allows saving multiple revisions of files and
directories in an encrypted repository stored on different backends.
`,
	SilenceErrors:     true,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}

func init() {
	cobra.OnInitialize(configureConfigFile)
}

func configureConfigFile() {

	configFile, _ := RootCmd.PersistentFlags().GetString("config-file")

	// configure config file default search paths
	// order: config-file flag, current directory, user folder, global
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("restic")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")

		if runtime.GOOS == "windows" {
			viper.AddConfigPath("%USERPROFILE%/.restic")
			viper.AddConfigPath("%ALLUSERSPROFILE%/restic")
		} else {
			viper.AddConfigPath("$HOME/.restic")
			viper.AddConfigPath("/etc/restic")
		}
	}

	err := viper.ReadInConfig()

	if err != nil {
		if _, isFileNotFoundError := err.(viper.ConfigFileNotFoundError); !isFileNotFoundError {
			fmt.Fprintf(os.Stderr, "Error reading the config file\n", err)
		}
	}
}
