package comment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var input struct {
		ThreadID        *uint  `json:"thread_id" binding:"required"`
		ParentCommentID *uint  `json:"parent_comment_id"`
		Content         string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ThreadID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID must be provided"})
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
		UserID:  currentUser.UserID,
		Content: input.Content,
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
	}

	if input.ParentCommentID != nil {
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

	if comment.UserID != currentUser.UserID && currentUser.RoleID <= 0 {
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
