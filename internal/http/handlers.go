package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/GabrielBarrantes/wnpp-backend-service/internal/repository"
)

type Handler struct {
	wnppRepo *repository.WNPPRepository
}

func NewHandler(wnppRepo *repository.WNPPRepository) *Handler {
	return &Handler{wnppRepo: wnppRepo}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) WNPP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	types := r.URL.Query()["type"] // <-- MULTI TYPE
	search := r.URL.Query().Get("q")
	order := r.URL.Query().Get("order")

	log.Println("search:", search)

	limit := 50
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			limit = i
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			offset = i
		}
	}

	items, err := h.wnppRepo.List(ctx, limit, offset, order, types, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h *Handler) WNPPCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	types := r.URL.Query()["type"]
	search := r.URL.Query().Get("q")

	total, err := h.wnppRepo.Count(ctx, types, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"total": total,
	})
}
