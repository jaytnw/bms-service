package utils

import (
	"github.com/gofiber/fiber/v3"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func JSON(c fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c fiber.Ctx, status int, msg, code string) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Message: msg,
			Code:    code,
		},
	})
}
