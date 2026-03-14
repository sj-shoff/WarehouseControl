package history_handler

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"warehouse-control/internal/domain"
	"warehouse-control/internal/http-server/handler/history/dto"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

type HistoryHandler struct {
	historyUsecase historyUsecase
	logger         *zlog.Zerolog
}

func NewHandler(historyUsecase historyUsecase, logger *zlog.Zerolog) *HistoryHandler {
	return &HistoryHandler{
		historyUsecase: historyUsecase,
		logger:         logger,
	}
}

func (h *HistoryHandler) GetHistory(c *gin.Context) {
	filter := domain.HistoryFilter{
		Limit:  100,
		Offset: 0,
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}
	if itemID := c.Query("item_id"); itemID != "" {
		if id, err := strconv.ParseInt(itemID, 10, 64); err == nil && id > 0 {
			filter.ItemID = &id
		}
	}
	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}
	if username := c.Query("username"); username != "" {
		filter.Username = &username
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &t
		}
	}
	records, err := h.historyUsecase.GetHistory(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetHistory failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
		Total:   len(records),
	}
	for i, rec := range records {
		resp.Records[i] = h.toHistoryRecordResponse(rec)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HistoryHandler) GetItemHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input"})
		return
	}
	records, err := h.historyUsecase.GetHistoryByItemID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItemHistory failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
	}
	for i, rec := range records {
		resp.Records[i] = h.toHistoryRecordResponse(rec)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HistoryHandler) ExportHistoryCSV(c *gin.Context) {
	filter := domain.HistoryFilter{
		Limit:  1000,
		Offset: 0,
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}
	if itemID := c.Query("item_id"); itemID != "" {
		if id, err := strconv.ParseInt(itemID, 10, 64); err == nil && id > 0 {
			filter.ItemID = &id
		}
	}
	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}
	if username := c.Query("username"); username != "" {
		filter.Username = &username
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &t
		}
	}
	records, err := h.historyUsecase.GetHistory(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("ExportHistoryCSV failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=history_export.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	if err := writer.Write([]string{"ID", "Item ID", "Action", "Changed By", "Changed At", "Old Name", "Old SKU", "Old Quantity", "Old Price", "New Name", "New SKU", "New Quantity", "New Price"}); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write CSV header")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	for _, rec := range records {
		oldName, oldSKU, oldQty, oldPrice := "", "", "0", "0"
		newName, newSKU, newQty, newPrice := "", "", "0", "0"
		if rec.OldData != nil {
			oldName = rec.OldData.Name
			oldSKU = rec.OldData.SKU
			oldQty = strconv.Itoa(rec.OldData.Quantity)
			oldPrice = strconv.FormatFloat(rec.OldData.Price, 'f', 2, 64)
		}
		if rec.NewData != nil {
			newName = rec.NewData.Name
			newSKU = rec.NewData.SKU
			newQty = strconv.Itoa(rec.NewData.Quantity)
			newPrice = strconv.FormatFloat(rec.NewData.Price, 'f', 2, 64)
		}
		if err := writer.Write([]string{
			strconv.FormatInt(rec.ID, 10),
			strconv.FormatInt(rec.ItemID, 10),
			rec.Action,
			rec.ChangedBy,
			rec.ChangedAt.Format(time.RFC3339),
			oldName, oldSKU, oldQty, oldPrice,
			newName, newSKU, newQty, newPrice,
		}); err != nil {
			h.logger.Error().Err(err).Msg("Failed to write CSV row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
	}
}

func (h *HistoryHandler) toHistoryRecordResponse(rec *domain.HistoryRecord) *dto.HistoryRecordResponse {
	resp := &dto.HistoryRecordResponse{
		ID:        rec.ID,
		ItemID:    rec.ItemID,
		Action:    rec.Action,
		ChangedBy: rec.ChangedBy,
		ChangedAt: rec.ChangedAt,
	}
	if rec.OldData != nil {
		resp.OldData = &dto.ItemData{
			ID:       rec.OldData.ID,
			Name:     rec.OldData.Name,
			SKU:      rec.OldData.SKU,
			Quantity: rec.OldData.Quantity,
			Price:    rec.OldData.Price,
		}
	}
	if rec.NewData != nil {
		resp.NewData = &dto.ItemData{
			ID:       rec.NewData.ID,
			Name:     rec.NewData.Name,
			SKU:      rec.NewData.SKU,
			Quantity: rec.NewData.Quantity,
			Price:    rec.NewData.Price,
		}
	}
	return resp
}
