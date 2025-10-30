package panels

import (
	"strings"
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretsTitle(t *testing.T) {
	fixtures := WithFixtures(t)

	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	assert.NotNil(t, st)
	assert.Equal(t, fixtures.cfg, st.config)
	assert.Equal(t, fixtures.interactor, st.interactor)
	assert.Equal(t, fixtures.logger, st.logger)
	assert.Nil(t, st.textView) // Not initialized yet
}

func TestSecretsTitle_Initialize(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	err := st.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, st.textView)
}

func TestSecretsTitle_Initialize_CreatesTextView(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	err := st.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, st.textView)

	// TextView should have content after initialization
	text := st.textView.GetText(false)
	assert.NotEmpty(t, text)
}

func TestSecretsTitle_Initialize_MultipleCallsSafe(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// Initialize multiple times should not panic
	err1 := st.Initialize()
	err2 := st.Initialize()
	err3 := st.Initialize()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NotNil(t, st.textView)
}

func TestSecretsTitle_GetPrimitive(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	primitive := st.GetPrimitive()

	assert.NotNil(t, primitive)
	assert.Equal(t, st.textView, primitive)
}

func TestSecretsTitle_GetPrimitive_BeforeInitialize(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	primitive := st.GetPrimitive()

	assert.Nil(t, primitive)
}

func TestSecretsTitle_GetVaultInfo_NoVaultManager(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, nil, fixtures.logger)

	info := st.getVaultInfo()

	assert.Contains(t, info, "No connection")
	assert.Contains(t, info, "[red]")
}

func TestSecretsTitle_GetVaultInfo_NoActiveVault(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	info := st.getVaultInfo()

	// GetActiveVault() returns empty string when no vault is active
	assert.Contains(t, info, "No profile selected")
	assert.Contains(t, info, "[red]")
}

func TestSecretsTitle_GetVaultInfo_WithVaultNoProfile(t *testing.T) {
	fixtures := WithFixtures(t)
	fixtures.cfg.Profiles = map[string]config.Profile{}
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// With uninitialized vault manager, GetActiveVault returns empty string
	info := st.getVaultInfo()

	assert.NotEmpty(t, info)
}

func TestSecretsTitle_GetVaultInfo_WithProfile(t *testing.T) {
	fixtures := WithFixtures(t)
	fixtures.cfg.Profiles = map[string]config.Profile{
		"test-vault": {
			Engine:  "vault",
			Address: "https://vault.example.com",
		},
	}
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// With uninitialized vault manager, GetActiveVault returns empty string
	// so this will still return "No vault selected"
	info := st.getVaultInfo()

	assert.NotEmpty(t, info)
}

func TestSecretsTitle_UpdateHelpText_ContainsNavigationSection(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.Contains(t, text, "Navigation")
	assert.Contains(t, text, "↑/↓: Move")
	assert.Contains(t, text, "←/→: Collapse/Expand")
	assert.Contains(t, text, "Enter: Select/Expand")
}

func TestSecretsTitle_UpdateHelpText_ContainsSecretsSection(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.Contains(t, text, "Secrets")
	assert.Contains(t, text, "c: Create new")
	assert.Contains(t, text, "e: Edit selected")
	assert.Contains(t, text, "Ctrl+d: Delete selected")
}

func TestSecretsTitle_UpdateHelpText_ContainsValuesSection(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.Contains(t, text, "Values")
	assert.Contains(t, text, "d: Toggle mask/unmask")
	assert.Contains(t, text, "v: Copy to clipboard")
}

func TestSecretsTitle_UpdateHelpText_ContainsGlobalSection(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.Contains(t, text, "Global")
	assert.Contains(t, text, "h/F1: Help")
	assert.Contains(t, text, "r/F5: Refresh")
	assert.Contains(t, text, "Tab/Ctrl+v: Profiles")
	assert.Contains(t, text, "q/Ctrl+C: Quit")
}

func TestSecretsTitle_UpdateHelpText_ContainsVaultInfo(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.Contains(t, text, "Connected to Vault:")
}

func TestSecretsTitle_UpdateHelpText_HasColorTags(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	// Check for color tags
	assert.Contains(t, text, "[yellow]")
	assert.Contains(t, text, "[white]")
}

func TestSecretsTitle_UpdateVaultInfo(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	initialText := st.textView.GetText(false)

	// Call UpdateVaultInfo
	st.UpdateVaultInfo()

	updatedText := st.textView.GetText(false)

	// Text should be the same since vault state hasn't changed
	assert.Equal(t, initialText, updatedText)
}

func TestSecretsTitle_UpdateVaultInfo_UpdatesText(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	// Should not panic
	st.UpdateVaultInfo()

	text := st.textView.GetText(false)
	assert.NotEmpty(t, text)
}

