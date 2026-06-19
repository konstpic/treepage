package service

import "testing"

func TestMaxSpaceRole(t *testing.T) {
	if got := MaxSpaceRole("viewer", "editor"); got != "editor" {
		t.Fatalf("expected editor, got %q", got)
	}
	if got := MaxSpaceRole("admin", "viewer"); got != "admin" {
		t.Fatalf("expected admin, got %q", got)
	}
	if got := MaxSpaceRole("", "viewer"); got != "viewer" {
		t.Fatalf("expected viewer, got %q", got)
	}
}

func TestGlobalRolesAsSpaceRole(t *testing.T) {
	if got := GlobalRolesAsSpaceRole([]string{"viewer", "editor"}); got != SpaceRoleEditor {
		t.Fatalf("expected editor, got %q", got)
	}
	if got := GlobalRolesAsSpaceRole([]string{"admin"}); got != SpaceRoleAdmin {
		t.Fatalf("expected admin, got %q", got)
	}
}

func TestCanEditInSpace(t *testing.T) {
	if !CanEditInSpace(SpaceRoleEditor, []string{"viewer"}) {
		t.Fatal("space editor should edit")
	}
	if CanEditInSpace(SpaceRoleViewer, []string{"viewer"}) {
		t.Fatal("space viewer should not edit")
	}
	if !CanEditInSpace("", []string{"super_admin"}) {
		t.Fatal("super_admin should edit")
	}
}
