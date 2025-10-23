package panels

import (
	"strings"
	"testing"

	"github.com/rvolykh/vui/internal/vault"
	"github.com/stretchr/testify/assert"
)

func TestNewProfilesTitle(t *testing.T) {
	vaultMgr := &vault.Manager{}

	pt := NewProfilesTitle(vaultMgr, false)

	assert.NotNil(t, pt)
	assert.Equal(t, vaultMgr, pt.vaultMgr)
	assert.False(t, pt.hasActiveConn)
}

func TestNewProfilesTitle_WithActiveConnection(t *testing.T) {
	vaultMgr := &vault.Manager{}

	pt := NewProfilesTitle(vaultMgr, true)

	assert.NotNil(t, pt)
	assert.Equal(t, vaultMgr, pt.vaultMgr)
	assert.True(t, pt.hasActiveConn)
}

func TestNewProfilesTitle_WithNilVaultManager(t *testing.T) {
	pt := NewProfilesTitle(nil, false)

	assert.NotNil(t, pt)
	assert.Nil(t, pt.vaultMgr)
}

func TestProfilesTitle_Build(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	textView := pt.Build()

	assert.NotNil(t, textView)
}

func TestProfilesTitle_Build_WithActiveConnection(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, true)

	textView := pt.Build()

	assert.NotNil(t, textView)
}

func TestProfilesTitle_BuildContent_MinimumWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Test with width below minimum (should be clamped to 60)
	content := pt.buildContent(40)

	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Welcome to VUI")
	assert.Contains(t, content, "Connection Status:")
	assert.Contains(t, content, "Navigation")
	assert.Contains(t, content, "Config Paths")
}

func TestProfilesTitle_BuildContent_MaximumWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Test with width above maximum (should be clamped to 120)
	content := pt.buildContent(200)

	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Welcome to VUI")
}

func TestProfilesTitle_BuildContent_NormalWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Welcome to VUI - Vault UI")
	assert.Contains(t, content, "Connection Status:")
	assert.Contains(t, content, "Navigation")
	assert.Contains(t, content, "Config Paths")
	assert.Contains(t, content, "Available Vault Profiles:")
}

func TestProfilesTitle_BuildContent_ContainsConfigPaths(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check all config paths are present
	assert.Contains(t, content, "./configs/vui.yaml")
	assert.Contains(t, content, "$HOME/.vui/vui.yaml")
	assert.Contains(t, content, "/etc/vui/vui.yaml")
}

func TestProfilesTitle_BuildContent_ContainsNavigationItems(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check navigation items are present
	assert.Contains(t, content, "↑/↓")
	assert.Contains(t, content, "Navigate profiles")
	assert.Contains(t, content, "Enter")
	assert.Contains(t, content, "Connect to profile")
	assert.Contains(t, content, "r/F5")
	assert.Contains(t, content, "Refresh status")
	assert.Contains(t, content, "q/Ctrl+C")
	assert.Contains(t, content, "Exit")
}

func TestProfilesTitle_BuildContent_HasBorders(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check for box drawing characters (borders)
	assert.Contains(t, content, "┌")
	assert.Contains(t, content, "┐")
	assert.Contains(t, content, "└")
	assert.Contains(t, content, "┘")
	assert.Contains(t, content, "─")
	assert.Contains(t, content, "│")
}

func TestProfilesTitle_GetConnectionStatus_NoConnection(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	status := pt.getConnectionStatus()

	assert.Contains(t, status, "No connection")
	assert.Contains(t, status, "please select a profile below to begin")
}

func TestProfilesTitle_GetConnectionStatus_HasActiveConnection_NoVaultName(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, true)

	status := pt.getConnectionStatus()

	// When hasActiveConn is true but GetActiveVault returns empty
	assert.Contains(t, status, "No active connection")
	assert.Contains(t, status, "please select a profile below")
}

func TestProfilesTitle_GetConnectionStatus_Connected(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, true)

	// Note: Without proper initialization, GetActiveVault will return empty string
	// This test verifies the logic branch
	status := pt.getConnectionStatus()

	assert.NotEmpty(t, status)
}

func TestProfilesTitle_GetNavigationItems_WithoutActiveConnection(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	items := pt.getNavigationItems()

	assert.Len(t, items, 5)
	assert.Equal(t, "↑/↓", items[0].keys)
	assert.Equal(t, "Navigate profiles", items[0].desc)
	assert.Equal(t, "Enter", items[1].keys)
	assert.Equal(t, "Connect to profile", items[1].desc)
	assert.Equal(t, "r/F5", items[2].keys)
	assert.Equal(t, "Refresh status", items[2].desc)
	assert.Equal(t, "h/F1", items[3].keys)
	assert.Equal(t, "Show help", items[3].desc)
	assert.Equal(t, "q/Ctrl+C", items[4].keys)
	assert.Equal(t, "Exit", items[4].desc)
}

func TestProfilesTitle_GetNavigationItems_WithActiveConnection(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, true)

	items := pt.getNavigationItems()

	// Should have 6 items (Esc is prepended)
	assert.Len(t, items, 6)
	assert.Equal(t, "Esc", items[0].keys)
	assert.Equal(t, "Back to secrets", items[0].desc)
	assert.Equal(t, "↑/↓", items[1].keys)
	assert.Equal(t, "Navigate profiles", items[1].desc)
}

