package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// SecretPanel represents the secret display panel
type SecretPanel struct {
	config        *config.Config
	vaultMgr      *vault.Manager
	textView      *tview.TextView
	currentSecret *vault.SecretNode
	logger        *logrus.Logger
}

// NewSecretPanel creates a new secret panel
func NewSecretPanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *SecretPanel {
	return &SecretPanel{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// Initialize initializes the secret panel
func (sp *SecretPanel) Initialize() error {
	sp.textView = tview.NewTextView()

	// Set up the text view appearance
	sp.textView.SetBorder(true).
		SetTitle("Secret Details").
		SetTitleAlign(tview.AlignLeft)

	// Enable dynamic colors and word wrap
	sp.textView.SetDynamicColors(true)
	sp.textView.SetWordWrap(true)
	sp.textView.SetScrollable(true)

	// Set up keyboard navigation
	sp.setupKeyboardNavigation()

	// Show initial message
	sp.showWelcomeMessage()

	return nil
}

// setupKeyboardNavigation sets up keyboard navigation for the secret panel
func (sp *SecretPanel) setupKeyboardNavigation() {
	sp.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				// Edit secret
				sp.editCurrentSecret()
				return nil
			case 'c':
				// Copy to clipboard
				sp.copyToClipboard()
				return nil
			case 'v':
				// Copy just the secret value to clipboard
				sp.copySecretValue()
				return nil
			}
		}

		return event
	})
}

// ShowSecret displays a secret
func (sp *SecretPanel) ShowSecret(path string) {
	sp.logger.Infof("Showing secret: %s", path)

	secretsManager, err := sp.vaultMgr.GetSecretsManager()
	if err != nil {
		sp.showError(fmt.Sprintf("Failed to get secrets manager: %v", err))
		return
	}

	// Get the secret
	secret, err := secretsManager.GetSecret(path)
	if err != nil {
		sp.showError(fmt.Sprintf("Failed to get secret: %v", err))
		return
	}

	// Store the current secret
	sp.currentSecret = secret

	// Format and display the secret
	sp.displaySecret(secret)
}

// ShowConnectionError displays a connection error message
func (sp *SecretPanel) ShowConnectionError(message string) {
	content := fmt.Sprintf(`[red]Connection Error[white]

%s

[yellow]Troubleshooting:[white]
• Check if your vault server is running
• Verify the vault address in your configuration
• Ensure network connectivity to the vault server
• Check your authentication credentials

[yellow]Press 'r' to refresh or Ctrl+C to exit[white]`, message)

	sp.textView.SetText(content)
}

// ShowDirectory displays directory information
func (sp *SecretPanel) ShowDirectory(path string) {
	sp.logger.Infof("Showing directory: %s", path)

	secretsManager, err := sp.vaultMgr.GetSecretsManager()
	if err != nil {
		sp.showError(fmt.Sprintf("Failed to get secrets manager: %v", err))
		return
	}

	// Get secrets in the directory
	secrets, err := secretsManager.ListSecrets(path)
	if err != nil {
		sp.showError(fmt.Sprintf("Failed to list directory: %v", err))
		return
	}

	// Format and display the directory
	sp.displayDirectory(path, secrets)
}

// displaySecret displays a secret in the text view
func (sp *SecretPanel) displaySecret(secret *vault.SecretNode) {
	var content strings.Builder

	// Header
	content.WriteString("[yellow]Secret: " + secret.Name + "[white]\n")
	content.WriteString("[gray]Path: " + secret.Path + "[white]\n")

	if secret.Metadata != nil {
		content.WriteString(fmt.Sprintf("[gray]Version: %d[white]\n", secret.Metadata.Version))
		content.WriteString(fmt.Sprintf("[gray]Created: %s[white]\n", secret.Metadata.CreatedTime.Format("2006-01-02 15:04:05")))
	}

	content.WriteString("\n")

	// Secret data
	content.WriteString("[yellow]Data:[white]\n")
	if secret.Data != nil {
		for key, value := range secret.Data {
			content.WriteString(fmt.Sprintf("[green]%s[white]: ", key))

			// Format the value
			switch v := value.(type) {
			case string:
				content.WriteString(v)
			case []byte:
				content.WriteString(string(v))
			default:
				content.WriteString(fmt.Sprintf("%v", v))
			}
			content.WriteString("\n")
		}
	} else {
		content.WriteString("[red]No data found[white]\n")
	}

	// Footer
	content.WriteString("\n[gray]Press 'e' to edit, 'c' to copy all, 'v' to copy value[white]")

	sp.textView.SetText(content.String())
}

