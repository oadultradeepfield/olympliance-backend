package thread

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
)

func (h *ThreadHandler) CreateThread(c *gin.Context) {
	var input struct {
		Title      string   `json:"title" binding:"required"`
		Content    string   `json:"content" binding:"required"`
		CategoryID uint     `json:"category_id" binding:"required"`
		Tags       []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	thread := models.Thread{
		UserID:     currentUser.UserID,
		Title:      input.Title,
		Content:    input.Content,
		CategoryID: input.CategoryID,
		Tags:       input.Tags,
	}

	if err := h.db.Create(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"thread": thread})
}

func (h *ThreadHandler) UpdateThread(c *gin.Context) {
	threadID := c.Param("id")
	var input struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var thread models.Thread
	if err := h.db.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
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

	if thread.UserID != currentUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this thread"})
		return
	}

	if input.Title != "" {
		thread.Title = input.Title
	}
	if input.Content != "" {
		thread.Content = input.Content
	}
	if input.Tags != nil {
		thread.Tags = input.Tags
	}

	if err := h.db.Save(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"thread": thread})
}

func (h *ThreadHandler) DeleteThread(c *gin.Context) {
	threadID := c.Param("id")

	var thread models.Thread
	if err := h.db.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
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

	if thread.UserID != currentUser.UserID && currentUser.RoleID <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this thread"})
		return
	}

	thread.IsDeleted = true
	if err := h.db.Save(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted"})
}
