package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	dbpkg "github.com/Debian-WNPP-Reloaded/wnpp-backend-service/internal/db"
	httppkg "github.com/Debian-WNPP-Reloaded/wnpp-backend-service/internal/http"
	"github.com/Debian-WNPP-Reloaded/wnpp-backend-service/internal/repository"

	"github.com/Debian-WNPP-Reloaded/wnpp-backend-service/internal/middleware"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	databaseUrl := os.Getenv("DATABASE_URL")

	if databaseUrl == "" {
		databaseUrl = "postgresql://udd-mirror:udd-mirror@udd-mirror.debian.net:5432/udd"
	}

	db, err := dbpkg.New(ctx, databaseUrl)
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

	corsMux := middleware.CORS(mux)

	log.Println("Server running on :8080")
	err2 := http.ListenAndServe(":8080", corsMux)
	if err2 != nil {
		log.Fatal(err2)
	}

	//log.Println("Listening on :8080")
	//log.Fatal(http.ListenAndServe(":8080", mux))
}
