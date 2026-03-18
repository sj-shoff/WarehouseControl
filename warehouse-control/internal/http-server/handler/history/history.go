package history_handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"
	"warehouse-control/internal/http-server/handler/history/dto"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

type HistoryHandler struct {
	historyUsecase HistoryUsecase
	logger         *zlog.Zerolog
}

func NewHandler(historyUsecase HistoryUsecase, logger *zlog.Zerolog) *HistoryHandler {
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
		h.writeError(c, err)
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
		Total:   len(records),
	}
	for i, rec := range records {
		resp.Records[i] = dto.ToHistoryRecordResponse(rec)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HistoryHandler) GetItemHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}
	records, err := h.historyUsecase.GetHistoryByItemID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItemHistory failed")
		h.writeError(c, err)
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
	}
	for i, rec := range records {
		resp.Records[i] = dto.ToHistoryRecordResponse(rec)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HistoryHandler) ExportHistoryCSV(c *gin.Context) {
	filter := domain.HistoryFilter{
		Limit:  10000,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, _ := strconv.Atoi(limitStr); l > 0 {
			filter.Limit = l
		}
	}
	if itemID := c.Query("item_id"); itemID != "" {
		if id, _ := strconv.ParseInt(itemID, 10, 64); id > 0 {
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
		h.logger.Error().Err(err).Msg("ExportHistoryCSV: failed to get history")
		h.writeError(c, err)
		return
	}

	filename := fmt.Sprintf("warehouse_history_%s_%s.csv",
		time.Now().Format("2006-01-02"),
		time.Now().Format("15-04-05"))

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if _, err := c.Writer.Write([]byte("\xef\xbb\xbf")); err != nil {
		h.logger.Error().Err(err).Msg("failed to write BOM")
		return
	}

	writer := csv.NewWriter(c.Writer)
	writer.Comma = ';'
	writer.UseCRLF = true
	defer writer.Flush()

	write := func(row []string) bool {
		if err := writer.Write(row); err != nil {
			h.logger.Error().Err(err).Msg("CSV write error")
			return false
		}
		return true
	}

	if !write([]string{"ОТЧЁТ ПО ИСТОРИИ ИЗМЕНЕНИЙ — Warehouse Control"}) {
		return
	}
	if !write([]string{""}) {
		return
	}

	period := "За всё время"
	if filter.DateFrom != nil && filter.DateTo != nil {
		period = fmt.Sprintf("Период: %s — %s",
			filter.DateFrom.Format("02.01.2006"),
			filter.DateTo.Format("02.01.2006"))
	}
	if !write([]string{period}) {
		return
	}
	if !write([]string{""}) {
		return
	}

	header := []string{
		"ID записи", "ID товара", "Действие", "Пользователь", "Дата изменения",
		"Старое название", "Старый SKU", "Старое кол-во", "Старая цена",
		"Новое название", "Новый SKU", "Новое кол-во", "Новая цена",
	}
	if !write(header) {
		return
	}

	for _, rec := range records {
		oldName, oldSKU, oldQty, oldPrice := "", "", "0", "0.00"
		newName, newSKU, newQty, newPrice := "", "", "0", "0.00"

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

		row := []string{
			strconv.FormatInt(rec.ID, 10),
			strconv.FormatInt(rec.ItemID, 10),
			rec.Action,
			rec.ChangedBy,
			rec.ChangedAt.Format("02.01.2006 15:04:05"),
			oldName, oldSKU, oldQty, oldPrice,
			newName, newSKU, newQty, newPrice,
		}

		if !write(row) {
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		h.logger.Error().Err(err).Msg("CSV writer final error")
	} else {
		h.logger.Info().Int("records", len(records)).Msg("CSV export complete")
	}
}

func (h *HistoryHandler) GetDiff(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.writeError(c, customErr.ErrInvalidInput)
		return
	}

	res, err := h.historyUsecase.GetDiff(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", id).Msg("GetDiff failed")
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MapToDTO(res))
}

func (h *HistoryHandler) writeError(c *gin.Context, err error) {
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
