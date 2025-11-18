package cache

import (
	"context"
	"fmt"
)

// InvalidateProjectStats invalidates project statistics cache
func InvalidateProjectStats(ctx context.Context, cache Cache, projectID int) error {
	if cache == nil {
		return nil
	}

	key := BuildProjectStatsKey(projectID)
	return cache.Delete(ctx, key)
}

// InvalidateIssueStats invalidates issue statistics cache
func InvalidateIssueStats(ctx context.Context, cache Cache, projectID int) error {
	if cache == nil {
		return nil
	}

	key := BuildIssueStatsKey(projectID)
	return cache.Delete(ctx, key)
}

// InvalidateAllProjectCaches invalidates all caches related to a project
func InvalidateAllProjectCaches(ctx context.Context, cache Cache, projectID int) error {
	if cache == nil {
		return nil
	}

	// Invalidate project stats
	if err := InvalidateProjectStats(ctx, cache, projectID); err != nil {
		return fmt.Errorf("failed to invalidate project stats: %w", err)
	}

	// Invalidate issue stats
	if err := InvalidateIssueStats(ctx, cache, projectID); err != nil {
		return fmt.Errorf("failed to invalidate issue stats: %w", err)
	}

	// Invalidate search caches for the project
	pattern := fmt.Sprintf("%s:*project:%d*", PrefixSearchIssue, projectID)
	if err := cache.DeleteByPattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate search caches: %w", err)
	}

	return nil
}

// InvalidateUserStats invalidates user statistics cache
func InvalidateUserStats(ctx context.Context, cache Cache, userID int) error {
	if cache == nil {
		return nil
	}

	key := BuildUserStatsKey(userID)
	return cache.Delete(ctx, key)
}

// InvalidateSearchCaches invalidates all search-related caches
func InvalidateSearchCaches(ctx context.Context, cache Cache) error {
	if cache == nil {
		return nil
	}

	// Invalidate all issue search caches
	issuePattern := fmt.Sprintf("%s:*", PrefixSearchIssue)
	if err := cache.DeleteByPattern(ctx, issuePattern); err != nil {
		return fmt.Errorf("failed to invalidate issue search caches: %w", err)
	}

	// Invalidate all project search caches
	projectPattern := fmt.Sprintf("%s:*", PrefixSearchProject)
	if err := cache.DeleteByPattern(ctx, projectPattern); err != nil {
		return fmt.Errorf("failed to invalidate project search caches: %w", err)
	}

	return nil
}
