package ui

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/ui/common"
	"github.com/rvolykh/vui/internal/ui/panels"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// Layout represents the main application layout
type Layout struct {
	config        *config.Config
	vaultMgr      *vault.Manager
	root          *tview.Flex
	helpPanel     *panels.SecretsTitle
	treePanel     *panels.SecretsTree
	metadataPanel *panels.SecretsMetadata
	valuePanel    *panels.SecretsValue
	statusBar     *panels.SecretsStatus
	dialogSvc     *common.DialogService
	app           *tview.Application
	logger        *logrus.Logger
}

// NewLayout creates a new layout
func NewLayout(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *Layout {
	return &Layout{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// SetApplication sets the tview application reference
func (l *Layout) SetApplication(app *tview.Application) {
	l.app = app
}

// Initialize initializes the layout components
func (l *Layout) Initialize() error {
	l.logger.Info("Initializing UI layout")

	// Create the help panel
	l.helpPanel = panels.NewSecretsTitle(l.config, l.vaultMgr, l.logger)
	if err := l.helpPanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize help panel: %w", err)
	}

	// Create the tree panel
	l.treePanel = panels.NewSecretsTree(l.config, l.vaultMgr, l.logger, l.app)
	if err := l.treePanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize tree panel: %w", err)
	}

	// Create the metadata panel
	l.metadataPanel = panels.NewSecretsMetadata(l.config, l.vaultMgr, l.logger)
	if err := l.metadataPanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize metadata panel: %w", err)
	}

	// Create the value panel
	l.valuePanel = panels.NewSecretsValue(l.config, l.vaultMgr, l.logger)
	if err := l.valuePanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize value panel: %w", err)
	}

	// Create the status bar
	l.statusBar = panels.NewSecretsStatus(l.config, l.vaultMgr, l.logger)
	if err := l.statusBar.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize status bar: %w", err)
	}

	// Set up the layout
	l.setupLayout()

	// Create dialog service
	l.dialogSvc = common.NewDialogService(l.app, l.root)

	// Set up event handlers
	l.setupEventHandlers()

	// Set value panel reference in tree panel for actions
	l.treePanel.SetValuePanel(l.valuePanel)

	return nil
}

// setupLayout creates the main layout structure
func (l *Layout) setupLayout() {
	// Create the right panel (vertical split: metadata on top, value on bottom)
	rightPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(l.metadataPanel.GetPrimitive(), 0, 1, false).
		AddItem(l.valuePanel.GetPrimitive(), 0, 2, false)

	// Create the main horizontal layout (tree on left, right panel on right)
	contentLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(l.treePanel.GetPrimitive(), 0, 1, true).
		AddItem(rightPanel, 0, 2, false)

	// Create the main vertical layout (help at top, content in middle, status bar at bottom)
	l.root = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(l.helpPanel.GetPrimitive(), 9, 0, false).
		AddItem(contentLayout, 0, 1, true).
		AddItem(l.statusBar.GetPrimitive(), 1, 0, false)
}

// setupEventHandlers sets up event handlers between components
func (l *Layout) setupEventHandlers() {
	// When a tree item is selected, update the metadata and value panels
	l.treePanel.SetSelectionHandler(func(node *vault.SecretNode, selectedKey string) {
		l.logger.Infof("Selection changed: path=%s, isSecret=%v, key=%s", node.Path, node.IsSecret, selectedKey)

		if selectedKey != "" {
			// A specific key within a secret is selected
			// We need to fetch the full secret data to show the key's value
			secretsManager, err := l.vaultMgr.GetSecretsManager()
			if err != nil {
				l.logger.Errorf("Failed to get secrets manager: %v", err)
				return
			}

			// Get the full secret data
			secret, err := secretsManager.GetSecret(node.Path)
			if err != nil {
				l.logger.Errorf("Failed to get secret: %v", err)
				return
			}

			l.metadataPanel.ShowKey(secret, selectedKey)
			l.valuePanel.ShowKey(secret, selectedKey)
			l.statusBar.UpdateSelection(fmt.Sprintf("%s/%s", node.Path, selectedKey), true)
		} else if node.IsSecret {
			// A secret is selected (but not a specific key)
			secretsManager, err := l.vaultMgr.GetSecretsManager()
			if err != nil {
				l.logger.Errorf("Failed to get secrets manager: %v", err)
				return
			}

			// Get the full secret data
			secret, err := secretsManager.GetSecret(node.Path)
			if err != nil {
				l.logger.Errorf("Failed to get secret: %v", err)
				return
			}

			l.metadataPanel.ShowSecret(secret)
			l.valuePanel.ShowSecret(secret)
			l.statusBar.UpdateSelection(node.Path, true)
		} else {
			// A directory is selected
			secretsManager, err := l.vaultMgr.GetSecretsManager()
			if err != nil {
				l.logger.Errorf("Failed to get secrets manager: %v", err)
				return
			}

			// Get directory contents to count items
			secrets, err := secretsManager.ListSecrets(node.Path)
			if err != nil {
				l.logger.Errorf("Failed to list secrets: %v", err)
				secrets = []*vault.SecretNode{}
			}

			l.metadataPanel.ShowDirectory(node.Path, len(secrets))
			l.valuePanel.ShowDirectory(node.Path)
			l.statusBar.UpdateSelection(node.Path, false)
		}
	})

	// When tree is refreshed, update status bar
	l.treePanel.SetRefreshHandler(func() {
		l.statusBar.UpdateConnectionStatus()
	})

	// Set up modal handler for tree panel
	l.treePanel.SetModalHandler(func(primitive tview.Primitive, show bool) {
		l.showModal(primitive, show)
	})
}

// GetRoot returns the root primitive
func (l *Layout) GetRoot() tview.Primitive {
	return l.root
}

// Refresh refreshes all layout components
func (l *Layout) Refresh() {
	l.logger.Info("Refreshing layout")

	// Refresh tree panel
	if l.treePanel != nil {
		l.treePanel.Refresh()
	}

	// Refresh metadata panel
	if l.metadataPanel != nil {
		l.metadataPanel.Refresh()
	}

	// Refresh value panel
	if l.valuePanel != nil {
		l.valuePanel.Refresh()
	}

	// Refresh status bar
	if l.statusBar != nil {
		l.statusBar.UpdateConnectionStatus()
	}
}

// GetTreePanel returns the tree panel
func (l *Layout) GetTreePanel() *panels.SecretsTree {
	return l.treePanel
}

// GetMetadataPanel returns the metadata panel
func (l *Layout) GetMetadataPanel() *panels.SecretsMetadata {
	return l.metadataPanel
}

// GetValuePanel returns the value panel
func (l *Layout) GetValuePanel() *panels.SecretsValue {
	return l.valuePanel
}

// GetStatusBar returns the status bar
func (l *Layout) GetStatusBar() *panels.SecretsStatus {
	return l.statusBar
}

// showModal shows or hides a modal dialog
func (l *Layout) showModal(primitive tview.Primitive, show bool) {
	if l.dialogSvc == nil {
		l.logger.Warn("Cannot show modal: dialog service not initialized")
		return
	}

	if show {
		l.dialogSvc.Show(primitive)
	} else {
		l.dialogSvc.Hide()
	}
}

// HasActiveModal returns true if a modal is currently displayed
func (l *Layout) HasActiveModal() bool {
	if l.dialogSvc == nil {
		return false
	}
	return l.dialogSvc.IsModalActive()
}
