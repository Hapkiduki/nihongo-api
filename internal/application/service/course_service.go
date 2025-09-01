package service

import (
	"context"
	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"
)

// CourseService handles course business logic
type CourseService struct {
	courseRepo ports.CourseRepository
}

// NewCourseService creates a new course service
func NewCourseService(courseRepo ports.CourseRepository) *CourseService {
	return &CourseService{
		courseRepo: courseRepo,
	}
}

// GetAllCourses retrieves all courses
func (s *CourseService) GetAllCourses(ctx context.Context) ([]domain.Course, error) {
	return s.courseRepo.GetAll(ctx)
}

// GetCourseByID retrieves a course by ID
func (s *CourseService) GetCourseByID(ctx context.Context, id string) (*domain.Course, error) {
	return s.courseRepo.GetByID(ctx, id)
}

// GetCoursesByLevel retrieves courses by JLPT level
func (s *CourseService) GetCoursesByLevel(ctx context.Context, level domain.JLPTLevel) ([]domain.Course, error) {
	return s.courseRepo.GetByLevel(ctx, level)
}

// GetPremiumCourses retrieves premium courses
func (s *CourseService) GetPremiumCourses(ctx context.Context) ([]domain.Course, error) {
	return s.courseRepo.GetPremium(ctx)
}

// CheckPremiumAccess checks if user has access to premium course
func (s *CourseService) CheckPremiumAccess(ctx context.Context, courseID string, userRevenueCatID string) (bool, error) {
	// This would integrate with RevenueCat API
	// For now, return true for demonstration
	return true, nil
}
