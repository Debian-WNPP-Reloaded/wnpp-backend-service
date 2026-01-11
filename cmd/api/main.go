package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	dbpkg "github.com/GabrielBarrantes/wnpp-backend-service/internal/db"
	httppkg "github.com/GabrielBarrantes/wnpp-backend-service/internal/http"
	"github.com/GabrielBarrantes/wnpp-backend-service/internal/repository"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := dbpkg.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewWNPPRepository(db)
	handler := httppkg.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/wnpp", handler.WNPP)
	mux.HandleFunc("/api/wnpp/count", handler.WNPPCount)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
