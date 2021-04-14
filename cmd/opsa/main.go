package main

import (
	"log"
	"os"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/na4ma4/config"
	"github.com/na4ma4/op-secret-automation-test/internal/mainconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//nolint: gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use:  "opsa",
	Args: cobra.NoArgs,
	Run:  mainCommand,
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	rootCmd.PersistentFlags().StringP("token", "t", "", "Connect Token")
	rootCmd.PersistentFlags().String("host", "", "Connect Host")

	_ = viper.BindPFlag("test.token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("test.host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindEnv("test.token", "OP_TOKEN")
	_ = viper.BindEnv("test.host", "OP_HOST")
}

func main() {
	_ = rootCmd.Execute()
}

func zapEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func zapConfig() zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zapEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// MySQL Home User
type DBConfig struct {
	URL      string           `opitem:"MySQL Home User" opfield:".urls"`
	Username string           `opitem:"MySQL Home User" opfield:".username"`
	Password string           `opitem:"MySQL Home User" opfield:".password"`
	APIKey   onepassword.Item `opitem:"MySQL Home User"`
}

func mainCommand(cmd *cobra.Command, args []string) {
	cfg := config.NewViperConfigFromViper(viper.GetViper(), "opsa")

	logger, _ := zapConfig().Build()
	defer logger.Sync() //nolint: errcheck

	client := connect.NewClient(cfg.GetString("test.host"), cfg.GetString("test.token"))
	vaults, err := client.GetVaults()
	if err != nil {
		log.Fatalf("error getting vaults: %s", err)
	}
	for _, vault := range vaults {
		log.Printf("Vault: %#v", vault)
		os.Setenv("OP_VAULT", vault.ID)
	}
	dbConf := &DBConfig{}

	if len(vaults) < 0 {
		log.Fatalf("error vault list too short: %s", err)
	}

	opItem, err := client.GetItemByTitle("MySQL Home User", vaults[0].ID)
	if err != nil {
		log.Fatalf("error getting item: %s", err)
	}
	log.Printf("OP Item: %#v", opItem.URLs)

	if err := connect.Load(client, dbConf); err != nil {
		log.Fatalf("error loading struct: %s", err)
	}

	log.Printf("DB Config: %#v", dbConf)
}
