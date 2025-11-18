package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
)

func setupProjectMemberRepo(t *testing.T) (*ProjectMemberRepository, *ProjectRepository, *UserRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM project_members")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	memberRepo := NewProjectMemberRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM project_members")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return memberRepo, projectRepo, userRepo, cleanup
}

func createTestUserForMember(t *testing.T, repo *UserRepository, email string) *models.User {
	user := &models.User{
		Email:        email,
		Username:     email,
		PasswordHash: "hashedpassword",
	}

	created, err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return created
}

func createTestProjectForMember(t *testing.T, projectRepo *ProjectRepository, userRepo *UserRepository) (*models.Project, *models.User) {
	owner := createTestUserForMember(t, userRepo, "owner@example.com")

	project := &models.Project{
		Name:        "Test Project",
		Key:         "TEST",
		Description: stringPtr("Test description"),
		OwnerID:     owner.ID,
	}

	created, err := projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return created, owner
}

func TestProjectMemberRepository_AddMember(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	newUser := createTestUserForMember(t, userRepo, "member@example.com")

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    newUser.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	}

	err := memberRepo.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Verify member was added
	members, err := memberRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	found := false
	for _, m := range members {
		if m.UserID == newUser.ID {
			found = true
			if m.Role != string(models.RoleMember) {
				t.Errorf("Expected role 'member', got '%s'", m.Role)
			}
			if m.InvitedBy == nil || *m.InvitedBy != owner.ID {
				t.Errorf("Expected invited_by %d, got %v", owner.ID, m.InvitedBy)
			}
		}
	}

	if !found {
		t.Error("Member not found in project members list")
	}
}

func TestProjectMemberRepository_AddMember_Duplicate(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	newUser := createTestUserForMember(t, userRepo, "member@example.com")

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    newUser.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	}

	// Add member first time
	err := memberRepo.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("First AddMember failed: %v", err)
	}

	// Try to add same member again (should fail due to PRIMARY KEY constraint)
	err = memberRepo.AddMember(ctx, member)
	if err == nil {
		t.Error("Expected error when adding duplicate member, got nil")
	}
}

func TestProjectMemberRepository_ListByProjectID(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	member1 := createTestUserForMember(t, userRepo, "member1@example.com")
	member2 := createTestUserForMember(t, userRepo, "member2@example.com")

	// Add owner as project member
	ownerMember := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    owner.ID,
		Role:      string(models.RoleOwner),
	}
	err := memberRepo.AddMember(ctx, ownerMember)
	if err != nil {
		t.Fatalf("Failed to add owner: %v", err)
	}

	// Add other members
	err = memberRepo.AddMember(ctx, &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    member1.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	})
	if err != nil {
		t.Fatalf("Failed to add member1: %v", err)
	}

	err = memberRepo.AddMember(ctx, &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    member2.ID,
		Role:      string(models.RoleAdmin),
		InvitedBy: &owner.ID,
	})
	if err != nil {
		t.Fatalf("Failed to add member2: %v", err)
	}

	// List all members
	members, err := memberRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Verify User field is populated
	for _, m := range members {
		if m.User == nil {
			t.Error("Expected User to be populated, got nil")
		} else if m.User.Email == "" {
			t.Error("Expected User.Email to be set")
		}
	}
}

func TestProjectMemberRepository_GetMember(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	newUser := createTestUserForMember(t, userRepo, "member@example.com")

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    newUser.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	}

	err := memberRepo.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Get member
	retrieved, err := memberRepo.GetMember(ctx, project.ID, newUser.ID)
	if err != nil {
		t.Fatalf("GetMember failed: %v", err)
	}

	if retrieved.ProjectID != project.ID {
		t.Errorf("Expected ProjectID %d, got %d", project.ID, retrieved.ProjectID)
	}

	if retrieved.UserID != newUser.ID {
		t.Errorf("Expected UserID %d, got %d", newUser.ID, retrieved.UserID)
	}

	if retrieved.Role != string(models.RoleMember) {
		t.Errorf("Expected role 'member', got '%s'", retrieved.Role)
	}

	// Verify joined_at is set
	if retrieved.JoinedAt.IsZero() {
		t.Error("Expected JoinedAt to be set")
	}
}

func TestProjectMemberRepository_GetMember_NotFound(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMember(t, projectRepo, userRepo)

	// Try to get non-existent member
	_, err := memberRepo.GetMember(ctx, project.ID, 99999)
	if err == nil {
		t.Error("Expected error for non-existent member, got nil")
	}
}

func TestProjectMemberRepository_UpdateRole(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	newUser := createTestUserForMember(t, userRepo, "member@example.com")

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    newUser.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	}

	err := memberRepo.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Update role
	err = memberRepo.UpdateRole(ctx, project.ID, newUser.ID, string(models.RoleAdmin))
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	// Verify role was updated
	updated, err := memberRepo.GetMember(ctx, project.ID, newUser.ID)
	if err != nil {
		t.Fatalf("GetMember failed: %v", err)
	}

	if updated.Role != string(models.RoleAdmin) {
		t.Errorf("Expected role 'admin', got '%s'", updated.Role)
	}
}

func TestProjectMemberRepository_RemoveMember(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, owner := createTestProjectForMember(t, projectRepo, userRepo)
	newUser := createTestUserForMember(t, userRepo, "member@example.com")

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    newUser.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner.ID,
	}

	err := memberRepo.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// Remove member
	err = memberRepo.RemoveMember(ctx, project.ID, newUser.ID)
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	// Verify member was removed
	_, err = memberRepo.GetMember(ctx, project.ID, newUser.ID)
	if err == nil {
		t.Error("Expected error after removing member, got nil")
	}
}

func TestProjectMemberRepository_ListByUserID(t *testing.T) {
	memberRepo, projectRepo, userRepo, cleanup := setupProjectMemberRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple projects
	project1, owner1 := createTestProjectForMember(t, projectRepo, userRepo)

	owner2 := createTestUserForMember(t, userRepo, "owner2@example.com")
	project2 := &models.Project{
		Name:        "Test Project 2",
		Key:         "TEST2",
		Description: stringPtr("Test description 2"),
		OwnerID:     owner2.ID,
	}
	project2, err := projectRepo.Create(ctx, project2)
	if err != nil {
		t.Fatalf("Failed to create project2: %v", err)
	}

	// Create a user that will be member of both projects
	member := createTestUserForMember(t, userRepo, "member@example.com")

	// Add member to project1
	err = memberRepo.AddMember(ctx, &models.ProjectMember{
		ProjectID: project1.ID,
		UserID:    member.ID,
		Role:      string(models.RoleMember),
		InvitedBy: &owner1.ID,
	})
	if err != nil {
		t.Fatalf("Failed to add member to project1: %v", err)
	}

	// Add member to project2
	err = memberRepo.AddMember(ctx, &models.ProjectMember{
		ProjectID: project2.ID,
		UserID:    member.ID,
		Role:      string(models.RoleAdmin),
		InvitedBy: &owner2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to add member to project2: %v", err)
	}

	// List projects for member
	projects, err := memberRepo.ListByUserID(ctx, member.ID)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	// Verify project information is populated
	for _, pm := range projects {
		if pm.ProjectID != project1.ID && pm.ProjectID != project2.ID {
			t.Errorf("Unexpected project ID %d", pm.ProjectID)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
