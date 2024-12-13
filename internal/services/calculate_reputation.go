package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type ReputationCalculator struct {
	db *gorm.DB
}

func NewReputationCalculator(db *gorm.DB) *ReputationCalculator {
	return &ReputationCalculator{db: db}
}

func (b *ReputationCalculator) CalculateReputation(userID uint) (int, error) {
	var totalReputation int

	var threads []models.Thread
	if err := b.db.Where("user_id = ?", userID).Find(&threads).Error; err != nil {
		return 0, fmt.Errorf("error fetching threads: %v", err)
	}

	for _, thread := range threads {
		var stats map[string]int
		if err := json.Unmarshal(thread.Stats, &stats); err != nil {
			log.Printf("error unmarshalling thread stats for thread_id %d: %v", thread.ThreadID, err)
			continue
		}

		threadReputation := max(stats["upvotes"]-stats["downvotes"], 0) + stats["followers"] + stats["comments"]
		totalReputation += threadReputation
	}

	var comments []models.Comment
	if err := b.db.Where("user_id = ?", userID).Find(&comments).Error; err != nil {
		return 0, fmt.Errorf("error fetching comments: %v", err)
	}

	for _, comment := range comments {
		var stats map[string]int
		if err := json.Unmarshal(comment.Stats, &stats); err != nil {
			log.Printf("error unmarshalling comment stats for comment_id %d: %v", comment.CommentID, err)
			continue
		}

		commentReputation := max(stats["upvotes"]-stats["downvotes"], 0)
		totalReputation += commentReputation
	}

	return totalReputation, nil
}

func (b *ReputationCalculator) AssignReputationToUser(userID uint) error {
	reputation, err := b.CalculateReputation(userID)
	if err != nil {
		return fmt.Errorf("error calculating reputation: %v", err)
	}

	var user models.User
	if err := b.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	user.Reputation = reputation

	if err := b.db.Save(&user).Error; err != nil {
		return fmt.Errorf("error updating user reputation: %v", err)
	}

	return nil
}

func (b *ReputationCalculator) ScheduleDailyReputationJob() {
	c := cron.New()

	// Schedule a daily reputation update
	_, err := c.AddFunc("0 0 * * *", func() {
		var users []models.User
		if err := b.db.Find(&users).Error; err != nil {
			log.Fatalf("Failed to fetch users: %v", err)
			return
		}

		for _, user := range users {
			if err := b.AssignReputationToUser(user.UserID); err != nil {
				log.Fatalf("Failed to assign reputation to user %d: %v", user.UserID, err)
			} else {
				log.Printf("Successfully updated reputation for user %d", user.UserID)
			}
		}
	})

	if err != nil {
		log.Fatalf("Failed to schedule cron job: %v", err)
	}

	c.Start()
}
