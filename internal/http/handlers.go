package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Debian-WNPP-Reloaded/wnpp-backend-service/internal/repository"
)

type Handler struct {
	wnppRepo *repository.WNPPRepository
}

var validWNPPTypes = map[string]struct{}{
	"ITP": {},
	"RFP": {},
	"O":   {},
	"ITA": {},
	"RFH": {},
	"RFA": {},
}

func filterValidTypes(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	out := make([]string, 0, len(input))
	for _, t := range input {
		if _, ok := validWNPPTypes[t]; !ok {
			return nil, fmt.Errorf("invalid type: %s", t)
		}
		out = append(out, t)
	}
	return out, nil
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

	// ---- type filter (multi)
	rawTypes := r.URL.Query()["type"]
	types, err := filterValidTypes(rawTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---- owner filter (tri-state)
	ownerParam := r.URL.Query().Get("owner")
	var owner *bool
	if ownerParam != "" {
		switch ownerParam {
		case "true":
			v := true
			owner = &v
		case "false":
			v := false
			owner = &v
		default:
			http.Error(w, "invalid owner parameter (must be true or false)", http.StatusBadRequest)
			return
		}
	}

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

	items, err := h.wnppRepo.List(
		ctx,
		limit,
		offset,
		order,
		types,
		search,
		owner,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h *Handler) WNPPCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawTypes := r.URL.Query()["type"]
	types, err := filterValidTypes(rawTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---- owner filter (tri-state)
	ownerParam := r.URL.Query().Get("owner")
	var owner *bool
	if ownerParam != "" {
		switch ownerParam {
		case "true":
			v := true
			owner = &v
		case "false":
			v := false
			owner = &v
		default:
			http.Error(w, "invalid owner parameter (must be true or false)", http.StatusBadRequest)
			return
		}
	}

	search := r.URL.Query().Get("q")

	total, err := h.wnppRepo.Count(ctx, types, search, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"total": total,
	})
}
