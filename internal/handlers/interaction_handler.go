package handlers

import (
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

	interaction := models.Interaction{
		UserID:          currentUser.UserID,
		ThreadID:        *input.ThreadID,
		CommentID:       *input.CommentID,
		InteractionType: input.InteractionType,
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
	var input struct {
		ThreadID        *uint  `json:"thread_id"`
		CommentID       *uint  `json:"comment_id"`
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

	var existingInteraction models.Interaction
	err := h.db.Where("user_id = ? AND (thread_id = ? OR comment_id = ?)", currentUser.UserID, input.ThreadID, input.CommentID).
		First(&existingInteraction).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing interaction"})
		return
	}

	if existingInteraction.InteractionType == input.InteractionType {
		if input.ThreadID != nil {
			h.updateThreadStats(*input.ThreadID, input.InteractionType, -1)
		}
		if input.CommentID != nil {
			h.updateCommentStats(*input.CommentID, input.InteractionType, -1)
		}
		if err := h.db.Delete(&existingInteraction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to undo interaction"})
			return
		}
		existingInteraction.InteractionType = ""
		c.JSON(http.StatusOK, gin.H{"message": "Interaction undone successfully"})
		return
	}

	if (existingInteraction.InteractionType == "upvote" && input.InteractionType == "downvote") || (existingInteraction.InteractionType == "downvote" && input.InteractionType == "upvote") {
		if input.ThreadID != nil {
			h.updateThreadStats(*input.ThreadID, input.InteractionType, 1)
			h.updateThreadStats(existingInteraction.ThreadID, existingInteraction.InteractionType, -1)
		}
		if input.CommentID != nil {
			h.updateCommentStats(*input.CommentID, input.InteractionType, 1)
			h.updateCommentStats(existingInteraction.CommentID, existingInteraction.InteractionType, -1)
		}
		if err := h.db.Delete(&existingInteraction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to undo interaction"})
			return
		}
		existingInteraction.InteractionType = input.InteractionType
		c.JSON(http.StatusOK, gin.H{"message": "Interaction toggled successfully"})
		return
	}
}

func (h *InteractionHandler) updateThreadStats(threadID uint, interactionType string, adjustment int) error {
	field := ""
	switch interactionType {
	case "upvote":
		field = "upvotes"
	case "downvote":
		field = "downvotes"
	case "follow":
		field = "followers"
	}

	if field == "" {
		return nil
	}

	return h.db.Model(&models.Thread{}).
		Where("thread_id = ?", threadID).
		Update("stats", gorm.Expr(
			"jsonb_set(stats, '{%s}', to_jsonb((stats->>'%s')::int + ?)::text::jsonb)",
			field, field, adjustment)).Error
}

func (h *InteractionHandler) updateCommentStats(commentID uint, interactionType string, adjustment int) error {
	field := ""
	switch interactionType {
	case "upvote":
		field = "upvotes"
	case "downvote":
		field = "downvotes"
	}

	if field == "" {
		return nil
	}

	return h.db.Model(&models.Comment{}).
		Where("comment_id = ?", commentID).
		Update("stats", gorm.Expr(
			"jsonb_set(stats, '{%s}', to_jsonb((stats->>'%s')::int + ?)::text::jsonb)",
			field, field, adjustment)).Error
}
