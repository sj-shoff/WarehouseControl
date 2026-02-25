package items_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"
	"warehouse-control/internal/http-server/handler/items/dto"
	"warehouse-control/internal/http-server/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type ItemsHandler struct {
	itemsUsecase itemsUsecase
	logger       *zlog.Zerolog
}

func NewHandler(itemsUsecase itemsUsecase, logger *zlog.Zerolog) *ItemsHandler {
	return &ItemsHandler{
		itemsUsecase: itemsUsecase,
		logger:       logger,
	}
}

func (h *ItemsHandler) writeError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := "internal_error"
	switch {
	case errors.Is(err, customErr.ErrInvalidInput):
		code = http.StatusBadRequest
		msg = "invalid_input"
	case errors.Is(err, customErr.ErrItemNotFound):
		code = http.StatusNotFound
		msg = "not_found"
	case errors.Is(err, customErr.ErrForbidden):
		code = http.StatusForbidden
		msg = "forbidden"
	case errors.Is(err, customErr.ErrDatabase):
		code = http.StatusInternalServerError
		msg = "database_error"
	}
	http.Error(w, msg, code)
}

func (h *ItemsHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, customErr.ErrUnauthorized)
		return
	}
	var req dto.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item := &domain.Item{
		Name:     req.Name,
		SKU:      req.SKU,
		Quantity: req.Quantity,
		Price:    req.Price,
		Category: req.Category,
		Location: req.Location,
	}
	id, err := h.itemsUsecase.CreateItem(r.Context(), item, claims.Username)
	if err != nil {
		h.logger.Error().Err(err).Msg("CreateItem failed")
		h.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item created")
}

func (h *ItemsHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.itemsUsecase.GetItems(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItems failed")
		h.writeError(w, err)
		return
	}
	resp := dto.ItemsResponse{
		Items: make([]*dto.ItemResponse, len(items)),
	}
	for i, item := range items {
		resp.Items[i] = h.toItemResponse(item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ItemsHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.toItemResponse(item))
}

func (h *ItemsHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, customErr.ErrUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	var req dto.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.SKU != "" {
		item.SKU = req.SKU
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Location != "" {
		item.Location = req.Location
	}
	err = h.itemsUsecase.UpdateItem(r.Context(), id, item, claims.Username)
	if err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item updated")
}

func (h *ItemsHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r)
	if claims == nil {
		h.writeError(w, customErr.ErrUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	err = h.itemsUsecase.DeleteItem(r.Context(), id, claims.Username)
	if err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item deleted")
}

func (h *ItemsHandler) toItemResponse(item *domain.Item) *dto.ItemResponse {
	return &dto.ItemResponse{
		ID:        item.ID,
		Name:      item.Name,
		SKU:       item.SKU,
		Quantity:  item.Quantity,
		Price:     item.Price,
		Category:  item.Category,
		Location:  item.Location,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
