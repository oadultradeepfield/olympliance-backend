package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

type InteractionHandler struct {
	db *gorm.DB
}

func NewInteractionHandler(db *gorm.DB) *InteractionHandler {
	return &InteractionHandler{db: db}
}

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

func (h *InteractionHandler) CreateInteraction(c *gin.Context) {
	var input struct {
		ThreadID        *uint  `json:"thread_id"`
		CommentID       *uint  `json:"comment_id"`
		InteractionType string `json:"interaction_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if (input.ThreadID == nil && input.CommentID == nil) || (input.ThreadID != nil && input.CommentID != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either thread_id or comment_id must be provided, but not both"})
		return
	}

	validTypes := map[string]bool{"upvote": true, "downvote": true, "follow": true}
	if !validTypes[input.InteractionType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interaction type"})
		return
	}

	if input.InteractionType == "follow" && input.CommentID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Follow interaction is not allowed for comments"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	var existingInteraction models.Interaction
	if input.ThreadID != nil {
		if input.InteractionType == "follow" {
			if err := h.db.Where("user_id = ? AND thread_id = ? AND interaction_type = ?", currentUser.UserID, *input.ThreadID, "follow").First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Already following this thread"})
				return
			}
		} else {
			if err := h.db.Where("user_id = ? AND thread_id = ? AND interaction_type = ?", currentUser.UserID, *input.ThreadID, input.InteractionType).First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction already exists for this thread"})
				return
			}

			if err := h.db.Where("user_id = ? AND thread_id = ? AND interaction_type IN (?, ?)", currentUser.UserID, *input.ThreadID, "upvote", "downvote").First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted on this thread"})
				return
			}
		}
	}

	if input.CommentID != nil {
		if input.InteractionType == "follow" {
			if err := h.db.Where("user_id = ? AND comment_id = ? AND interaction_type = ?", currentUser.UserID, *input.CommentID, "follow").First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Already following this comment"})
				return
			}
		} else {
			if err := h.db.Where("user_id = ? AND comment_id = ? AND interaction_type = ?", currentUser.UserID, *input.CommentID, input.InteractionType).First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Interaction already exists for this comment"})
				return
			}

			if err := h.db.Where("user_id = ? AND comment_id = ? AND interaction_type IN (?, ?)", currentUser.UserID, *input.CommentID, "upvote", "downvote").First(&existingInteraction).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted on this comment"})
				return
			}
		}
	}

	interaction := models.Interaction{
		UserID:          currentUser.UserID,
		InteractionType: input.InteractionType,
	}

	if input.ThreadID != nil {
		interaction.ThreadID = *input.ThreadID
	}

	if input.CommentID != nil {
		interaction.CommentID = *input.CommentID
	}

	if err := h.db.Create(&interaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create interaction"})
		return
	}

	if input.ThreadID != nil {
		if err := h.updateThreadStats(*input.ThreadID, input.InteractionType, 1); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread stats"})
			return
		}
	}

	if input.CommentID != nil {
		if err := h.updateCommentStats(*input.CommentID, input.InteractionType, 1); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment stats"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interaction created successfully"})
}

func (h *InteractionHandler) UpdateInteraction(c *gin.Context) {
	interactionID := c.Param("id")

	var input struct {
		InteractionType string `json:"interaction_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	validTypes := map[string]bool{"upvote": true, "downvote": true, "follow": true}
	if !validTypes[input.InteractionType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interaction type"})
		return
	}

	var existingInteraction models.Interaction
	if err := h.db.Where("interaction_id = ? AND user_id = ?", interactionID, currentUser.UserID).
		First(&existingInteraction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Interaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interaction"})
		}
		return
	}

	if existingInteraction.InteractionType == input.InteractionType {
		if err := h.db.Delete(&existingInteraction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove interaction"})
			return
		}
		if existingInteraction.ThreadID != 0 {
			if err := h.updateThreadStats(existingInteraction.ThreadID, existingInteraction.InteractionType, -1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread stats"})
				return
			}
		}
		if existingInteraction.CommentID != 0 {
			if err := h.updateCommentStats(existingInteraction.CommentID, existingInteraction.InteractionType, -1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment stats"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "Interaction removed"})
		return
	}

	if (existingInteraction.InteractionType == "upvote" && input.InteractionType == "downvote") ||
		(existingInteraction.InteractionType == "downvote" && input.InteractionType == "upvote") {
		if existingInteraction.ThreadID != 0 {
			if err := h.updateThreadStats(existingInteraction.ThreadID, existingInteraction.InteractionType, -1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread stats"})
				return
			}
			if err := h.updateThreadStats(existingInteraction.ThreadID, input.InteractionType, 1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread stats"})
				return
			}
		}
		if existingInteraction.CommentID != 0 {
			if err := h.updateCommentStats(existingInteraction.CommentID, existingInteraction.InteractionType, -1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment stats"})
				return
			}
			if err := h.updateCommentStats(existingInteraction.CommentID, input.InteractionType, 1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment stats"})
				return
			}
		}
		existingInteraction.InteractionType = input.InteractionType
		if err := h.db.Save(&existingInteraction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interaction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Interaction updated", "interaction": existingInteraction})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected interaction update"})
}

func (h *InteractionHandler) updateThreadStats(threadID uint, interactionType string, adjustment int) error {
	fields := map[string]string{
		"upvote":   "upvotes",
		"downvote": "downvotes",
		"follow":   "followers",
	}

	field, exists := fields[interactionType]
	if !exists {
		return fmt.Errorf("invalid interaction type: %s", interactionType)
	}

	return h.db.Model(&models.Thread{}).
		Where("thread_id = ?", threadID).
		Update("stats", gorm.Expr(
			"jsonb_set(stats, '{"+field+"}', to_jsonb(((stats->>'"+field+"')::int + ?)::int))",
			adjustment)).Error
}

func (h *InteractionHandler) updateCommentStats(commentID uint, interactionType string, adjustment int) error {
	fields := map[string]string{
		"upvote":   "upvotes",
		"downvote": "downvotes",
	}

	field, exists := fields[interactionType]
	if !exists {
		return fmt.Errorf("invalid interaction type: %s", interactionType)
	}

	return h.db.Model(&models.Comment{}).
		Where("comment_id = ?", commentID).
		Update("stats", gorm.Expr(
			"jsonb_set(stats, '{"+field+"}', to_jsonb(((stats->>'"+field+"')::int + ?)::int))",
			adjustment)).Error
}
