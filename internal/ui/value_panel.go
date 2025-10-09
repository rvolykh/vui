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

// ValuePanel represents the secret value display panel
type ValuePanel struct {
	config        *config.Config
	vaultMgr      *vault.Manager
	textView      *tview.TextView
	currentSecret *vault.SecretNode
	currentKey    string
	logger        *logrus.Logger
}

// NewValuePanel creates a new value panel
func NewValuePanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *ValuePanel {
	return &ValuePanel{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// Initialize initializes the value panel
func (vp *ValuePanel) Initialize() error {
	vp.textView = tview.NewTextView()

	// Set up the text view appearance
	vp.textView.SetBorder(true).
		SetTitle("Secret Value").
		SetTitleAlign(tview.AlignLeft)

	// Enable dynamic colors and word wrap
	vp.textView.SetDynamicColors(true)
	vp.textView.SetWordWrap(true)
	vp.textView.SetScrollable(true)

	// Set up keyboard navigation
	vp.setupKeyboardNavigation()

	// Show initial message
	vp.showDefaultMessage()

	return nil
}

// setupKeyboardNavigation sets up keyboard navigation for the value panel
func (vp *ValuePanel) setupKeyboardNavigation() {
	vp.textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'c':
				// Copy to clipboard
				vp.copyToClipboard()
				return nil
			case 'v':
				// Copy just the value to clipboard
				vp.copySecretValue()
				return nil
			}
		}

		return event
	})
}

// ShowSecret displays all key-value pairs of a secret
func (vp *ValuePanel) ShowSecret(secret *vault.SecretNode) {
	vp.currentSecret = secret
	vp.currentKey = ""
	vp.displaySecretData(secret)
}

// ShowKey displays a specific key's value from a secret
func (vp *ValuePanel) ShowKey(secret *vault.SecretNode, key string) {
	vp.currentSecret = secret
	vp.currentKey = key
	vp.displayKeyValue(secret, key)
}

// ShowDirectory displays directory information
func (vp *ValuePanel) ShowDirectory(path string) {
	vp.currentSecret = nil
	vp.currentKey = ""
	content := fmt.Sprintf(`[yellow]Directory:[white] %s

[gray]This is a directory containing secrets.
Select a secret to view its contents.[white]`, path)
	vp.textView.SetText(content)
}

// displaySecretData displays all key-value pairs of a secret
func (vp *ValuePanel) displaySecretData(secret *vault.SecretNode) {
	var content strings.Builder

	content.WriteString("[yellow]Secret Data:[white]\n\n")

	if secret.Data != nil && len(secret.Data) > 0 {
		for key, value := range secret.Data {
			content.WriteString(fmt.Sprintf("[green]%s:[white]\n", key))

			// Format the value
			valueStr := vp.formatValue(value)
			content.WriteString(valueStr)
			content.WriteString("\n\n")
		}

		content.WriteString("[gray]Press 'c' to copy all, 'v' to copy value, expand keys to view individually[white]")
	} else {
		content.WriteString("[red]No data found[white]\n")
	}

	vp.textView.SetText(content.String())
}

// displayKeyValue displays a specific key's value
func (vp *ValuePanel) displayKeyValue(secret *vault.SecretNode, key string) {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("[yellow]Key:[white] [green]%s[white]\n\n", key))

	if secret.Data != nil {
		if value, ok := secret.Data[key]; ok {
			valueStr := vp.formatValue(value)
			content.WriteString(valueStr)
			content.WriteString("\n\n")
			content.WriteString("[gray]Press 'v' to copy this value[white]")
		} else {
			content.WriteString("[red]Key not found[white]\n")
		}
	} else {
		content.WriteString("[red]No data found[white]\n")
	}

	vp.textView.SetText(content.String())
}

// formatValue formats a value for display
func (vp *ValuePanel) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// showDefaultMessage shows the default message
func (vp *ValuePanel) showDefaultMessage() {
	content := `[yellow]Secret Value[white]

Select a secret from the tree to view its values.

[yellow]Features:[white]
• View all key-value pairs
• Expand secrets to view individual keys
• Copy values to clipboard
• Search within values

[yellow]Navigation:[white]
• Select a secret to view all its data
• Expand a secret to see individual keys
• Select a key to view its specific value`

	vp.textView.SetText(content)
}

// copyToClipboard copies the current content to clipboard
func (vp *ValuePanel) copyToClipboard() {
	if vp.currentSecret == nil {
		vp.logger.Warn("No secret selected to copy")
		return
	}

	var clipboardContent strings.Builder

	if vp.currentKey != "" {
		// Copy specific key value
		if value, ok := vp.currentSecret.Data[vp.currentKey]; ok {
			clipboardContent.WriteString(fmt.Sprintf("%v", value))
		}
	} else {
		// Copy all secret data
		for key, value := range vp.currentSecret.Data {
			clipboardContent.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(clipboardContent.String()); err != nil {
		vp.logger.Errorf("Failed to copy to clipboard: %v", err)
		return
	}

	vp.logger.Info("Copied to clipboard")
	vp.showCopySuccess()
}

// copySecretValue copies just the secret value to clipboard
func (vp *ValuePanel) copySecretValue() {
	if vp.currentSecret == nil {
		vp.logger.Warn("No secret selected to copy")
		return
	}

	var valueStr string

	if vp.currentKey != "" {
		// Copy specific key value
		if value, ok := vp.currentSecret.Data[vp.currentKey]; ok {
			valueStr = fmt.Sprintf("%v", value)
		}
	} else {
		// If there's only one key-value pair, copy just the value
		if len(vp.currentSecret.Data) == 1 {
			for _, value := range vp.currentSecret.Data {
				valueStr = fmt.Sprintf("%v", value)
			}
		} else {
			// Multiple values - copy all values
			var values []string
			for _, value := range vp.currentSecret.Data {
				values = append(values, fmt.Sprintf("%v", value))
			}
			valueStr = strings.Join(values, "\n")
		}
	}

	if valueStr == "" {
		vp.logger.Warn("No value to copy")
		return
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(valueStr); err != nil {
		vp.logger.Errorf("Failed to copy value to clipboard: %v", err)
		return
	}

	vp.logger.Info("Copied value to clipboard")
	vp.showCopySuccess()
}

// showCopySuccess shows a brief success message
func (vp *ValuePanel) showCopySuccess() {
	// Store the current text
	currentText := vp.textView.GetText(false)

	// Show success message
	vp.textView.SetText("[green]✓ Copied to clipboard![white]\n\n" + currentText)

	// Schedule to restore the original text after 1.5 seconds
	go func() {
		time.Sleep(1500 * time.Millisecond)
		vp.textView.SetText(currentText)
	}()
}

// Refresh refreshes the value panel
func (vp *ValuePanel) Refresh() {
	if vp.currentSecret != nil {
		if vp.currentKey != "" {
			vp.displayKeyValue(vp.currentSecret, vp.currentKey)
		} else {
			vp.displaySecretData(vp.currentSecret)
		}
	} else {
		vp.showDefaultMessage()
	}
}

// GetPrimitive returns the underlying tview primitive
func (vp *ValuePanel) GetPrimitive() tview.Primitive {
	return vp.textView
}
