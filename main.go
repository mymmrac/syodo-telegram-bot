package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kataras/golog"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

var (
	configFile       = flag.String("config", "config.toml", "Config file")
	versionRequest   = flag.Bool("version", false, "Version")
	buildInfoRequest = flag.Bool("build-info", false, "Build info")
)

func main() {
	flag.Parse()

	// ==== Build Info ====
	if *buildInfoRequest {
		displayBuildInfo()
		return
	}

	if *versionRequest {
		displayVersion()
		return
	}
	// ==== Build Info End ====

	// ==== Config ====
	cfg, err := config.LoadConfig(*configFile)
	assert(err == nil, fmt.Errorf("load config: %w", err))
	// ==== Config End ====

	// ==== Logger ====
	log := logger.NewLog(golog.New())
	err = cfg.ConfigureLogger(log)
	assert(err == nil, fmt.Errorf("configure logger: %w", err))
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

	bh, err := th.NewBotHandler(bot, updates, th.WithStopTimeout(cfg.Settings.StopTimeout))
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

		err = log.Close()
		assert(err == nil, fmt.Errorf("close logger: %w", err))

		done <- struct{}{}
	}()

	log.Info("Handling updates")
	go bh.Start()

	<-done
	log.Info("Done")
	// ==== Stop Handling End ====
}

func displayBuildInfo() {
	build, ok := debug.ReadBuildInfo()
	assert(ok, "no build info found")

	fmt.Println(build.String())
}

func displayVersion() {
	build, ok := debug.ReadBuildInfo()
	assert(ok, "no build info found")

	var (
		err       error
		commit    string
		buildTime time.Time
		modified  bool
	)

	for _, setting := range build.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			buildTime, err = time.Parse(time.RFC3339, setting.Value)
			assert(err == nil, fmt.Errorf("parse build time: %w", err))
		case "vcs.modified":
			modified, err = strconv.ParseBool(setting.Value)
			assert(err == nil, fmt.Errorf("parse modifed: %w", err))
		}
	}

	fmt.Printf("Syodo Telegram Bot\nCommit: %s (modified :%t)\nBuild Time: %s\n", commit, modified,
		buildTime.Local())
}

func assert(ok bool, args ...any) {
	if !ok {
		fmt.Println(append([]any{"FATAL:"}, args...)...)
		os.Exit(1)
	}
}
