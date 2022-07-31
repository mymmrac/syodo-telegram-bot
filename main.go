package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kataras/golog"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

var configFile = flag.String("config", "config.toml", "Config file")

func main() {
	// ==== Config ====
	flag.Parse()
	cfg, err := config.LoadConfig(*configFile)
	assert(err == nil, fmt.Errorf("load config: %w", err))
	// ==== Config End ====

	// ==== Logger ====
	log := logger.NewLog(golog.New())
	err = cfg.ConfigureLogger(log)
	assert(err == nil, fmt.Errorf("configure logger: %w", err))
	defer func() {
		err = log.Close()
		assert(err == nil, fmt.Errorf("close logger: %w", err))
	}()
	// ==== Logger End ====

	// ==== Bot Setup ====
	log.Info("Setting up")

	bot, err := telego.NewBot(cfg.Settings.BotToken, telego.WithLogger(log), telego.WithHealthCheck())
	if err != nil {
		log.Fatalf("Create bot: %s", err)
	}

	updates, err := bot.UpdatesViaLongPulling(nil)
	if err != nil {
		log.Fatalf("Get updates: %s", err)
	}

	bh, err := th.NewBotHandler(bot, updates, th.WithStopTimeout(time.Second*10))
	if err != nil {
		log.Fatalf("Create bot handler: %s", err)
	}

	handler := NewHandler(cfg, log, bh)
	handler.RegisterHandlers()
	// ==== Bot Setup End ====

	// ==== Stop Handling ====
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{}, 1)

	go func() {
		<-sigs
		log.Info("Stopping")

		bot.StopLongPulling()
		bh.Stop()

		done <- struct{}{}
	}()

	log.Info("Handling updates")
	go bh.Start()

	<-done
	log.Info("Done")
	// ==== Stop Handling End ====
}

func assert(ok bool, args ...any) {
	if !ok {
		fmt.Println(append([]any{"FATAL:"}, args...)...)
		os.Exit(1)
	}
}