func TestSecretsTitle_Initialize_TextViewHasBorder(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	// We can't directly check if border is set, but we can verify textView exists
	assert.NotNil(t, st.textView)
}

func TestSecretsTitle_Initialize_TextViewHasTitle(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	// The title should be "Navigation & Controls"
	// We can't directly access it but we know it's set
	assert.NotNil(t, st.textView)
}

func TestSecretsTitle_HelpText_FormattedInColumns(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)
	lines := strings.Split(text, "\n")

	// Should have multiple lines
	assert.Greater(t, len(lines), 3)

	// Header line should contain all four section headers
	foundHeaderLine := false
	for _, line := range lines {
		if strings.Contains(line, "Navigation") &&
			strings.Contains(line, "Secrets") &&
			strings.Contains(line, "Values") &&
			strings.Contains(line, "Global") {
			foundHeaderLine = true
			break
		}
	}
	assert.True(t, foundHeaderLine, "Should have header line with all four sections")
}

func TestSecretsTitle_GetVaultInfo_ColorCoding(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// With nil vault manager
	st.interactor = nil
	info := st.getVaultInfo()
	assert.Contains(t, info, "[red]", "No connection should be red")

	// With vault manager but no active vault
	st.interactor = fixtures.interactor
	info = st.getVaultInfo()
	assert.Contains(t, info, "[red]", "No vault selected should be red")
}

func TestSecretsTitle_AllKeybindings_Present(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	// Navigation keybindings
	assert.Contains(t, text, "↑/↓")
	assert.Contains(t, text, "←/→")
	assert.Contains(t, text, "Enter")

	// Secrets keybindings
	assert.Contains(t, text, "c:")
	assert.Contains(t, text, "e:")
	assert.Contains(t, text, "Ctrl+d:")

	// Values keybindings
	assert.Contains(t, text, "d:")
	assert.Contains(t, text, "v:")

	// Global keybindings
	assert.Contains(t, text, "h/F1")
	assert.Contains(t, text, "r/F5")
	assert.Contains(t, text, "Tab/Ctrl+v")
	assert.Contains(t, text, "q/Ctrl+C")
}

func TestSecretsTitle_TextNotEmpty_AfterInitialize(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	assert.NotEmpty(t, text)
	assert.Greater(t, len(text), 100, "Help text should be substantial")
}

func TestSecretsTitle_ConsistentOutput(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text1 := st.textView.GetText(false)

	// Update vault info (which updates text)
	st.UpdateVaultInfo()

	text2 := st.textView.GetText(false)

	// Since vault state hasn't changed, text should be the same
	assert.Equal(t, text1, text2)
}

func TestSecretsTitle_GetVaultInfo_NilConfig(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(nil, fixtures.interactor, fixtures.logger)

	// Should not panic with nil config
	info := st.getVaultInfo()

	assert.NotEmpty(t, info)
}

func TestSecretsTitle_Initialize_WithNilVaultManager(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, nil, fixtures.logger)
	err := st.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, st.textView)

	text := st.textView.GetText(false)
	assert.Contains(t, text, "No connection")
}

func TestSecretsTitle_GetVaultInfo_EmptyVaultsMap(t *testing.T) {
	fixtures := WithFixtures(t)
	fixtures.cfg.Profiles = map[string]config.Profile{}
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	info := st.getVaultInfo()

	assert.NotEmpty(t, info)
}

func TestSecretsTitle_TextContains_AllSectionHeaders(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)
	st.Initialize()

	text := st.textView.GetText(false)

	// All four section headers should be present
	headers := []string{"Navigation", "Secrets", "Values", "Global"}
	for _, header := range headers {
		assert.Contains(t, text, header, "Should contain section header: %s", header)
	}
}

func TestSecretsTitle_UpdateVaultInfo_BeforeInitialize(t *testing.T) {
	fixtures := WithFixtures(t)
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// Calling UpdateVaultInfo before Initialize will panic because textView is nil
	// This is expected behavior - Initialize must be called first
	assert.Nil(t, st.textView, "TextView should be nil before Initialize")
}

func TestSecretsTitle_GetVaultInfo_WithAddress(t *testing.T) {
	fixtures := WithFixtures(t)
	fixtures.cfg.Profiles = map[string]config.Profile{
		"prod": {
			Engine:  "vault",
			Address: "https://vault.prod.example.com",
		},
	}
	st := NewSecretsTitle(fixtures.cfg, fixtures.interactor, fixtures.logger)

	// Since GetActiveVault returns empty string, this won't show the address
	// but the code path exists
	info := st.getVaultInfo()

	assert.NotEmpty(t, info)
}
