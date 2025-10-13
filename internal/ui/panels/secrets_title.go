package panels

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// SecretsTitle represents the navigation secrets title
type SecretsTitle struct {
	config   *config.Config
	vaultMgr *vault.Manager
	textView *tview.TextView
	logger   *logrus.Logger
}

// NewSecretsTitle creates a new secrets title
func NewSecretsTitle(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *SecretsTitle {
	return &SecretsTitle{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// Initialize initializes the secrets title
func (hp *SecretsTitle) Initialize() error {
	hp.textView = tview.NewTextView()

	// Set up the text view appearance
	hp.textView.SetBorder(true).
		SetTitle("Navigation & Controls").
		SetTitleAlign(tview.AlignLeft)

	// Enable dynamic colors
	hp.textView.SetDynamicColors(true)

	// Set the help content
	hp.updateHelpText()

	return nil
}

// updateHelpText updates the help text
func (hp *SecretsTitle) updateHelpText() {
	// Get current vault connection info
	vaultInfo := hp.getVaultInfo()

	// Create a formatted help text with 4 equal columns: Navigation, Secrets, Values, Global
	helpText := fmt.Sprintf(`[yellow]Connected to Vault:[white] %s

[yellow]Navigation[white]                [yellow]Secrets[white]                  [yellow]Values[white]                   [yellow]Global[white]
  ↑/↓: Move             c: Create new            d: Toggle mask/unmask   h/F1: Help
  ←/→: Collapse/Expand  e: Edit selected         v: Copy to clipboard    r/F5: Refresh
  Enter: Select/Expand  Ctrl+d: Delete selected                          Tab/Ctrl+v: Profiles
                                                                         q/Ctrl+C: Quit`, vaultInfo)

	hp.textView.SetText(helpText)
}

// getVaultInfo returns the current vault connection information
func (hp *SecretsTitle) getVaultInfo() string {
	if hp.vaultMgr == nil {
		return "[red]No connection[white]"
	}

	currentVault := hp.vaultMgr.GetActiveVault()
	if currentVault == "" {
		return "[red]No vault selected[white]"
	}

	// Get vault profile for address
	if profile, ok := hp.config.Vaults[currentVault]; ok {
		return fmt.Sprintf("[green]%s[white] (%s)", currentVault, profile.Address)
	}

	return fmt.Sprintf("[green]%s[white]", currentVault)
}

// UpdateVaultInfo updates the vault connection information
func (hp *SecretsTitle) UpdateVaultInfo() {
	hp.updateHelpText()
}

// GetPrimitive returns the underlying tview primitive
func (hp *SecretsTitle) GetPrimitive() tview.Primitive {
	return hp.textView
}
