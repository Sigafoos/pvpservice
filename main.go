package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sigafoos/pvpservice/handler"
	"github.com/Sigafoos/pvpservice/pvp"

	"github.com/NYTimes/gziphandler"
	"github.com/Sigafoos/middleware"
	"github.com/Sigafoos/middleware/logger"
	"github.com/gocraft/dbr/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	log := logger.New(os.Stdout)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("no dsn found")
	}
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		log.Printf("error parsing dsn: %s", err.Error())
		return
	}

	db, err := dbr.Open("postgres", parsedDSN.String(), nil)
	if err != nil {
		log.Panicln("cannot open database: " + err.Error())
	}
	defer db.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("error getting working directory: %s", err.Error())
		return
	}

	migrationTable := os.Getenv("MIGRATION_TABLE")
	if migrationTable == "" {
		migrationTable = "migrations_pvp"
	}
	parsedDSN.RawQuery += "&x-migrations-table=" + migrationTable

	m, err := migrate.New(fmt.Sprintf("file:///%s/migrations", cwd), parsedDSN.String())
	if err != nil {
		log.Printf("error creating migration driver: %s", err.Error())
		return
	}
	m.Log = log

	// Up() returns an error if there are no migrations
	_ = m.Up()

	// we don't need the error or the database reference
	_, _ = m.Close()

	pvp := pvp.New(db, log)

	// TODO pass log here as well
	h := handler.New(pvp)
	mux := http.NewServeMux()
	mux.Handle("/register", http.HandlerFunc(h.Register))
	mux.Handle("/player", http.HandlerFunc(h.Player))
	mux.Handle("/player/list", http.HandlerFunc(h.List))

	chain := gziphandler.GzipHandler(mux)
	chain = middleware.UseJSON(chain)
	chain = middleware.UseAuth(chain)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}
	log.Println("server running on port " + port)
	log.Println(server.ListenAndServe())

	// gracefully shut down
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
