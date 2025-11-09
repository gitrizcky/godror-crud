package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "go-demo-crud/internal/api"
    "go-demo-crud/internal/db"
    "go-demo-crud/internal/docs"
    appconfig "go-demo-crud/internal/config"
    "go-demo-crud/internal/repo"
)

func main() {
    dbCfg, err := appconfig.LoadDBConfig("config/application.properties")
    if err != nil {
        log.Fatalf("load DB config failed: %v", err)
    }

    database, err := db.Open(dbCfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer func() { _ = database.Close() }()

    // Hot-reloadable SQL queries from properties file
    qman := appconfig.NewManager("config/application.properties")
	stop := make(chan struct{})
	go qman.Start(stop)

	pr := repo.NewProductRepo(database, qman)
	srv := api.NewServer(pr)

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	mux.Handle("/docs/", http.StripPrefix("/docs/", docs.Handler()))

	// Gracefully handle stop signals to stop the config watcher
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		close(stop)
	}()

	log.Println("CRUD API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
