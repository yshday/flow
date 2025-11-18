package service

import (
	"context"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// StatisticsService handles statistics business logic
type StatisticsService struct {
	statisticsRepo *repository.StatisticsRepository
	projectRepo    *repository.ProjectRepository
	memberRepo     *repository.ProjectMemberRepository
	cache          pkgcache.Cache
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(
	statisticsRepo *repository.StatisticsRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	cache pkgcache.Cache,
) *StatisticsService {
	return &StatisticsService{
		statisticsRepo: statisticsRepo,
		projectRepo:    projectRepo,
		memberRepo:     memberRepo,
		cache:          cache,
	}
}

// GetProjectStatistics retrieves statistics for a project
func (s *StatisticsService) GetProjectStatistics(ctx context.Context, projectID int, userID int) (*models.ProjectStatistics, error) {
	// Check if user has access to the project
	if err := s.checkProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := pkgcache.BuildProjectStatsKey(projectID)
	var stats models.ProjectStatistics

	if s.cache != nil {
		if redisCache, ok := s.cache.(*pkgcache.RedisCache); ok {
			if err := redisCache.GetJSON(ctx, cacheKey, &stats); err == nil {
				return &stats, nil
			}
		}
	}

	// Cache miss - get from database
	result, err := s.statisticsRepo.GetProjectStatistics(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, result, time.Duration(pkgcache.TTLProjectStats)*time.Second)
	}

	return result, nil
}

// GetIssueStatistics retrieves issue statistics for a project
func (s *StatisticsService) GetIssueStatistics(ctx context.Context, projectID int, userID int) (*models.IssueStatistics, error) {
	// Check if user has access to the project
	if err := s.checkProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := pkgcache.BuildIssueStatsKey(projectID)
	var stats models.IssueStatistics

	if s.cache != nil {
		if redisCache, ok := s.cache.(*pkgcache.RedisCache); ok {
			if err := redisCache.GetJSON(ctx, cacheKey, &stats); err == nil {
				return &stats, nil
			}
		}
	}

	// Cache miss - get from database
	result, err := s.statisticsRepo.GetIssueStatistics(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, result, time.Duration(pkgcache.TTLIssueStats)*time.Second)
	}

	return result, nil
}

// GetUserActivityStatistics retrieves user activity statistics
func (s *StatisticsService) GetUserActivityStatistics(ctx context.Context, userID int) (*models.UserActivityStatistics, error) {
	return s.statisticsRepo.GetUserActivityStatistics(ctx, userID)
}

// checkProjectAccess verifies that a user has access to a project
func (s *StatisticsService) checkProjectAccess(ctx context.Context, projectID int, userID int) error {
	// Check if project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Check if user is project owner
	if project.OwnerID == userID {
		return nil
	}

	// Check if user is a project member
	member, err := s.memberRepo.GetMember(ctx, projectID, userID)
	if err != nil {
		return pkgerrors.ErrForbidden
	}

	if member == nil {
		return pkgerrors.ErrForbidden
	}

	return nil
}
