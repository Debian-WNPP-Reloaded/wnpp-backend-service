package http

import (
	"encoding/json"
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

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) WNPP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	offset := 0
	order := r.URL.Query().Get("order")

	if v := r.URL.Query().Get("limit"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			limit = i
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			offset = i
		}
	}

	items, err := h.wnppRepo.List(ctx, limit, offset, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) WNPPCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	total, err := h.wnppRepo.Count(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"total": total,
	})
}