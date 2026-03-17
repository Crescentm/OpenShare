package model

import "testing"

func TestNormalizeAdminPermissions(t *testing.T) {
	normalized := NormalizeAdminPermissions([]AdminPermission{
		AdminPermissionManageSystem,
		AdminPermissionReviewSubmissions,
		AdminPermissionManageSystem,
		AdminPermission("invalid"),
	})

	expected := "manage_system,submission_moderation"
	if normalized != expected {
		t.Fatalf("expected normalized permissions %q, got %q", expected, normalized)
	}
}

func TestParseAdminPermissions(t *testing.T) {
	permissions := ParseAdminPermissions("manage_system, review_submissions,manage_system,invalid")

	if len(permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %v", permissions)
	}
	if permissions[0] != AdminPermissionManageSystem {
		t.Fatalf("expected first permission %q, got %q", AdminPermissionManageSystem, permissions[0])
	}
	if permissions[1] != AdminPermissionSubmissionModeration {
		t.Fatalf("expected second permission %q, got %q", AdminPermissionSubmissionModeration, permissions[1])
	}
}

func TestValidateAdminRoleAndStatus(t *testing.T) {
	if err := ValidateAdminRole(string(AdminRoleSuperAdmin)); err != nil {
		t.Fatalf("expected valid super admin role, got %v", err)
	}
	if err := ValidateAdminRole("unknown"); err == nil {
		t.Fatal("expected invalid role error")
	}

	if err := ValidateAdminStatus(AdminStatusActive); err != nil {
		t.Fatalf("expected valid active status, got %v", err)
	}
	if err := ValidateAdminStatus(AdminStatus("archived")); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestDefaultAdminPermissions(t *testing.T) {
	adminPermissions := DefaultAdminPermissions(AdminRoleAdmin)
	if len(adminPermissions) != 1 || adminPermissions[0] != AdminPermissionSubmissionModeration {
		t.Fatalf("expected default admin permission set to contain submission_moderation, got %v", adminPermissions)
	}

	superAdminPermissions := DefaultAdminPermissions(AdminRoleSuperAdmin)
	if len(superAdminPermissions) != 0 {
		t.Fatalf("expected super admin explicit permission list to be empty, got %v", superAdminPermissions)
	}
}

func TestAdminHasPermission(t *testing.T) {
	admin := Admin{
		Role:        string(AdminRoleAdmin),
		Permissions: NormalizeAdminPermissions([]AdminPermission{AdminPermissionManageSystem}),
		Status:      AdminStatusActive,
	}

	if !admin.HasPermission(AdminPermissionManageSystem) {
		t.Fatal("expected admin to have manage_system permission")
	}
	if admin.HasPermission(AdminPermissionDirectUpload) {
		t.Fatal("expected admin to not have direct_upload permission")
	}

	superAdmin := Admin{
		Role:   string(AdminRoleSuperAdmin),
		Status: AdminStatusActive,
	}
	if !superAdmin.HasPermission(AdminPermissionManageSystem) {
		t.Fatal("expected super admin to bypass permission checks")
	}
}
