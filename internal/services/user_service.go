package services

import (
	"context"
	"errors"
	"time"

	"projectpi-backend/internal/models"
	"projectpi-backend/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	DB *mongo.Database
}

func (s *UserService) CreateUser(username, email, password string) error {
	collection := s.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	count, err := collection.CountDocuments(ctx, bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": username},
		},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	// Create user with auto-generated ID
	userID := utils.GenerateUserID(uint(time.Now().UnixNano() % 10000))
	user := models.User{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, user)
	return err
}

func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	collection := s.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}