// displayDirectory displays directory information
func (sp *SecretPanel) displayDirectory(path string, secrets []*vault.SecretNode) {
	var content strings.Builder

	// Header
	content.WriteString("[yellow]Directory: " + path + "[white]\n")
	content.WriteString(fmt.Sprintf("[gray]Items: %d[white]\n", len(secrets)))
	content.WriteString("\n")

	// Directory contents
	if len(secrets) == 0 {
		content.WriteString("[gray]Directory is empty[white]\n")
	} else {
		content.WriteString("[yellow]Contents:[white]\n")
		for _, secret := range secrets {
			if secret.IsSecret {
				content.WriteString("🔐 ")
			} else {
				content.WriteString("📁 ")
			}
			content.WriteString(secret.Name + "\n")
		}
	}

	// Footer
	content.WriteString("\n[gray]Select an item to view details[white]")

	sp.textView.SetText(content.String())
}

// showWelcomeMessage shows the welcome message
func (sp *SecretPanel) showWelcomeMessage() {
	content := `[yellow]Welcome to VUI - Vault UI[white]

This is the secret details panel. Select a secret from the tree to view its contents.

[yellow]Navigation:[white]
• Use arrow keys to navigate the tree
• Press Enter to select an item
• Press Tab to switch between panels

[yellow]Actions:[white]
• Press 'c' to create a new secret
• Press 'e' to edit the selected secret
• Press 'r' to refresh the view
• Press 's' to search secrets

[yellow]Vault Management:[white]
• Press Ctrl+v to switch vaults
• Press F1 for help
• Press Ctrl+C to exit

Select a secret from the tree to get started!`

	sp.textView.SetText(content)
}

// showError shows an error message
func (sp *SecretPanel) showError(message string) {
	content := fmt.Sprintf("[red]Error:[white]\n%s", message)
	sp.textView.SetText(content)
}

// editCurrentSecret edits the currently displayed secret
func (sp *SecretPanel) editCurrentSecret() {
	// This will be implemented in the forms component
	sp.logger.Info("Edit current secret")
}

