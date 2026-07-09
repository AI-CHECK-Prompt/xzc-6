package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"windpower-monitor/config"
	"windpower-monitor/internal/controller"
	"windpower-monitor/internal/model"
	"windpower-monitor/internal/service"
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

	initHealthConfig()

	go startHealthCalcScheduler()

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

		health := api.Group("/health")
		{
			health.GET("/snapshot/:id", controller.GetHealthSnapshot)
			health.GET("/history/:id", controller.GetHealthHistory)
			health.GET("/alert/:id", controller.GetActiveAlert)
			health.GET("/alerts", controller.GetAllActiveAlerts)
			health.POST("/adjust/:id", controller.ManualAdjustHealth)
			health.GET("/adjustments/:id", controller.GetAdjustments)
			health.GET("/template", controller.GetTemplate)
			health.GET("/templates", controller.GetAllTemplates)
			health.POST("/template", controller.CreateTemplate)
			health.PUT("/template/:id", controller.UpdateTemplate)
			health.DELETE("/template/:id", controller.DeleteTemplate)
			health.GET("/config", controller.GetConfig)
			health.PUT("/config", controller.UpdateConfig)
			health.POST("/calc/:id", controller.TriggerHealthCalc)
			health.GET("/ranking", controller.GetTurbineHealthRanking)
			health.POST("/backfill/:id", controller.BackfillHealth)
		}
	}

	port := viper.GetString("server.port")
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}

func initHealthConfig() {
	svc := service.NewHealthService()
	if err := svc.InitDefaultConfig(); err != nil {
		log.Printf("【健康配置-错误】默认配置初始化失败: %v", err)
	} else {
		log.Printf("【健康配置-初始化】默认配置和模板已创建")
	}
}

func startHealthCalcScheduler() {
	log.Printf("【健康计算-定时任务】启动每小时健康指数计算")

	for {
		now := time.Now()
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		sleepDuration := nextHour.Sub(now)

		log.Printf("【健康计算-定时任务】下次计算时间: %v", nextHour)
		time.Sleep(sleepDuration)

		log.Printf("【健康计算-定时任务】开始执行批量健康指数计算")
		svc := service.NewHealthService()
		if err := svc.CalculateAllTurbines(); err != nil {
			log.Printf("【健康计算-定时任务】批量计算失败: %v", err)
		}
	}
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
