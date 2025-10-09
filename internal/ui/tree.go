package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// TreePanel represents the secrets tree panel
type TreePanel struct {
	config           *config.Config
	vaultMgr         *vault.Manager
	tree             *tview.TreeView
	rootNode         *tview.TreeNode
	formsMgr         *FormsManager
	selectionHandler func(string, bool)
	refreshHandler   func()
	modalHandler     func(tview.Primitive, bool) // Handler to show/hide modals
	logger           *logrus.Logger
}

// NewTreePanel creates a new tree panel
func NewTreePanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *TreePanel {
	return &TreePanel{
		config:   config,
		vaultMgr: vaultMgr,
		formsMgr: NewFormsManager(config, vaultMgr, logger),
		logger:   logger,
	}
}

// Initialize initializes the tree panel
func (tp *TreePanel) Initialize() error {
	tp.tree = tview.NewTreeView()

	// Set up the tree appearance
	tp.tree.SetBorder(true).
		SetTitle("Secrets").
		SetTitleAlign(tview.AlignLeft)

	// Set up node changed handler (for navigation)
	tp.tree.SetChangedFunc(func(node *tview.TreeNode) {
		tp.handleNodeChanged(node)
	})

	// Set up keyboard navigation
	tp.setupKeyboardNavigation()

	// Load the initial tree
	return tp.loadTree()
}

// setupKeyboardNavigation sets up keyboard navigation for the tree
func (tp *TreePanel) setupKeyboardNavigation() {
	tp.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyRight:
			// Handle selection/expansion on Enter or Right arrow
			node := tp.tree.GetCurrentNode()
			if node != nil {
				tp.handleNodeSelection(node)
			}
			return nil
		case tcell.KeyLeft:
			// Collapse directory on Left arrow
			node := tp.tree.GetCurrentNode()
			if node != nil && node.IsExpanded() {
				node.SetExpanded(false)
				tp.logger.Info("Collapsed directory")
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r':
				// Refresh tree
				tp.Refresh()
				return nil
			case 'c':
				// Create new secret
				tp.createSecret()
				return nil
			case 'e':
				// Edit selected secret
				tp.editSecret()
				return nil
			case 'd':
				// Delete selected secret (with Ctrl)
				if event.Modifiers()&tcell.ModCtrl != 0 {
					tp.deleteSecret()
					return nil
				}
			case 's':
				// Search secrets
				tp.searchSecrets()
				return nil
			}
		}

		return event
	})
}

// loadTree loads the secrets tree
func (tp *TreePanel) loadTree() error {
	tp.logger.Info("Loading secrets tree...")

	secretsManager, err := tp.vaultMgr.GetSecretsManager()
	if err != nil {
		tp.logger.Warnf("Failed to get secrets manager: %v", err)
		// Create an empty tree with a message
		return tp.createEmptyTree("No vault connection available")
	}

	// Build the tree structure
	rootNode, err := tp.buildTree(secretsManager, "")
	if err != nil {
		tp.logger.Warnf("Failed to build tree: %v", err)
		// Create an empty tree with an error message
		return tp.createEmptyTree(fmt.Sprintf("Failed to load secrets: %v", err))
	}

	tp.rootNode = rootNode
	tp.tree.SetRoot(rootNode)
	tp.tree.SetCurrentNode(rootNode)

	tp.logger.Infof("Successfully loaded secrets tree with %d children", len(rootNode.GetChildren()))

	return nil
}

// createEmptyTree creates an empty tree with a message
func (tp *TreePanel) createEmptyTree(message string) error {
	rootNode := tview.NewTreeNode("secrets").
		SetSelectable(true)

	// Add a message node
	messageNode := tview.NewTreeNode("⚠️ " + message).
		SetSelectable(false).
		SetColor(tcell.ColorYellow)

	rootNode.AddChild(messageNode)

	tp.rootNode = rootNode
	tp.tree.SetRoot(rootNode)

	return nil
}

