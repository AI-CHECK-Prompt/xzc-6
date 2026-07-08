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
		f.SetCellValue(sheet, string(rune('A'+i))+"1", header)
	}

	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheet, "A"+strconv.Itoa(row), item.ID)
		f.SetCellValue(sheet, "B"+strconv.Itoa(row), item.TurbineID)
		f.SetCellValue(sheet, "C"+strconv.Itoa(row), item.Timestamp.Format(time.RFC3339))
		f.SetCellValue(sheet, "D"+strconv.Itoa(row), item.RPM)
		f.SetCellValue(sheet, "E"+strconv.Itoa(row), item.Power)
		f.SetCellValue(sheet, "F"+strconv.Itoa(row), item.Temperature)
		f.SetCellValue(sheet, "G"+strconv.Itoa(row), item.Humidity)
		f.SetCellValue(sheet, "H"+strconv.Itoa(row), item.Vibration)
		f.SetCellValue(sheet, "I"+strconv.Itoa(row), item.CreatedAt.Format(time.RFC3339))
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=turbine_data.xlsx")
	
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
