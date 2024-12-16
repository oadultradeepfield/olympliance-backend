package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

type ThreadHandler struct {
	db *gorm.DB
}

func NewThreadHandler(db *gorm.DB) *ThreadHandler {
	return &ThreadHandler{db: db}
}

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

func (h *ThreadHandler) GetThread(c *gin.Context) {
	threadID := c.Param("id")
	var thread models.Thread

	if err := h.db.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.IsDeleted {
		c.JSON(http.StatusGone, gin.H{"error": "Thread is deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"thread": thread})
}

func (h *ThreadHandler) GetFollowedThreads(c *gin.Context) {
	userId := c.Param("id")
	var threads []models.Thread

	isDeleted := c.DefaultQuery("is_deleted", "false")
	showDeleted := isDeleted == "true"

	sortBy := c.DefaultQuery("sort_by", "updated_at")

	validSortFields := []string{"upvotes", "comments", "created_at", "updated_at"}
	if !contains(validSortFields, sortBy) {
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

	var interactions []models.Interaction
	if err := h.db.Where("user_id = ? AND interaction_type = ?", userId, "follow").Find(&interactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interactions"})
		return
	}

	var threadIds []uint
	for _, interaction := range interactions {
		threadIds = append(threadIds, interaction.ThreadID)
	}

	if len(threadIds) > 0 {
		query := h.db.Model(&models.Thread{}).
			Where("id IN ?", threadIds).
			Where("is_deleted = ?", showDeleted).
			Limit(perPageInt).
			Offset(offset)

		if sortBy == "followers" || sortBy == "upvotes" || sortBy == "comments" {
			query = query.Order(fmt.Sprintf("stats->>'%s' DESC", sortBy))
		} else {
			query = query.Order(sortBy + " DESC")
		}

		if err := query.Find(&threads).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"threads": threads})
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

func (h *ThreadHandler) GetAllThreadsByCategory(c *gin.Context) {
	categoryID := c.Param("category_id")
	var threads []models.Thread

	isDeleted := c.DefaultQuery("is_deleted", "false")
	showDeleted := isDeleted == "true"

	sortBy := c.DefaultQuery("sort_by", "updated_at")

	validSortFields := []string{"upvotes", "comments", "created_at", "updated_at"}
	if !contains(validSortFields, sortBy) {
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

	query := h.db.Model(&models.Thread{}).
		Where("category_id = ?", categoryID).
		Where("is_deleted = ?", showDeleted).
		Limit(perPageInt).
		Offset(offset)

	if sortBy == "followers" || sortBy == "upvotes" || sortBy == "comments" {
		query = query.Order(fmt.Sprintf("stats->>'%s' DESC", sortBy))
	} else {
		query = query.Order(sortBy + " DESC")
	}

	if err := query.Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"threads": threads})
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
