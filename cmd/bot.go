package main

import (
	"KinopoiskTwoActors/configs"
	"KinopoiskTwoActors/configs/loader/dotEnvLoader"
	"KinopoiskTwoActors/internal/delivery/telegram"
	"KinopoiskTwoActors/internal/repository/SessionStates"
	"KinopoiskTwoActors/internal/repository/cachedRepo"
	"KinopoiskTwoActors/internal/repository/kinopoisk"
	"KinopoiskTwoActors/internal/repository/redisCache"
	"KinopoiskTwoActors/internal/usecase"
	"KinopoiskTwoActors/pkg/logger"
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	//wd, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//f, err := os.Create(filepath.Join(wd, "0004_profiling/01_collect_profile/cpu.out"))
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//defer f.Close()
	//
	//if err = pprof.StartCPUProfile(f); err != nil {
	//	log.Fatal(err)
	//}
	//
	//defer pprof.StopCPUProfile()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	loader := dotEnvLoader.DotEnvLoader{}
	cfg := configs.MustLoad(loader)
	log := logger.NewLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := kinopoisk.NewRepo(cfg)
	cache, err := redisCache.NewCache(ctx, cfg, "movie:", log)
	var actor telegram.ActorProvider
	var film telegram.FilmProvider

	if err == nil {
		cachedRepo := cachedRepo.NewCachedRepo(repo, cache, log)
		actor = usecase.NewActor(cachedRepo)
		film = usecase.NewFilm(cachedRepo)
	} else {
		actor = usecase.NewActor(repo)
		film = usecase.NewFilm(repo)
	}

	states := SessionStates.NewUserStates()

	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: promhttp.Handler(),
	}
	go func() {
		log.Info("Запуск prometheus на порту 8080")
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error: ", err)
			os.Exit(1)
		}
	}()

	bot, err := telegram.NewBot(cfg, states, actor, film, log)
	if err != nil {
		log.Error("ошибка при создании бота: ", err)
		os.Exit(1)
	}
	log.Info("Запуск бота")

	go bot.Run(ctx)

	<-done
	gracefulShutdown(ctx, httpSrv, bot, log)

}

func gracefulShutdown(parentCtx context.Context, httpSrv *http.Server, bot *telegram.Bot, log *slog.Logger) {
	log.Info("Остановка сервисов")

	shutdownCtx, shutdownCancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer shutdownCancel()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.Stop(shutdownCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Error("Ошибка остановки HTTP сервера: ", err)
		}
	}()

	completed := make(chan struct{})

	go func() {
		wg.Wait()
		close(completed)
	}()

	select {
	case <-completed:
		log.Info("Все сервисы корректно остановлены")
	case <-shutdownCtx.Done():
		log.Info("Таймаут заверешения работы превышен, принудительная остановка")
	}
}
