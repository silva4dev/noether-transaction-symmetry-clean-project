package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/infra/db"
	httphandler "github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/infra/http"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/infra/kafka/consumer"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/infra/repository"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/pkg/uow"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ctx := context.Background()
	dtb, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/cartola?parseTime=true")

	if err != nil {
		panic(err)
	}
	defer dtb.Close()
	uow, err := uow.NewUow(ctx, dtb)
	if err != nil {
		panic(err)
	}
	registerRepositories(uow)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Get("/players", httphandler.ListPlayersHandler(ctx, *db.New(dtb)))
	r.Get("/my-teams/{teamID}/players", httphandler.ListMyTeamPlayersHandler(ctx, *db.New(dtb)))
	r.Get("/my-teams/{teamID}/balance", httphandler.GetMyTeamBalanceHandler(ctx, *db.New(dtb)))
	r.Get("/matches", httphandler.ListMatchesHandler(ctx, repository.NewMatchRepository(dtb)))
	r.Get("/matches/{matchID}", httphandler.ListMatchByIDHandler(ctx, repository.NewMatchRepository(dtb)))

	go http.ListenAndServe(":8080", r)

	var topics = []string{"newMatch", "chooseTeam", "newPlayer", "matchUpdateResult", "newAction"}
	msgChan := make(chan *kafka.Message)
	go consumer.Consume(topics, "host.docker.internal:9094", msgChan)
	consumer.ProcessEvents(ctx, msgChan, uow)
}

func registerRepositories(uow *uow.Uow) {
	uow.Register("PlayerRepository", func(tx *sql.Tx) interface{} {
		repo := repository.NewPlayerRepository(uow.Db)
		repo.Queries = db.New(tx)
		return repo
	})

	uow.Register("MatchRepository", func(tx *sql.Tx) interface{} {
		repo := repository.NewMatchRepository(uow.Db)
		repo.Queries = db.New(tx)
		return repo
	})

	uow.Register("TeamRepository", func(tx *sql.Tx) interface{} {
		repo := repository.NewTeamRepository(uow.Db)
		repo.Queries = db.New(tx)
		return repo
	})

	uow.Register("MyTeamRepository", func(tx *sql.Tx) interface{} {
		repo := repository.NewMyTeamRepository(uow.Db)
		repo.Queries = db.New(tx)
		return repo
	})
}
