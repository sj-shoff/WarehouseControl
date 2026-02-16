package history_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"warehouse-control/internal/domain"
	"warehouse-control/internal/http-server/handler/history/dto"

	"github.com/go-chi/chi/v5"
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

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	filter := domain.HistoryFilter{
		Limit:  100,
		Offset: 0,
	}
	if itemID := r.URL.Query().Get("item_id"); itemID != "" {
		if id, err := strconv.ParseInt(itemID, 10, 64); err == nil && id > 0 {
			filter.ItemID = &id
		}
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}
	if username := r.URL.Query().Get("username"); username != "" {
		filter.Username = &username
	}
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &t
		}
	}
	records, err := h.historyUsecase.GetHistory(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetHistory failed")
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
	}
	for i, rec := range records {
		resp.Records[i] = h.toHistoryRecordResponse(rec)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HistoryHandler) GetItemHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid_input", http.StatusBadRequest)
		return
	}
	records, err := h.historyUsecase.GetHistoryByItemID(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItemHistory failed")
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	resp := dto.HistoryResponse{
		Records: make([]*dto.HistoryRecordResponse, len(records)),
	}
	for i, rec := range records {
		resp.Records[i] = h.toHistoryRecordResponse(rec)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HistoryHandler) ExportHistoryCSV(w http.ResponseWriter, r *http.Request) {
	filter := domain.HistoryFilter{
		Limit:  1000,
		Offset: 0,
	}
	records, err := h.historyUsecase.GetHistory(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("ExportHistoryCSV failed")
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=history_export.csv")
	w.Write([]byte("ID,Item ID,Action,Changed By,Changed At,Old Name,New Name\n"))
	for _, rec := range records {
		oldName := ""
		newName := ""
		if rec.OldData != nil {
			oldName = rec.OldData.Name
		}
		if rec.NewData != nil {
			newName = rec.NewData.Name
		}
		line := []byte(
			strconv.FormatInt(rec.ID, 10) + "," +
				strconv.FormatInt(rec.ItemID, 10) + "," +
				rec.Action + "," +
				rec.ChangedBy + "," +
				rec.ChangedAt.Format(time.RFC3339) + "," +
				oldName + "," +
				newName + "\n",
		)
		w.Write(line)
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
