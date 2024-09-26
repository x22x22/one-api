package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"net/http"
)

type LLMCacheQuery struct {
	Model    string                `json:"model"`
	Contents []common.ContentQuery `json:"contents"`
	PageSize int                   `json:"page_size"`
	Cursor   uint64                `json:"cursor"`
}

func GetLLMCache(c *gin.Context) {
	if !common.RedisLLMEnabled {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Cache query is not supported in file_cache mode"})
		return
	}

	var query LLMCacheQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if query.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model is required"})
		return
	}

	if query.PageSize == 0 {
		query.PageSize = 10 // 默认每页返回10条
	}

	cacheItems, nextCursor, err := common.RedisSearchCacheWithMultiPath("llm_cache:"+query.Model+":*", query.Contents, query.Cursor, int64(query.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cache from Redis"})
		return
	}

	if cacheItems == nil {
		cacheItems = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"cache_items": cacheItems,
		"next_cursor": nextCursor,
		"page_size":   query.PageSize,
		"count":       len(cacheItems),
	})
}

func DeleteLLMCache(c *gin.Context) {
	if !common.RedisLLMEnabled {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Cache deletion is not supported in file_cache mode"})
		return
	}

	var query LLMCacheQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if query.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model is required"})
		return
	}

	deletedCount, err := common.RedisDelByContentWithMultiPath("llm_cache:"+query.Model+":", query.Contents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cache from Redis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cache deleted successfully", "deleted_count": deletedCount})
}
