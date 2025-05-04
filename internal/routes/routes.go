package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jaytnw/bms-service/internal/handlers"
)

func Setup(app fiber.Router, statusHandler *handlers.StatusHandler) {

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Welcome to BMS Service 👋")
	})

	v1 := app.Group("/v1")

	status := v1.Group("/status")
	status.Get("/", statusHandler.GetAllStatus)
	status.Get("/:washerID", statusHandler.GetStatusByWasherID)
	status.Get("/:washerID/history", statusHandler.GetStatusHistoryByWasherID)

	status.Get("/dorm/report", statusHandler.GetDormStatusReport)

}
