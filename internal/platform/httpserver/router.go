package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter creates the shared Gin router with process liveness.
// Business routes and authentication are registered by modules through app wiring.
func NewRouter() *gin.Engine {
	router := gin.New()
	// No reverse proxy is trusted by default; forwarded headers cannot spoof IPs.
	_ = router.SetTrustedProxies(nil)
	router.HandleMethodNotAllowed = true
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipQueryString: true}), gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})
	health := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	router.GET("/healthz", health)
	router.HEAD("/healthz", health)
	return router
}
