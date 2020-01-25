package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Sigafoos/pvpservice/handler"
	"github.com/Sigafoos/pvpservice/pvp"

	"github.com/NYTimes/gziphandler"
	"github.com/Sigafoos/middleware"
	"github.com/Sigafoos/middleware/logger"
	"github.com/gocraft/dbr"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("no dsn found")
	}

	db, err := dbr.Open("postgres", dsn, nil)
	if err != nil {
		log.Panicln("cannot open database: " + err.Error())
	}
	defer db.Close()

	pvp, err := pvp.New(db, logger.New(os.Stderr))
	if err != nil {
		log.Println(err)
		return
	}

	h := handler.New(pvp)
	mux := http.NewServeMux()
	mux.Handle("/register", http.HandlerFunc(h.Register))
	mux.Handle("/list", http.HandlerFunc(h.List))

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
}