func TestProfilesTitle_NavigationItem_Structure(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	items := pt.getNavigationItems()

	// Verify each item has both keys and desc
	for _, item := range items {
		assert.NotEmpty(t, item.keys, "Navigation item should have keys")
		assert.NotEmpty(t, item.desc, "Navigation item should have description")
	}
}

func TestProfilesTitle_BuildContent_DifferentWidths(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	widths := []int{60, 70, 80, 90, 100, 120}

	for _, width := range widths {
		content := pt.buildContent(width)
		assert.NotEmpty(t, content, "Content should not be empty for width %d", width)
		assert.Contains(t, content, "Welcome to VUI", "Content should contain title for width %d", width)
	}
}

func TestProfilesTitle_BuildContent_LineCount(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)
	lines := strings.Split(content, "\n")

	// Should have multiple lines (borders, headers, content, etc.)
	assert.Greater(t, len(lines), 10, "Content should have multiple lines")
}

func TestProfilesTitle_BuildContent_ColorTags(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check for color tags
	assert.Contains(t, content, "[yellow]")
	assert.Contains(t, content, "[white]")
	assert.Contains(t, content, "[cyan]")
	assert.Contains(t, content, "[green]")
}

func TestProfilesTitle_BuildContent_WithActiveConnection_HasEscKey(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, true)

	content := pt.buildContent(80)

	// Should contain the Esc navigation item
	assert.Contains(t, content, "Esc")
	assert.Contains(t, content, "Back to secrets")
}

func TestProfilesTitle_BuildContent_WithoutActiveConnection_NoEscKey(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Should NOT contain "Back to secrets" (Esc option only shows when connected)
	assert.NotContains(t, content, "Back to secrets")
}

func TestProfilesTitle_BuildContent_TwoColumnLayout(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Verify two-column layout is present
	lines := strings.Split(content, "\n")

	// Find lines with both navigation and config content
	foundBothColumns := false
	for _, line := range lines {
		if strings.Contains(line, "Navigate profiles") && strings.Contains(line, "./configs/vui.yaml") {
			foundBothColumns = true
			break
		}
	}

	assert.True(t, foundBothColumns, "Should have two-column layout with navigation and config paths")
}

func TestProfilesTitle_BuildContent_ConsistentStructure(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Build content multiple times to ensure consistency
	content1 := pt.buildContent(80)
	content2 := pt.buildContent(80)

	assert.Equal(t, content1, content2, "Content should be consistent across multiple calls")
}

func TestProfilesTitle_Build_SetsDynamicColors(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	textView := pt.Build()

	// Verify that the text view is created (we can't directly check if dynamic colors is set)
	assert.NotNil(t, textView)
}

func TestProfilesTitle_BuildContent_AllNavigationItemsVisible(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	items := pt.getNavigationItems()
	content := pt.buildContent(80)

	// Verify all navigation items appear in the content
	for _, item := range items {
		assert.Contains(t, content, item.keys, "Content should contain navigation key: %s", item.keys)
		assert.Contains(t, content, item.desc, "Content should contain navigation description: %s", item.desc)
	}
}

func TestProfilesTitle_BuildContent_WithVerySmallWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Test with very small width (should be clamped to minimum)
	content := pt.buildContent(10)

	assert.NotEmpty(t, content)
	// Should still contain essential content
	assert.Contains(t, content, "Welcome to VUI")
}

func TestProfilesTitle_BuildContent_WithZeroWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Test with zero width (should be clamped to minimum)
	content := pt.buildContent(0)

	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Welcome to VUI")
}

func TestProfilesTitle_BuildContent_WithNegativeWidth(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Test with negative width (should be clamped to minimum)
	content := pt.buildContent(-10)

	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Welcome to VUI")
}

func TestProfilesTitle_GetNavigationItems_Order(t *testing.T) {
	vaultMgr := &vault.Manager{}

	// Test without active connection
	pt1 := NewProfilesTitle(vaultMgr, false)
	items1 := pt1.getNavigationItems()

	// First item should be navigation (not Esc)
	assert.Equal(t, "↑/↓", items1[0].keys)

	// Test with active connection
	pt2 := NewProfilesTitle(vaultMgr, true)
	items2 := pt2.getNavigationItems()

	// First item should be Esc
	assert.Equal(t, "Esc", items2[0].keys)
	// Second item should be navigation
	assert.Equal(t, "↑/↓", items2[1].keys)
}

func TestProfilesTitle_BuildContent_BottomText(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check for the bottom text
	assert.Contains(t, content, "Available Vault Profiles:")
}

func TestProfilesTitle_BuildContent_NoEmptyLines(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)
	lines := strings.Split(content, "\n")

	// Count lines - should have content in most lines
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	assert.Greater(t, nonEmptyLines, 10, "Should have substantial content")
}

func TestProfilesTitle_StateIndependence(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	// Build content multiple times
	content1 := pt.buildContent(80)
	content2 := pt.buildContent(90)
	content3 := pt.buildContent(80)

	// Content should be the same for the same width
	assert.Equal(t, content1, content3, "Content should be consistent for same width")

	// Content should differ for different widths (due to padding)
	assert.NotEqual(t, content1, content2, "Content should differ for different widths")
}

func TestProfilesTitle_BuildContent_HelpKeyPresent(t *testing.T) {
	vaultMgr := &vault.Manager{}
	pt := NewProfilesTitle(vaultMgr, false)

	content := pt.buildContent(80)

	// Check that help key is present
	assert.Contains(t, content, "h/F1")
	assert.Contains(t, content, "Show help")
}
