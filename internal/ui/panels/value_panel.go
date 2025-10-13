package panels

import (
	"fmt"
	"sort"
	"strings"

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
	isMasked      bool // Track if values are currently masked
	logger        *logrus.Logger
}

// NewValuePanel creates a new value panel
func NewValuePanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *ValuePanel {
	return &ValuePanel{
		config:   config,
		vaultMgr: vaultMgr,
		isMasked: true, // Values are masked by default
		logger:   logger,
	}
}

// Initialize initializes the value panel
func (vp *ValuePanel) Initialize() error {
	vp.textView = tview.NewTextView()

	// Set up the text view appearance
	vp.textView.SetBorder(true).
		SetTitle("Value").
		SetTitleAlign(tview.AlignLeft)

	// Enable dynamic colors and word wrap
	vp.textView.SetDynamicColors(true)
	vp.textView.SetWordWrap(true)
	vp.textView.SetScrollable(true)

	// Show initial message
	vp.showDefaultMessage()

	return nil
}

// ShowSecret displays all key-value pairs of a secret
func (vp *ValuePanel) ShowSecret(secret *vault.SecretNode) {
	vp.currentSecret = secret
	vp.currentKey = ""
	vp.isMasked = true // Reset to masked when showing a new secret
	vp.displaySecretData(secret)
}

// ShowKey displays a specific key's value from a secret
func (vp *ValuePanel) ShowKey(secret *vault.SecretNode, key string) {
	vp.currentSecret = secret
	vp.currentKey = key
	vp.isMasked = true // Reset to masked when showing a new key
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

	if len(secret.Data) > 0 {
		// Sort keys alphabetically
		keys := make([]string, 0, len(secret.Data))
		for key := range secret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Display key-value pairs in sorted order
		for _, key := range keys {
			value := secret.Data[key]
			content.WriteString(fmt.Sprintf("[green]%s:[white]\n", key))

			// Format the value (masked or unmasked)
			valueStr := vp.formatValue(value)
			if vp.isMasked {
				content.WriteString("••••••••")
			} else {
				content.WriteString(valueStr)
			}
			content.WriteString("\n\n")
		}
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
			if vp.isMasked {
				content.WriteString("••••••••")
			} else {
				content.WriteString(valueStr)
			}
			content.WriteString("\n\n")
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

[yellow]Navigation:[white]
• Select a secret to view all its data
• Expand a secret to see individual keys
• Select a key to view its specific value`

	vp.textView.SetText(content)
}

// ToggleMasking toggles the masking of secret values (public for TreePanel)
func (vp *ValuePanel) ToggleMasking() {
	vp.isMasked = !vp.isMasked
	vp.logger.Infof("Toggled value masking: masked=%v", vp.isMasked)

	// Refresh the display
	if vp.currentSecret != nil {
		if vp.currentKey != "" {
			vp.displayKeyValue(vp.currentSecret, vp.currentKey)
		} else {
			vp.displaySecretData(vp.currentSecret)
		}
	}
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
