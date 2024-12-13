package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var input struct {
		ThreadID        *uint  `json:"thread_id"`
		ParentCommentID *uint  `json:"parent_comment_id"`
		Content         string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if (input.ThreadID == nil && input.ParentCommentID == nil) || (input.ThreadID != nil && input.ParentCommentID != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either thread_id or parent_comment_id must be provided, but not both"})
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

	comment := models.Comment{
		UserID:          currentUser.UserID,
		Content:         input.Content,
		ThreadID:        0,
		ParentCommentID: 0,
	}

	if input.ThreadID != nil {
		comment.ThreadID = *input.ThreadID

		if err := h.db.Model(&models.Thread{}).
			Where("thread_id = ?", *input.ThreadID).
			Update("stats", gorm.Expr("jsonb_set(stats, '{comments}', to_jsonb((stats->>'comments')::int + 1)::text::jsonb)")).
			Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread stats"})
			return
		}

	} else if input.ParentCommentID != nil {
		comment.ParentCommentID = *input.ParentCommentID
	}

	if err := h.db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID := c.Param("id")
	var input struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var comment models.Comment
	if err := h.db.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
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

	if comment.UserID != currentUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this comment"})
		return
	}

	if input.Content != "" {
		comment.Content = input.Content
	}

	if err := h.db.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")

	var comment models.Comment
	if err := h.db.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
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

	if comment.UserID != currentUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this comment"})
		return
	}

	comment.IsDeleted = true
	if err := h.db.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

func (h *CommentHandler) GetAllComments(c *gin.Context) {
	var comments []models.Comment

	isDeleted := c.DefaultQuery("is_deleted", "false")
	showDeleted := isDeleted == "true"

	sortBy := c.DefaultQuery("sort_by", "updated_at")
	validSortFields := []string{"upvotes", "created_at", "updated_at"}

	if !contains(validSortFields, sortBy) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_by field"})
		return
	}

	query := h.db.Model(&models.Comment{}).
		Where("is_deleted = ?", showDeleted)

	query = query.Order(sortBy + " DESC")

	if err := query.Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
