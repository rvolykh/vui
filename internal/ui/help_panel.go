package ui

import (
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

// HelpPanel represents the navigation help panel
type HelpPanel struct {
	config   *config.Config
	textView *tview.TextView
	logger   *logrus.Logger
}

// NewHelpPanel creates a new help panel
func NewHelpPanel(config *config.Config, logger *logrus.Logger) *HelpPanel {
	return &HelpPanel{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the help panel
func (hp *HelpPanel) Initialize() error {
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
func (hp *HelpPanel) updateHelpText() {
	helpText := `[yellow]Navigation:[white] ↑/↓:Move  ←/→:Collapse/Expand  Enter:Select  Tab:Switch Panel  [yellow]|[white]  ` +
		`[yellow]Actions:[white] c:Create  e:Edit  Ctrl+d:Delete  r:Refresh  s:Search  [yellow]|[white]  ` +
		`[yellow]Secret:[white] c:Copy All  v:Copy Value  [yellow]|[white]  ` +
		`[yellow]Global:[white] F1:Help  F5:Refresh  Ctrl+v:Switch Vault  Ctrl+C:Exit`

	hp.textView.SetText(helpText)
}

// GetPrimitive returns the underlying tview primitive
func (hp *HelpPanel) GetPrimitive() tview.Primitive {
	return hp.textView
}
