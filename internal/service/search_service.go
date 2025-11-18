package service

import (
	"context"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
)

// SearchService handles search business logic
type SearchService struct {
	searchRepo  *repository.SearchRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	cache       pkgcache.Cache
}

// NewSearchService creates a new search service
func NewSearchService(
	searchRepo *repository.SearchRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	cache pkgcache.Cache,
) *SearchService {
	return &SearchService{
		searchRepo:  searchRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		cache:       cache,
	}
}

// SearchIssues searches for issues
func (s *SearchService) SearchIssues(ctx context.Context, req *models.IssueSearchRequest, userID int) (*models.SearchResponse, error) {
	// If searching within a specific project, verify access
	if req.ProjectID != nil {
		// Check if user has access to the project
		project, err := s.projectRepo.GetByID(ctx, *req.ProjectID)
		if err != nil {
			return nil, err
		}

		// Check if user is owner or member
		if project.OwnerID != userID {
			member, err := s.memberRepo.GetMember(ctx, *req.ProjectID, userID)
			if err != nil || member == nil {
				// User doesn't have access, return empty results
				return &models.SearchResponse{
					Results: []*models.IssueSearchResult{},
					Total:   0,
					Limit:   req.Limit,
					Offset:  req.Offset,
				}, nil
			}
		}
	} else if len(req.ProjectIDs) == 0 {
		// If no project filter is specified, limit search to projects the user has access to
		// Get all projects accessible to the user (both owned and where they are a member)
		accessibleProjects, err := s.projectRepo.ListByUserID(ctx, userID)
		if err != nil {
			// If error fetching projects, return empty results
			return &models.SearchResponse{
				Results: []*models.IssueSearchResult{},
				Total:   0,
				Limit:   req.Limit,
				Offset:  req.Offset,
			}, nil
		}

		// If user has no accessible projects, return empty results
		if len(accessibleProjects) == 0 {
			return &models.SearchResponse{
				Results: []*models.IssueSearchResult{},
				Total:   0,
				Limit:   req.Limit,
				Offset:  req.Offset,
			}, nil
		}

		// Extract project IDs
		accessibleProjectIDs := make([]int, len(accessibleProjects))
		for i, project := range accessibleProjects {
			accessibleProjectIDs[i] = project.ID
		}

		// Set ProjectIDs to filter by accessible projects
		req.ProjectIDs = accessibleProjectIDs
	}

	results, total, err := s.searchRepo.SearchIssues(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.SearchResponse{
		Results: results,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// SearchProjects searches for projects
func (s *SearchService) SearchProjects(ctx context.Context, req *models.ProjectSearchRequest, userID int) (*models.SearchResponse, error) {
	// Search all projects (public ones or ones user has access to)
	// For now, we'll search all projects and filter accessible ones
	results, _, err := s.searchRepo.SearchProjects(ctx, req)
	if err != nil {
		return nil, err
	}

	// Filter results to only include projects the user has access to
	var accessibleResults []*models.ProjectSearchResult
	for _, project := range results {
		// Check if user is owner
		if project.OwnerID == userID {
			accessibleResults = append(accessibleResults, project)
			continue
		}

		// Check if user is a member
		member, err := s.memberRepo.GetMember(ctx, project.ID, userID)
		if err == nil && member != nil {
			accessibleResults = append(accessibleResults, project)
		}
	}

	return &models.SearchResponse{
		Results: accessibleResults,
		Total:   len(accessibleResults),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}