// copyToClipboard copies the current secret to clipboard
func (sp *SecretPanel) copyToClipboard() {
	if sp.currentSecret == nil {
		sp.logger.Warn("No secret selected to copy")
		return
	}

	// Get the secret data
	secretsManager, err := sp.vaultMgr.GetSecretsManager()
	if err != nil {
		sp.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	secret, err := secretsManager.GetSecret(sp.currentSecret.Path)
	if err != nil {
		sp.logger.Errorf("Failed to get secret: %v", err)
		return
	}

	// Build clipboard content
	var clipboardContent strings.Builder
	clipboardContent.WriteString(fmt.Sprintf("Secret: %s\n", secret.Name))
	clipboardContent.WriteString(fmt.Sprintf("Path: %s\n", secret.Path))

	if secret.Metadata != nil {
		clipboardContent.WriteString("Metadata:\n")
		clipboardContent.WriteString(fmt.Sprintf("  Version: %d\n", secret.Metadata.Version))
		clipboardContent.WriteString(fmt.Sprintf("  Created: %s\n", secret.Metadata.CreatedTime.Format("2006-01-02 15:04:05")))
		if secret.Metadata.Destroyed {
			clipboardContent.WriteString("  Status: Destroyed\n")
		}
	}

	if secret.Data != nil {
		clipboardContent.WriteString("Data:\n")
		for key, value := range secret.Data {
			clipboardContent.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(clipboardContent.String()); err != nil {
		sp.logger.Errorf("Failed to copy to clipboard: %v", err)
		return
	}

	sp.logger.Infof("Copied secret '%s' to clipboard", secret.Name)

	// Show a brief success message
	sp.showCopySuccess()
}

// showCopySuccess shows a brief success message
func (sp *SecretPanel) showCopySuccess() {
	// Store the current text
	currentText := sp.textView.GetText(false)

	// Show success message
	sp.textView.SetText("[green]✓ Secret copied to clipboard![white]\n\n" + currentText)

	// Schedule to restore the original text after 2 seconds
	go func() {
		// This is a simple approach - in a real implementation, you might want to use
		// a more sophisticated approach with proper UI updates
		time.Sleep(2 * time.Second)
		sp.textView.SetText(currentText)
	}()
}

// copySecretValue copies just the secret value to clipboard
func (sp *SecretPanel) copySecretValue() {
	if sp.currentSecret == nil {
		sp.logger.Warn("No secret selected to copy")
		return
	}

	// Get the secret data
	secretsManager, err := sp.vaultMgr.GetSecretsManager()
	if err != nil {
		sp.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	secret, err := secretsManager.GetSecret(sp.currentSecret.Path)
	if err != nil {
		sp.logger.Errorf("Failed to get secret: %v", err)
		return
	}

	// If there's only one key-value pair, copy just the value
	if len(secret.Data) == 1 {
		for _, value := range secret.Data {
			valueStr := fmt.Sprintf("%v", value)
			if err := clipboard.WriteAll(valueStr); err != nil {
				sp.logger.Errorf("Failed to copy value to clipboard: %v", err)
				return
			}
			sp.logger.Infof("Copied secret value to clipboard")
			sp.showCopyValueSuccess()
			return
		}
	}

	// If there are multiple values, show a selection dialog
	sp.showValueSelectionDialog(secret)
}

// showValueSelectionDialog shows a dialog to select which value to copy
func (sp *SecretPanel) showValueSelectionDialog(secret *vault.SecretNode) {
	// Create a list of values
	list := tview.NewList()

	for key, value := range secret.Data {
		key, value := key, value // Capture loop variables
		valueStr := fmt.Sprintf("%v", value)
		list.AddItem(fmt.Sprintf("%s: %s", key, valueStr), "", 0, func() {
			if err := clipboard.WriteAll(valueStr); err != nil {
				sp.logger.Errorf("Failed to copy value to clipboard: %v", err)
				return
			}
			sp.logger.Infof("Copied value '%s' to clipboard", key)
			sp.showCopyValueSuccess()
		})
	}

	list.SetTitle("Select Value to Copy").
		SetBorder(true)

	list.SetDoneFunc(func() {
		// Return to main UI - this would need to be integrated with the modal system
		sp.logger.Info("Value selection cancelled")
	})

	// This would need to be integrated with the modal system
	sp.logger.Info("Showing value selection dialog")
}

// showCopyValueSuccess shows a brief success message for value copy
func (sp *SecretPanel) showCopyValueSuccess() {
	// Store the current text
	currentText := sp.textView.GetText(false)

	// Show success message
	sp.textView.SetText("[green]✓ Value copied to clipboard![white]\n\n" + currentText)

	// Schedule to restore the original text after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		sp.textView.SetText(currentText)
	}()
}

// Refresh refreshes the secret panel
func (sp *SecretPanel) Refresh() {
	sp.logger.Info("Refreshing secret panel")
	// For now, just show the welcome message
	sp.showWelcomeMessage()
}

// GetPrimitive returns the underlying tview primitive
func (sp *SecretPanel) GetPrimitive() tview.Primitive {
	return sp.textView
}
