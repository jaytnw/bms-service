package handlers

import (
	"github.com/jaytnw/bms-service/internal/apperr"
	"github.com/jaytnw/bms-service/internal/services"
	"github.com/jaytnw/bms-service/internal/utils"

	"github.com/gofiber/fiber/v3"
)

type StatusHandler struct {
	service services.StatusService
}

func NewStatusHandler(service services.StatusService) *StatusHandler {
	return &StatusHandler{service: service}
}

func (h *StatusHandler) GetAllStatus(c fiber.Ctx) error {
	statuses, err := h.service.GetAllStatus(c.Context())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to get statuses", "GET_STATUS_ERROR")
	}

	return utils.JSON(c, fiber.StatusOK, statuses)
}

func (h *StatusHandler) GetStatusByWasherID(c fiber.Ctx) error {
	washerID := c.Params("washer_id")

	status, err := h.service.GetStatusByWasherID(c.Context(), washerID)
	if err != nil {
		if ae, ok := err.(*apperr.AppError); ok {
			return utils.Error(c, ae.Status, ae.Message, ae.Code)
		}
		return utils.Error(c, 500, "Unexpected error", "INTERNAL_ERROR")
	}

	return utils.JSON(c, 200, status)
}

func (h *StatusHandler) GetStatusHistoryByWasherID(c fiber.Ctx) error {
	washerID := c.Params("washer_id")

	history, err := h.service.GetStatusHistoryByWasherID(c.Context(), washerID)
	if err != nil {
		if ae, ok := err.(*apperr.AppError); ok {
			return utils.Error(c, ae.Status, ae.Message, ae.Code)
		}

		return utils.Error(c, 500, "Failed to get history", "HISTORY_FETCH_FAILED")
	}

	return utils.JSON(c, 200, history)
}

func (h *StatusHandler) GetDormStatusReport(c fiber.Ctx) error {
	report, err := h.service.GetDormStatusReport(c.Context())
	if err != nil {
		if ae, ok := err.(*apperr.AppError); ok {
			return utils.Error(c, ae.Status, ae.Message, ae.Code)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to generate dorm report", "DORM_REPORT_ERROR")
	}
	return utils.JSON(c, fiber.StatusOK, report)
}
