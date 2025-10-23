package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/ui/forms"
	"github.com/rvolykh/vui/internal/ui/handlers"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// SecretsTree represents the secrets tree panel
type SecretsTree struct {
	config           *config.Config
	vaultMgr         *vault.Manager
	tree             *tview.TreeView
	rootNode         *tview.TreeNode
	formsMgr         *forms.FormsManager
	clipboardHandler *handlers.ClipboardHandler
	selectionHandler func(*vault.SecretNode, string)
	refreshHandler   func()
	modalHandler     func(tview.Primitive, bool) // Handler to show/hide modals
	valuePanel       *SecretsValue               // Reference to value panel for mask toggle
	dialogSvc        dialogService               // Dialog service interface for notifications
	app              *tview.Application
	logger           *logrus.Logger
}

// NewSecretsTree creates a new tree panel
func NewSecretsTree(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger, app *tview.Application) *SecretsTree {
	return &SecretsTree{
		config:           config,
		vaultMgr:         vaultMgr,
		formsMgr:         forms.NewFormsManager(config, vaultMgr, logger, app),
		clipboardHandler: handlers.NewClipboardHandler(vaultMgr),
		app:              app,
		logger:           logger,
	}
}

// Initialize initializes the tree panel
func (tp *SecretsTree) Initialize() error {
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

	// Create an empty tree - tree will be automatically populated after profile selection
	return tp.createEmptyTree("Loading secrets...")
}

// setupKeyboardNavigation sets up keyboard navigation for the tree
func (tp *SecretsTree) setupKeyboardNavigation() {
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
			// Collapse directory on Left arrow (but not the root node)
			node := tp.tree.GetCurrentNode()
			if node != nil && node.IsExpanded() {
				// Don't collapse the root node
				if node == tp.rootNode {
					tp.logger.Debug("Cannot collapse root node")
					return nil
				}
				node.SetExpanded(false)
				tp.logger.Info("Collapsed directory")
			}
			return nil
		case tcell.KeyCtrlD:
			// Delete selected secret
			tp.deleteSecret()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'c':
				// Create new secret
				tp.createSecret()
				return nil
			case 'e':
				// Edit selected secret
				tp.editSecret()
				return nil
			case 'd':
				// Toggle value masking
				tp.toggleValueMasking()
				return nil
			case 'v':
				// Copy value only
				tp.copySecretValue()
				return nil
			}
		}

		return event
	})
}

// loadTree loads the secrets tree
func (tp *SecretsTree) loadTree() error {
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

	// Ensure root node is always expanded
	rootNode.SetExpanded(true)

	tp.logger.Infof("Successfully loaded secrets tree with %d children", len(rootNode.GetChildren()))

	return nil
}

// createEmptyTree creates an empty tree with a message
func (tp *SecretsTree) createEmptyTree(message string) error {
	rootNode := tview.NewTreeNode("secrets").
		SetSelectable(true)

	// Add a message node
	messageNode := tview.NewTreeNode("⚠️ " + message).
		SetSelectable(false).
		SetColor(tcell.ColorYellow)

	rootNode.AddChild(messageNode)

	tp.rootNode = rootNode
	tp.tree.SetRoot(rootNode)

	// Ensure root node is always expanded
	rootNode.SetExpanded(true)

	return nil
}

// buildTree builds the tree structure recursively
func (tp *SecretsTree) buildTree(secretsManager *vault.SecretsManager, path string) (*tview.TreeNode, error) {
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
func (tp *SecretsTree) handleNodeChanged(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		return
	}

	// Check if this is a key node (child of a secret)
	if keyRef, ok := reference.(string); ok {
		// This is a key within a secret
		parent := tp.findParentNode(node)
		if parent != nil {
			if parentRef := parent.GetReference(); parentRef != nil {
				if secret, ok := parentRef.(*vault.SecretNode); ok {
					tp.logger.Infof("Key navigated to: %s in secret %s", keyRef, secret.Path)
					if tp.selectionHandler != nil {
						tp.selectionHandler(secret, keyRef)
					}
					return
				}
			}
		}
	}

	secret, ok := reference.(*vault.SecretNode)
	if !ok {
		return
	}

	tp.logger.Infof("Node navigated to: %s (isSecret: %v)", secret.Path, secret.IsSecret)

	// Update the right panel to show this secret/directory
	if tp.selectionHandler != nil {
		tp.selectionHandler(secret, "")
	}
}

