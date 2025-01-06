package interaction

import (
	"fmt"

	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

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
