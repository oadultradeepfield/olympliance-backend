package comment

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"github.com/oadultradeepfield/olympliance-server/internal/services"
)

func (h *CommentHandler) GetAllComments(c *gin.Context) {
	var comments []models.Comment

	threadIDStr := c.DefaultQuery("thread_id", "")
	sortBy := c.DefaultQuery("sort_by", "updated_at")

	validSortFields := []string{"upvotes", "created_at", "updated_at"}
	if !services.Contains(validSortFields, sortBy) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_by field"})
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "10")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	perPageInt, err := strconv.Atoi(perPage)
	if err != nil || perPageInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid per_page number"})
		return
	}

	offset := (pageInt - 1) * perPageInt

	query := h.db.Model(&models.Comment{}).
		Limit(perPageInt).
		Offset(offset)

	if threadIDStr != "" {
		threadID, err := strconv.ParseUint(threadIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread_id"})
			return
		}
		query = query.Where("thread_id = ?", threadID)
	}

	if sortBy == "upvotes" {
		query = query.Order(fmt.Sprintf("stats->>'%s' DESC", sortBy) + ", created_at ASC")
	} else {
		query = query.Order(sortBy + " ASC")
	}

	if err := query.Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"page":     pageInt,
		"per_page": perPageInt,
		"total":    len(comments),
	})
}