// buildTree builds the tree structure recursively
func (tp *TreePanel) buildTree(secretsManager *vault.SecretsManager, path string) (*tview.TreeNode, error) {
	tp.logger.Infof("Building tree for path: '%s'", path)

	// Get secrets at this path
	secrets, err := secretsManager.ListSecrets(path)
	if err != nil {
		tp.logger.Errorf("Failed to list secrets at path '%s': %v", path, err)
		return nil, fmt.Errorf("failed to list secrets at path '%s': %w", path, err)
	}

	tp.logger.Infof("Found %d items at path '%s'", len(secrets), path)

	// Create root node
	var nodeText string
	if path == "" {
		nodeText = "secrets"
	} else {
		nodeText = strings.TrimPrefix(path, "secrets/")
		if nodeText == "" {
			nodeText = "secrets"
		}
	}

	rootNode := tview.NewTreeNode(nodeText).
		SetSelectable(true)

	// Add child nodes
	for _, secret := range secrets {
		tp.logger.Infof("Adding node: %s (path: %s, isSecret: %v)", secret.Name, secret.Path, secret.IsSecret)

		childNode := tview.NewTreeNode(secret.Name).
			SetSelectable(true)

		// Store the secret info in the node reference
		childNode.SetReference(secret)

		if secret.IsSecret {
			// This is a secret
			childNode.SetText("🔐 " + secret.Name)
			tp.logger.Infof("  -> Added as secret")
		} else {
			// This is a directory
			childNode.SetText("📁 " + secret.Name)
			childNode.SetColor(tcell.ColorYellow)
			tp.logger.Infof("  -> Added as directory (collapsed)")
		}

		rootNode.AddChild(childNode)
	}

	tp.logger.Infof("Tree built successfully for path '%s'", path)
	return rootNode, nil
}

// handleNodeChanged handles when the selected node changes (navigation)
func (tp *TreePanel) handleNodeChanged(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		return
	}

	secret, ok := reference.(*vault.SecretNode)
	if !ok {
		return
	}

	tp.logger.Infof("Node navigated to: %s (isSecret: %v)", secret.Path, secret.IsSecret)

	// Update the right panel to show this secret/directory
	if tp.selectionHandler != nil {
		tp.selectionHandler(secret.Path, secret.IsSecret)
	}
}

// handleNodeSelection handles when a tree node is explicitly selected (Enter/Right arrow)
func (tp *TreePanel) handleNodeSelection(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		tp.logger.Debug("Node selection: no reference found")
		return
	}

	secret, ok := reference.(*vault.SecretNode)
	if !ok {
		tp.logger.Debug("Node selection: reference is not a SecretNode")
		return
	}

	tp.logger.Infof("Node explicitly selected: %s (isSecret: %v)", secret.Path, secret.IsSecret)

	// If this is a directory, expand/collapse it
	if !secret.IsSecret {
		tp.expandDirectory(node, secret.Path)
	}
}

// expandDirectory expands a directory node
func (tp *TreePanel) expandDirectory(node *tview.TreeNode, path string) {
	tp.logger.Infof("Expanding directory: %s", path)

	// Check if already loaded (has children)
	children := node.GetChildren()
	if len(children) > 0 {
		// Already loaded, just toggle expansion
		tp.logger.Infof("Directory already loaded, toggling: %s", path)
		node.SetExpanded(!node.IsExpanded())
		return
	}

	// First time expanding - load the children
	tp.logger.Infof("Loading children for directory: %s", path)

	secretsManager, err := tp.vaultMgr.GetSecretsManager()
	if err != nil {
		tp.logger.Errorf("Failed to get secrets manager: %v", err)
		// Add error node
		errorNode := tview.NewTreeNode("❌ Error loading").
			SetSelectable(false).
			SetColor(tcell.ColorRed)
		node.AddChild(errorNode)
		return
	}

	// Get secrets in this directory
	secrets, err := secretsManager.ListSecrets(path)
	if err != nil {
		tp.logger.Errorf("Failed to list secrets in directory '%s': %v", path, err)
		// Add error node
		errorNode := tview.NewTreeNode(fmt.Sprintf("❌ Error: %v", err)).
			SetSelectable(false).
			SetColor(tcell.ColorRed)
		node.AddChild(errorNode)
		return
	}

	tp.logger.Infof("Found %d items in directory: %s", len(secrets), path)

	// Add child nodes
	for _, secret := range secrets {
		tp.logger.Infof("Adding child: %s (isSecret: %v)", secret.Name, secret.IsSecret)

		childNode := tview.NewTreeNode(secret.Name).
			SetSelectable(true).
			SetReference(secret)

		if secret.IsSecret {
			childNode.SetText("🔐 " + secret.Name)
		} else {
			childNode.SetText("📁 " + secret.Name)
			childNode.SetColor(tcell.ColorYellow)
		}

		node.AddChild(childNode)
	}

	// If we found children, expand the node
	if len(secrets) > 0 {
		node.SetExpanded(true)
		tp.logger.Infof("Successfully expanded directory '%s' with %d children", path, len(secrets))
	} else {
		// No children, add an empty placeholder
		emptyNode := tview.NewTreeNode("(empty)").
			SetSelectable(false).
			SetColor(tcell.ColorGray)
		node.AddChild(emptyNode)
		node.SetExpanded(true)
		tp.logger.Infof("Directory '%s' is empty", path)
	}
}

