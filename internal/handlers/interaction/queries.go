package interaction

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

func (h *InteractionHandler) GetInteraction(c *gin.Context) {
	userID := c.Query("user_id")
	threadID := c.Query("thread_id")
	commentID := c.Query("comment_id")

	if (threadID == "" && commentID == "") || (threadID != "" && commentID != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either thread_id or comment_id must be provided"})
		return
	}

	var interactions []models.Interaction
	query := h.db.Where("user_id = ?", userID)

	if threadID != "" {
		query = query.Where("thread_id = ?", threadID)
	}
	if commentID != "" {
		query = query.Where("comment_id = ?", commentID)
	}

	if err := query.Find(&interactions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"interactions": []models.Interaction{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interactions": interactions})
}