// handleNodeSelection handles when a tree node is explicitly selected (Enter/Right arrow)
func (tp *SecretsTree) handleNodeSelection(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		tp.logger.Debug("Node selection: no reference found")
		return
	}

	// Check if this is a key node (child of a secret)
	if _, ok := reference.(string); ok {
		// This is a key within a secret - keys can't be expanded further
		tp.logger.Debug("Node selection: key node selected")
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
	} else {
		// If this is a secret, expand/collapse it to show keys
		tp.expandSecret(node, secret)
	}
}

// expandDirectory expands a directory node
func (tp *SecretsTree) expandDirectory(node *tview.TreeNode, path string) {
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

// expandSecret expands a secret node to show its keys
func (tp *SecretsTree) expandSecret(node *tview.TreeNode, secret *vault.SecretNode) {
	tp.logger.Infof("Expanding secret: %s", secret.Path)

	// Check if already loaded (has children)
	children := node.GetChildren()
	if len(children) > 0 {
		// Already loaded, just toggle expansion
		tp.logger.Infof("Secret already loaded, toggling: %s", secret.Path)
		node.SetExpanded(!node.IsExpanded())
		return
	}

	// First time expanding - load the secret data to get keys
	tp.logger.Infof("Loading keys for secret: %s", secret.Path)

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

	// Get the full secret with data
	fullSecret, err := secretsManager.GetSecret(secret.Path)
	if err != nil {
		tp.logger.Errorf("Failed to get secret '%s': %v", secret.Path, err)
		// Add error node
		errorNode := tview.NewTreeNode(fmt.Sprintf("❌ Error: %v", err)).
			SetSelectable(false).
			SetColor(tcell.ColorRed)
		node.AddChild(errorNode)
		return
	}

	tp.logger.Infof("Found %d keys in secret: %s", len(fullSecret.Data), secret.Path)

	// Add child nodes for each key
	if len(fullSecret.Data) > 0 {
		// Sort keys alphabetically
		keys := make([]string, 0, len(fullSecret.Data))
		for key := range fullSecret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Add child nodes in sorted order
		for _, key := range keys {
			tp.logger.Infof("Adding key: %s", key)

			keyNode := tview.NewTreeNode("🔑 " + key).
				SetSelectable(true).
				SetReference(key). // Store the key name as reference
				SetColor(tcell.ColorLightBlue)

			node.AddChild(keyNode)
		}
		node.SetExpanded(true)
		tp.logger.Infof("Successfully expanded secret '%s' with %d keys", secret.Path, len(fullSecret.Data))
	} else {
		// No keys, add an empty placeholder
		emptyNode := tview.NewTreeNode("(empty secret)").
			SetSelectable(false).
			SetColor(tcell.ColorGray)
		node.AddChild(emptyNode)
		node.SetExpanded(true)
		tp.logger.Infof("Secret '%s' has no keys", secret.Path)
	}
}

// findParentNode finds the parent node of a given node
func (tp *SecretsTree) findParentNode(targetNode *tview.TreeNode) *tview.TreeNode {
	if tp.rootNode == nil {
		return nil
	}
	return tp.findParentNodeRecursive(tp.rootNode, targetNode)
}

// findParentNodeRecursive recursively searches for the parent of a target node
func (tp *SecretsTree) findParentNodeRecursive(current *tview.TreeNode, target *tview.TreeNode) *tview.TreeNode {
	children := current.GetChildren()
	for _, child := range children {
		if child == target {
			return current
		}
		if parent := tp.findParentNodeRecursive(child, target); parent != nil {
			return parent
		}
	}
	return nil
}

// createSecret creates a new secret
func (tp *SecretsTree) createSecret() {
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
func (tp *SecretsTree) editSecret() {
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
func (tp *SecretsTree) deleteSecret() {
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

// Refresh refreshes the tree
func (tp *SecretsTree) Refresh() {
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
func (tp *SecretsTree) SetSelectionHandler(handler func(*vault.SecretNode, string)) {
	tp.selectionHandler = handler
}

// SetRefreshHandler sets the refresh handler
func (tp *SecretsTree) SetRefreshHandler(handler func()) {
	tp.refreshHandler = handler
}

// SetModalHandler sets the modal handler
func (tp *SecretsTree) SetModalHandler(handler func(tview.Primitive, bool)) {
	tp.modalHandler = handler
}

// SetValuePanel sets the value panel reference for actions
func (tp *SecretsTree) SetValuePanel(valuePanel *SecretsValue) {
	tp.valuePanel = valuePanel
}

// SetDialogService sets the dialog service for showing notifications
func (tp *SecretsTree) SetDialogService(dialogSvc dialogService) {
	tp.dialogSvc = dialogSvc
}

// toggleValueMasking toggles the masking of values in the value panel
func (tp *SecretsTree) toggleValueMasking() {
	if tp.valuePanel != nil {
		tp.valuePanel.ToggleMasking()
		tp.logger.Info("Toggled value masking")
	}
}

// copySecretValue copies the secret value to clipboard
func (tp *SecretsTree) copySecretValue() {
	node := tp.tree.GetCurrentNode()
	if node == nil {
		tp.logger.Warn("No node selected")
		return
	}

	reference := node.GetReference()
	if reference == nil {
		tp.logger.Warn("No reference in selected node")
		return
	}

	var copiedMessage string
	defer func() {
		// Display info dialog if something was copied
		if copiedMessage != "" && tp.dialogSvc != nil {
			tp.dialogSvc.ShowInfo("Clipboard", copiedMessage, nil)
		}
	}()

	// Check if this is a key node (child of a secret)
	if keyRef, ok := reference.(string); ok {
		// This is a key within a secret - copy that specific key's value
		parent := tp.findParentNode(node)
		if parent != nil {
			if parentRef := parent.GetReference(); parentRef != nil {
				if secret, ok := parentRef.(*vault.SecretNode); ok {
					if tp.copyKeyValue(secret, keyRef) {
						copiedMessage = fmt.Sprintf("Copied '%s' key '%s' value to clipboard", secret.Path, keyRef)
					}
					return
				}
			}
		}
	}

	// Otherwise, check if it's a secret node
	if secret, ok := reference.(*vault.SecretNode); ok && secret.IsSecret {
		if tp.copySecretValues(secret) {
			copiedMessage = fmt.Sprintf("Copied '%s' value(s) to clipboard", secret.Path)
		}
	}
}

// copyKeyValue copies a specific key's value from a secret
func (tp *SecretsTree) copyKeyValue(secret *vault.SecretNode, key string) bool {
	if err := tp.clipboardHandler.CopyKeyValue(secret, key); err != nil {
		tp.logger.Errorf("Failed to copy key value: %v", err)
		return false
	}
	tp.logger.Infof("Copied key '%s' value to clipboard", key)
	return true
}

// copySecretValues copies all values from a secret (or single value if only one key)
func (tp *SecretsTree) copySecretValues(secret *vault.SecretNode) bool {
	if err := tp.clipboardHandler.CopySecretValues(secret); err != nil {
		tp.logger.Errorf("Failed to copy secret values: %v", err)
		return false
	}
	tp.logger.Info("Copied secret values to clipboard")
	return true
}

// showModal shows or hides a modal
func (tp *SecretsTree) showModal(primitive tview.Primitive, show bool) {
	if tp.modalHandler != nil {
		tp.modalHandler(primitive, show)
	}
}

// GetPrimitive returns the underlying tview primitive
func (tp *SecretsTree) GetPrimitive() tview.Primitive {
	return tp.tree
}