// createSecret creates a new secret
func (tp *TreePanel) createSecret() {
	// Get current selection
	node := tp.tree.GetCurrentNode()
	if node == nil {
		return
	}

	// Determine the path for the new secret
	var basePath string
	reference := node.GetReference()
	if reference != nil {
		if secret, ok := reference.(*vault.SecretNode); ok {
			if secret.IsSecret {
				// If a secret is selected, use its parent directory
				basePath = strings.TrimSuffix(secret.Path, "/"+secret.Name)
			} else {
				// If a directory is selected, use it
				basePath = secret.Path
			}
		}
	}

	// Show create secret form as modal
	form := tp.formsMgr.CreateSecretForm(basePath, func() {
		// Refresh the tree after creating secret
		tp.Refresh()
		// Return to main UI
		tp.showModal(nil, false)
	})

	tp.showModal(form, true)
	tp.logger.Infof("Create secret in path: %s", basePath)
}

// editSecret edits the selected secret
func (tp *TreePanel) editSecret() {
	node := tp.tree.GetCurrentNode()
	if node == nil {
		return
	}

	reference := node.GetReference()
	if reference == nil {
		return
	}

	secret, ok := reference.(*vault.SecretNode)
	if !ok || !secret.IsSecret {
		return
	}

	// Show edit secret form as modal
	form := tp.formsMgr.EditSecretForm(secret.Path, func() {
		// Refresh the tree after editing secret
		tp.Refresh()
		// Return to main UI
		tp.showModal(nil, false)
	})

	tp.showModal(form, true)
	tp.logger.Infof("Edit secret: %s", secret.Path)
}

// deleteSecret deletes the selected secret
func (tp *TreePanel) deleteSecret() {
	node := tp.tree.GetCurrentNode()
	if node == nil {
		return
	}

	reference := node.GetReference()
	if reference == nil {
		return
	}

	secret, ok := reference.(*vault.SecretNode)
	if !ok || !secret.IsSecret {
		return
	}

	// Show confirmation dialog as modal
	form := tp.formsMgr.DeleteSecretForm(secret.Path, func() {
		// Refresh the tree after deleting secret
		tp.Refresh()
		// Return to main UI
		tp.showModal(nil, false)
	})

	tp.showModal(form, true)
	tp.logger.Infof("Delete secret: %s", secret.Path)
}

// searchSecrets searches for secrets
func (tp *TreePanel) searchSecrets() {
	// Show search dialog as modal
	form := tp.formsMgr.SearchForm(func() {
		// Return to main UI
		tp.showModal(nil, false)
	})

	tp.showModal(form, true)
	tp.logger.Info("Search secrets")
}

// Refresh refreshes the tree
func (tp *TreePanel) Refresh() {
	tp.logger.Info("Refreshing tree")

	// Reload the tree
	if err := tp.loadTree(); err != nil {
		tp.logger.Errorf("Failed to refresh tree: %v", err)
		return
	}

	// Call refresh handler
	if tp.refreshHandler != nil {
		tp.refreshHandler()
	}
}

// SetSelectionHandler sets the selection handler
func (tp *TreePanel) SetSelectionHandler(handler func(string, bool)) {
	tp.selectionHandler = handler
}

// SetRefreshHandler sets the refresh handler
func (tp *TreePanel) SetRefreshHandler(handler func()) {
	tp.refreshHandler = handler
}

// SetModalHandler sets the modal handler
func (tp *TreePanel) SetModalHandler(handler func(tview.Primitive, bool)) {
	tp.modalHandler = handler
}

// showModal shows or hides a modal
func (tp *TreePanel) showModal(primitive tview.Primitive, show bool) {
	if tp.modalHandler != nil {
		tp.modalHandler(primitive, show)
	}
}

// GetPrimitive returns the underlying tview primitive
func (tp *TreePanel) GetPrimitive() tview.Primitive {
	return tp.tree
}
