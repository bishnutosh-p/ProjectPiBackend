package services

import (
	"errors"

	"projectpi-backend/internal/models"
	"projectpi-backend/internal/utils"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func (s *UserService) CreateUser(username, email, password string) error {
	// Check if user already exists
	var count int64
	s.DB.Model(&models.User{}).Where("email = ? OR username = ?", email, username).Count(&count)
	if count > 0 {
		return errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	// Create user
	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	err = s.DB.Create(&user).Error
	if err != nil {
		return err
	}

	// After user is created and has a DB ID:
	user.UserID = utils.GenerateUserID(user.ID)
	return s.DB.Save(&user).Error
}

func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}
	return &user, nil
}
