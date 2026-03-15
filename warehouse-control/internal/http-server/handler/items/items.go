package items_handler

import (
	"errors"
	"net/http"
	"strconv"

	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"
	"warehouse-control/internal/http-server/handler/items/dto"
	"warehouse-control/internal/http-server/middleware"

	"github.com/gin-gonic/gin"
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

func (h *ItemsHandler) CreateItem(c *gin.Context) {
	claims := middleware.GetClaimsFromContext(c)
	if claims == nil {
		h.writeError(c, customErr.ErrUnauthorized)
		return
	}
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		h.writeError(c, customErr.ErrInvalidInput)
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
	id, err := h.itemsUsecase.CreateItem(c.Request.Context(), item, claims.Username)
	if err != nil {
		h.logger.Error().Err(err).Msg("CreateItem failed")
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item created")
}

func (h *ItemsHandler) GetItems(c *gin.Context) {
	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 100
	}
	offsetStr := c.Query("offset")
	offset, _ := strconv.Atoi(offsetStr)
	search := c.Query("search")
	items, total, err := h.itemsUsecase.GetItems(c.Request.Context(), limit, offset, search)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItems failed")
		h.writeError(c, err)
		return
	}
	resp := dto.ItemsResponse{
		Items: make([]*dto.ItemResponse, len(items)),
		Total: total,
	}
	for i, item := range items {
		resp.Items[i] = dto.ToItemResponse(item)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ItemsHandler) GetItemByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToItemResponse(item))
}

func (h *ItemsHandler) UpdateItem(c *gin.Context) {
	claims := middleware.GetClaimsFromContext(c)
	if claims == nil {
		h.writeError(c, customErr.ErrUnauthorized)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
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
	err = h.itemsUsecase.UpdateItem(c.Request.Context(), id, item, claims.Username)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item updated")
}

func (h *ItemsHandler) DeleteItem(c *gin.Context) {
	claims := middleware.GetClaimsFromContext(c)
	if claims == nil {
		h.writeError(c, customErr.ErrUnauthorized)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	err = h.itemsUsecase.DeleteItem(c.Request.Context(), id, claims.Username)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	h.logger.Info().Int64("id", id).Str("user", claims.Username).Msg("Item deleted")
}

func (h *ItemsHandler) BulkDeleteItems(c *gin.Context) {
	claims := middleware.GetClaimsFromContext(c)
	if claims == nil {
		h.writeError(c, customErr.ErrUnauthorized)
		return
	}
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	err := h.itemsUsecase.BulkDeleteItems(c.Request.Context(), req.IDs, claims.Username)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	h.logger.Info().Str("user", claims.Username).Msg("Items bulk deleted")
}

func (h *ItemsHandler) writeError(c *gin.Context, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, customErr.ErrInvalidInput):
		code = http.StatusBadRequest
	case errors.Is(err, customErr.ErrItemNotFound):
		code = http.StatusNotFound
	case errors.Is(err, customErr.ErrForbidden):
		code = http.StatusForbidden
	case errors.Is(err, customErr.ErrDatabase):
		code = http.StatusInternalServerError
	case errors.Is(err, customErr.ErrInternal):
		code = http.StatusInternalServerError
	}
	c.JSON(code, gin.H{"error": err.Error()})
}
