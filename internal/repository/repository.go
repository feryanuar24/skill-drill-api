package repository

import (
	"context"
	"time"

	"skill-drill-api/internal/models"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateQuestions(ctx context.Context, questions []models.Question) error {
	return r.db.WithContext(ctx).CreateInBatches(questions, 100).Error
}

func (r *Repository) TopicExists(ctx context.Context, topicID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Topic{}).Where("id = ?", topicID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) FindQuestion(ctx context.Context, questionID, topicID uint) (*models.Question, error) {
	var question models.Question
	err := r.db.WithContext(ctx).Where("id = ? AND topic_id = ?", questionID, topicID).First(&question).Error
	return &question, err
}

func (r *Repository) FindUser(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	return &user, err
}

func (r *Repository) SaveUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) CreateSession(ctx context.Context, session *models.StudySession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *Repository) UsersToNotify(ctx context.Context, today time.Time) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("fcm_token IS NOT NULL AND (last_study_date IS NULL OR last_study_date::date <> ?)", today.Format("2006-01-02")).Find(&users).Error
	return users, err
}

// User CRUD Repository methods
func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) GetUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) DeleteUser(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// Topic CRUD Repository methods
func (r *Repository) CreateTopic(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

func (r *Repository) GetTopics(ctx context.Context) ([]models.Topic, error) {
	var topics []models.Topic
	err := r.db.WithContext(ctx).Find(&topics).Error
	return topics, err
}

func (r *Repository) GetTopicByID(ctx context.Context, id uint) (*models.Topic, error) {
	var topic models.Topic
	err := r.db.WithContext(ctx).First(&topic, id).Error
	return &topic, err
}

func (r *Repository) UpdateTopic(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Save(topic).Error
}

func (r *Repository) DeleteTopic(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Topic{}, id).Error
}
