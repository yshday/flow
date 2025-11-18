package cache

import "fmt"

// Cache key prefixes
const (
	PrefixProjectStats = "stats:project"
	PrefixIssueStats   = "stats:issue"
	PrefixUserStats    = "stats:user"
	PrefixSearchIssue  = "search:issue"
	PrefixSearchProject = "search:project"
)

// TTL durations
const (
	TTLProjectStats = 5 * 60  // 5 minutes
	TTLIssueStats   = 5 * 60  // 5 minutes
	TTLUserStats    = 3 * 60  // 3 minutes
	TTLSearch       = 2 * 60  // 2 minutes
)

// BuildProjectStatsKey builds a cache key for project statistics
func BuildProjectStatsKey(projectID int) string {
	return fmt.Sprintf("%s:%d", PrefixProjectStats, projectID)
}

// BuildIssueStatsKey builds a cache key for issue statistics
func BuildIssueStatsKey(projectID int) string {
	return fmt.Sprintf("%s:%d", PrefixIssueStats, projectID)
}

// BuildUserStatsKey builds a cache key for user statistics
func BuildUserStatsKey(userID int) string {
	return fmt.Sprintf("%s:%d", PrefixUserStats, userID)
}

// BuildSearchIssueKey builds a cache key for issue search results
func BuildSearchIssueKey(query string, filters map[string]string) string {
	key := fmt.Sprintf("%s:%s", PrefixSearchIssue, query)
	for k, v := range filters {
		key += fmt.Sprintf(":%s=%s", k, v)
	}
	return key
}

// BuildSearchProjectKey builds a cache key for project search results
func BuildSearchProjectKey(query string) string {
	return fmt.Sprintf("%s:%s", PrefixSearchProject, query)
}

// GetProjectInvalidationPattern returns pattern to invalidate all project-related caches
func GetProjectInvalidationPattern(projectID int) string {
	return fmt.Sprintf("*:project:%d:*", projectID)
}

// GetIssueInvalidationPattern returns pattern to invalidate issue-related caches
func GetIssueInvalidationPattern(projectID int) string {
	return fmt.Sprintf("*:issue*project:%d*", projectID)
}
