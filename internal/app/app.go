package app

import (
	"context"
	"net/http"
	"time"

	"prservice/internal/config"
	"prservice/internal/usecase"
	"prservice/internal/adapter/repo/postgres"
	httpadapter "prservice/internal/adapter/http"
)

func Run(cfg config.Config) error {
	// Инициализация БД (Postgres адаптер)
	db, err := postgres.New(cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer db.Close(context.Background())

	teamRepo := postgres.NewTeamRepo(db)
	userRepo := postgres.NewUserRepo(db)
	prRepo := postgres.NewPRRepo(db)

	// Usecases
	teamSvc := usecase.NewTeamService(teamRepo, userRepo)
	userSvc := usecase.NewUserService(userRepo)
	prSvc := usecase.NewPRService(prRepo, userRepo)

	// HTTP сервер (оapi-codegen router подключим в adapter/http)
	server := httpadapter.NewServer(teamSvc, userSvc, prSvc, prRepo)
	router := httpadapter.NewRouter(server)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return srv.ListenAndServe()
}
