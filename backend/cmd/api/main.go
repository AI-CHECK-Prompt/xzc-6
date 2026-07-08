package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"windpower-monitor/config"
	"windpower-monitor/internal/controller"
	"windpower-monitor/internal/model"
	"windpower-monitor/pkg/database"
	"windpower-monitor/pkg/redis"
)

func main() {
	config.Init()

	err := database.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	redis.Init()
	defer redis.Close()

	model.AutoMigrate(database.DB)

	r := gin.Default()

	r.Use(CORSMiddleware())

	api := r.Group("/api")
	{
		turbine := api.Group("/turbines")
		{
			turbine.POST("", controller.CreateTurbine)
			turbine.GET("", controller.GetTurbines)
			turbine.GET("/:id", controller.GetTurbine)
			turbine.PUT("/:id", controller.UpdateTurbine)
			turbine.DELETE("/:id", controller.DeleteTurbine)
		}

		data := api.Group("/data")
		{
			data.POST("/collect", controller.CollectSensorData)
			data.GET("/turbine/:id", controller.GetTurbineData)
			data.GET("/status/:id", controller.GetTurbineStatus)
			data.GET("/trend/:id", controller.GetTrendData)
			data.GET("/statistics", controller.GetStatistics)
			data.GET("/export/:id", controller.ExportData)
		}

		stats := api.Group("/statistics")
		{
			stats.GET("/system", controller.GetSystemStatistics)
		}
	}

	port := viper.GetString("server.port")
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
