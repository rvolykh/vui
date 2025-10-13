package panels

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// MetadataPanel represents the secret metadata panel
type MetadataPanel struct {
	config        *config.Config
	vaultMgr      *vault.Manager
	textView      *tview.TextView
	currentSecret *vault.SecretNode
	logger        *logrus.Logger
}

// NewMetadataPanel creates a new metadata panel
func NewMetadataPanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *MetadataPanel {
	return &MetadataPanel{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// Initialize initializes the metadata panel
func (mp *MetadataPanel) Initialize() error {
	mp.textView = tview.NewTextView()

	// Set up the text view appearance
	mp.textView.SetBorder(true).
		SetTitle("Metadata").
		SetTitleAlign(tview.AlignLeft)

	// Enable dynamic colors and word wrap
	mp.textView.SetDynamicColors(true)
	mp.textView.SetWordWrap(true)

	// Show initial message
	mp.showDefaultMessage()

	return nil
}

// ShowSecret displays secret metadata
func (mp *MetadataPanel) ShowSecret(secret *vault.SecretNode) {
	mp.currentSecret = secret
	mp.displayMetadata(secret)
}

// ShowDirectory displays directory information
func (mp *MetadataPanel) ShowDirectory(path string, itemCount int) {
	mp.currentSecret = nil
	content := fmt.Sprintf(`[yellow]Type:[white] Directory
[yellow]Path:[white] %s
[yellow]Items:[white] %d

[gray]Select an item to view its metadata[white]`, path, itemCount)
	mp.textView.SetText(content)
}

// displayMetadata displays secret metadata
func (mp *MetadataPanel) displayMetadata(secret *vault.SecretNode) {
	content := fmt.Sprintf(`[yellow]Type:[white] Secret
[yellow]Name:[white] %s
[yellow]Path:[white] %s`, secret.Name, secret.Path)

	if secret.Metadata != nil {
		content += fmt.Sprintf(`
[yellow]Version:[white] %d
[yellow]Created:[white] %s`, secret.Metadata.Version, secret.Metadata.CreatedTime.Format("2006-01-02 15:04:05"))

		if secret.Metadata.Destroyed {
			content += `
[red]Status:[white] Destroyed`
		}

		if !secret.Metadata.DeletionTime.IsZero() {
			content += fmt.Sprintf(`
[yellow]Deletion Time:[white] %s`, secret.Metadata.DeletionTime.Format("2006-01-02 15:04:05"))
		}
	}

	if secret.Data != nil {
		content += fmt.Sprintf(`
[yellow]Keys:[white] %d`, len(secret.Data))
	}

	mp.textView.SetText(content)
}

// ShowKey displays metadata for a specific key within a secret
func (mp *MetadataPanel) ShowKey(secret *vault.SecretNode, key string) {
	mp.currentSecret = secret
	content := fmt.Sprintf(`[yellow]Type:[white] Secret Key
[yellow]Secret:[white] %s
[yellow]Key:[white] %s
[yellow]Path:[white] %s
`, secret.Name, key, secret.Path)

	mp.textView.SetText(content)
}

// showDefaultMessage shows the default message
func (mp *MetadataPanel) showDefaultMessage() {
	content := `[yellow]Item Metadata[white]

Select a secret or directory from the tree to view its metadata.

[gray]Metadata includes:[white]
• Item type (secret/directory)
• Path information
• Version information
• Creation and update times
• Number of keys (for secrets)`

	mp.textView.SetText(content)
}

// Refresh refreshes the metadata panel
func (mp *MetadataPanel) Refresh() {
	if mp.currentSecret != nil {
		mp.displayMetadata(mp.currentSecret)
	} else {
		mp.showDefaultMessage()
	}
}

// GetPrimitive returns the underlying tview primitive
func (mp *MetadataPanel) GetPrimitive() tview.Primitive {
	return mp.textView
}
