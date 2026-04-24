package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	db, err := InitDB("./series.db")
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	app := &App{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.healthHandler)
	mux.HandleFunc("/series", app.seriesCollectionHandler)
	mux.HandleFunc("/series/", app.seriesItemHandler)
	mux.HandleFunc("/swagger.json", app.swaggerJSONHandler)
	mux.HandleFunc("/docs", app.swaggerUIHandler)

	// 👇 CAMBIO IMPORTANTE
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Backend running on port", port)

	if err := http.ListenAndServe(":"+port, app.withCORS(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
