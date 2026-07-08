package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"windpower-monitor/internal/model"
	"windpower-monitor/internal/service"
)

var validate = validator.New()

func CreateTurbine(c *gin.Context) {
	var turbine model.WindTurbine
	if err := c.ShouldBindJSON(&turbine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(&turbine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewTurbineService()
	if err := svc.CreateTurbine(&turbine); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, turbine)
}

func GetTurbines(c *gin.Context) {
	svc := service.NewTurbineService()
	turbines, err := svc.GetAllTurbines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, turbines)
}

func GetTurbine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	svc := service.NewTurbineService()
	turbine, err := svc.GetTurbineByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Turbine not found"})
		return
	}

	c.JSON(http.StatusOK, turbine)
}

func UpdateTurbine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var turbine model.WindTurbine
	if err := c.ShouldBindJSON(&turbine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	turbine.ID = uint(id)

	svc := service.NewTurbineService()
	if err := svc.UpdateTurbine(&turbine); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, turbine)
}

func DeleteTurbine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	svc := service.NewTurbineService()
	if err := svc.DeleteTurbine(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func GetSystemStatistics(c *gin.Context) {
	svc := service.NewTurbineService()
	stats, err := svc.GetSystemStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
