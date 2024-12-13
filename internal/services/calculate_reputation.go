package services

import (
	"encoding/json"
	"log"

	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

type ReputationCalculator struct {
	db *gorm.DB
}

func NewReputationCalculator(db *gorm.DB) *ReputationCalculator {
	return &ReputationCalculator{db: db}
}

func (b *ReputationCalculator) CalculateReputation(userID uint) int {
	var totalReputation int

	var threads []models.Thread
	if err := b.db.Where("user_id = ?", userID).Find(&threads).Error; err != nil {
		log.Fatalf("Error fetching threads: %v", err)
		return 0
	}

	for _, thread := range threads {
		var stats map[string]int
		if err := json.Unmarshal(thread.Stats, &stats); err != nil {
			log.Printf("Error unmarshalling thread stats for thread_id %d: %v", thread.ThreadID, err)
			continue
		}

		threadReputation := max(stats["upvotes"]-stats["downvotes"], 0) + stats["followers"] + stats["comments"]
		totalReputation += threadReputation
	}

	var comments []models.Comment
	if err := b.db.Where("user_id = ?", userID).Find(&comments).Error; err != nil {
		log.Fatalf("error fetching comments: %v", err)
		return 0
	}

	for _, comment := range comments {
		var stats map[string]int
		if err := json.Unmarshal(comment.Stats, &stats); err != nil {
			log.Fatalf("error unmarshalling comment stats for comment_id %d: %v", comment.CommentID, err)
			continue
		}

		commentReputation := max(stats["upvotes"]-stats["downvotes"], 0)
		totalReputation += commentReputation
	}

	return totalReputation
}

func (b *ReputationCalculator) AssignReputationToUser(userID uint) error {
	reputation := b.CalculateReputation(userID)
	var user models.User
	if err := b.db.First(&user, userID).Error; err != nil {
		log.Fatalf("Error fetching a user: %v", err)
	}

	user.Reputation = reputation

	if err := b.db.Save(&user).Error; err != nil {
		log.Fatalf("Error updating user reputation: %v", err)
	}

	return nil
}

func (b *ReputationCalculator) CalculateReputationOnStartup() {
	var users []models.User
	if err := b.db.Find(&users).Error; err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}

	for _, user := range users {
		if err := b.AssignReputationToUser(user.UserID); err != nil {
			log.Printf("Error assigning reputation to user %d: %v", user.UserID, err)
		} else {
			log.Printf("Successfully updated reputation for user %d", user.UserID)
		}
	}
}
