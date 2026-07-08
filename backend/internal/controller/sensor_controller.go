package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"windpower-monitor/internal/model"
	"windpower-monitor/internal/service"
)

func CollectSensorData(c *gin.Context) {
	var sensorData model.SensorData
	if err := c.ShouldBindJSON(&sensorData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(&sensorData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewSensorService()
	if err := svc.CollectSensorData(&sensorData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sensorData)
}

func GetTurbineData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	svc := service.NewSensorService()
	data, err := svc.GetSensorDataByTurbineID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func GetTurbineStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	svc := service.NewSensorService()
	status, err := svc.GetTurbineStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func GetTrendData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	svc := service.NewSensorService()
	data, err := svc.GetTrendData(uint(id), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func GetStatistics(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	svc := service.NewSensorService()
	stats, err := svc.GetAllTurbineStatistics(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func ExportData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	svc := service.NewSensorService()
	data, err := svc.ExportData(uint(id), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "风机数据"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "风机ID", "时间戳", "转速(RPM)", "功率(kW)", "温度(°C)", "湿度(%)", "振动", "采集时间"}
	for i, header := range headers {
		cellName, _ := excelize.CoordinateToCellName(i+1, 1)
		f.SetCellValue(sheet, cellName, header)
	}

	for i, item := range data {
		row := i + 2
		values := []interface{}{
			item.ID,
			item.TurbineID,
			item.Timestamp.Format(time.RFC3339),
			item.RPM,
			item.Power,
			item.Temperature,
			item.Humidity,
			item.Vibration,
			item.CreatedAt.Format(time.RFC3339),
		}
		for col, value := range values {
			cellName, _ := excelize.CoordinateToCellName(col+1, row)
			f.SetCellValue(sheet, cellName, value)
		}
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=turbine_data.xlsx")
	
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
