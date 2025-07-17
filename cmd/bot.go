package main

import (
	"KinopoiskTwoActors/configs"
	"KinopoiskTwoActors/configs/loader/dotEnvLoader"
	"KinopoiskTwoActors/internal/delivery/telegram"
	"KinopoiskTwoActors/internal/repository/kinopoisk"
	"KinopoiskTwoActors/internal/repository/userState"
	"KinopoiskTwoActors/internal/usecase"
	"KinopoiskTwoActors/pkg/logger"
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	loader := dotEnvLoader.DotEnvLoader{}
	cfg := configs.MustLoad(loader)
	log := logger.NewLogger(cfg)

	repo := kinopoisk.NewRepo(cfg)
	useCase := usecase.NewActorFilmRepository(repo)
	userStates := userState.NewUserStates()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
	log.Info("Starting prometheus at port 8080")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := telegram.NewBot(ctx, cfg, userStates, useCase, log)
	if err != nil {
		log.Error("failed to create bot:", err)
		os.Exit(1)
	}
	log.Info("Starting bot")
	go bot.Run()
	<-done
	log.Info("Shutting down bot")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bot.Stop(ctx)
	log.Info("Service stopped")
}
