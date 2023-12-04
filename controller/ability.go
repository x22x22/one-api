package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/model"
)

func UpdateAllAbilities(c *gin.Context) {
	err := model.UpdateAllAbilities()
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
	})
	return
}
