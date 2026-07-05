package router

import (
	"pickup-helper/internal/handler"

	"github.com/gin-gonic/gin"
)

// Register wires application routes onto the engine.
// Middleware chain will be assembled here in plan 01-03; for now it just
// delegates to HealthHandler.Register.
func Register(engine *gin.Engine, h *handler.HealthHandler) {
	h.Register(engine)
}
