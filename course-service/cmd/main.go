package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/BON4/elearn-demo/course-service/internal/app"
	"github.com/BON4/elearn-demo/course-service/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	var configPath string
	pflag.StringVarP(&configPath, "config", "c", "", "path to ini config file")
	pflag.Parse()

	var (
		cfg *config.Config
		err error
	)

	if configPath != "" {
		cfg, err = config.LoadFromINI(configPath)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	} else {
		cfg, err = config.LoadFromEnv()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
