package panels

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretsTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	assert.NotNil(t, st)
	assert.Equal(t, cfg, st.config)
	assert.Equal(t, vaultMgr, st.vaultMgr)
	assert.Equal(t, logger, st.logger)
	assert.Equal(t, app, st.app)
	assert.NotNil(t, st.formsMgr)
	assert.NotNil(t, st.clipboardHandler)
	assert.Nil(t, st.tree) // Not initialized yet
	assert.Nil(t, st.rootNode)
}

func TestNewSecretsTree_WithNilConfig(t *testing.T) {
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(nil, vaultMgr, logger, app)

	assert.NotNil(t, st)
	assert.Nil(t, st.config)
}

func TestNewSecretsTree_WithNilVaultManager(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, nil, logger, app)

	assert.NotNil(t, st)
	assert.Nil(t, st.vaultMgr)
}

func TestNewSecretsTree_WithNilLogger(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, nil, app)

	assert.NotNil(t, st)
	assert.Nil(t, st.logger)
}

func TestNewSecretsTree_WithNilApp(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()

	st := NewSecretsTree(cfg, vaultMgr, logger, nil)

	assert.NotNil(t, st)
	assert.Nil(t, st.app)
}

func TestSecretsTree_Initialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	err := st.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, st.tree)
	assert.NotNil(t, st.rootNode)
}

func TestSecretsTree_Initialize_CreatesTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	err := st.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, st.tree)

	// Tree should have a root node
	root := st.tree.GetRoot()
	assert.NotNil(t, root)
}

func TestSecretsTree_Initialize_MultipleCallsSafe(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Initialize multiple times should not panic
	err1 := st.Initialize()
	err2 := st.Initialize()
	err3 := st.Initialize()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NotNil(t, st.tree)
}

func TestSecretsTree_GetPrimitive(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	primitive := st.GetPrimitive()

	assert.NotNil(t, primitive)
	assert.Equal(t, st.tree, primitive)
}

func TestSecretsTree_GetPrimitive_BeforeInitialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	primitive := st.GetPrimitive()

	assert.Nil(t, primitive)
}

func TestSecretsTree_SetSelectionHandler(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	handlerCalled := false
	handler := func(secret *vault.SecretNode, key string) {
		handlerCalled = true
	}

	st.SetSelectionHandler(handler)

	assert.NotNil(t, st.selectionHandler)

	// Test the handler
	st.selectionHandler(nil, "")
	assert.True(t, handlerCalled)
}

func TestSecretsTree_SetSelectionHandler_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic with nil handler
	st.SetSelectionHandler(nil)

	assert.Nil(t, st.selectionHandler)
}

func TestSecretsTree_SetRefreshHandler(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	handlerCalled := false
	handler := func() {
		handlerCalled = true
	}

	st.SetRefreshHandler(handler)

	assert.NotNil(t, st.refreshHandler)

	// Test the handler
	st.refreshHandler()
	assert.True(t, handlerCalled)
}

func TestSecretsTree_SetRefreshHandler_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic with nil handler
	st.SetRefreshHandler(nil)

	assert.Nil(t, st.refreshHandler)
}

func TestSecretsTree_SetModalHandler(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	handlerCalled := false
	handler := func(primitive tview.Primitive, show bool) {
		handlerCalled = true
	}

	st.SetModalHandler(handler)

	assert.NotNil(t, st.modalHandler)

	// Test the handler
	st.modalHandler(nil, false)
	assert.True(t, handlerCalled)
}

func TestSecretsTree_SetModalHandler_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic with nil handler
	st.SetModalHandler(nil)

	assert.Nil(t, st.modalHandler)
}

func TestSecretsTree_SetValuePanel(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	valuePanel := &SecretsValue{}
	st.SetValuePanel(valuePanel)

	assert.NotNil(t, st.valuePanel)
	assert.Equal(t, valuePanel, st.valuePanel)
}

func TestSecretsTree_SetValuePanel_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic with nil value panel
	st.SetValuePanel(nil)

	assert.Nil(t, st.valuePanel)
}

func TestSecretsTree_SetDialogService(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Mock dialog service
	mockDialogSvc := &mockDialogService{}
	st.SetDialogService(mockDialogSvc)

	assert.NotNil(t, st.dialogSvc)
}

func TestSecretsTree_SetDialogService_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic with nil dialog service
	st.SetDialogService(nil)

	assert.Nil(t, st.dialogSvc)
}

func TestSecretsTree_ShowModal(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	handlerCalled := false
	var receivedPrimitive tview.Primitive
	var receivedShow bool

	st.SetModalHandler(func(primitive tview.Primitive, show bool) {
		handlerCalled = true
		receivedPrimitive = primitive
		receivedShow = show
	})

	modal := tview.NewModal()
	st.showModal(modal, true)

	assert.True(t, handlerCalled)
	assert.Equal(t, modal, receivedPrimitive)
	assert.True(t, receivedShow)
}

func TestSecretsTree_ShowModal_WithoutHandler(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic without modal handler
	st.showModal(nil, false)
}

func TestSecretsTree_CreateEmptyTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.tree = tview.NewTreeView()

	err := st.createEmptyTree("Test message")

	assert.NoError(t, err)
	assert.NotNil(t, st.rootNode)
	assert.True(t, st.rootNode.IsExpanded())
}

func TestSecretsTree_CreateEmptyTree_WithDifferentMessages(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.tree = tview.NewTreeView()

	messages := []string{
		"Loading...",
		"No connection",
		"Error occurred",
		"",
	}

	for _, msg := range messages {
		err := st.createEmptyTree(msg)
		assert.NoError(t, err)
		assert.NotNil(t, st.rootNode)
	}
}

func TestSecretsTree_FindParentNode_NilRootNode(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.rootNode = nil

	node := tview.NewTreeNode("test")
	parent := st.findParentNode(node)

	assert.Nil(t, parent)
}

func TestSecretsTree_FindParentNode_SimpleTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Create a simple tree structure
	root := tview.NewTreeNode("root")
	child := tview.NewTreeNode("child")
	root.AddChild(child)

	st.rootNode = root

	parent := st.findParentNode(child)

	assert.NotNil(t, parent)
	assert.Equal(t, root, parent)
}

func TestSecretsTree_FindParentNode_NotFound(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Create a tree
	root := tview.NewTreeNode("root")
	st.rootNode = root

	// Create a node not in the tree
	orphan := tview.NewTreeNode("orphan")

	parent := st.findParentNode(orphan)

	assert.Nil(t, parent)
}

func TestSecretsTree_FindParentNode_NestedTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Create a nested tree structure
	root := tview.NewTreeNode("root")
	child1 := tview.NewTreeNode("child1")
	child2 := tview.NewTreeNode("child2")
	grandchild := tview.NewTreeNode("grandchild")

	root.AddChild(child1)
	child1.AddChild(child2)
	child2.AddChild(grandchild)

	st.rootNode = root

	// Find parent of grandchild
	parent := st.findParentNode(grandchild)

	assert.NotNil(t, parent)
	assert.Equal(t, child2, parent)
}

func TestSecretsTree_Initialize_SetsUpTreeView(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	// TreeView should be configured
	assert.NotNil(t, st.tree)

	// Root should be set
	root := st.tree.GetRoot()
	assert.NotNil(t, root)
}

func TestSecretsTree_ToggleValueMasking_WithValuePanel(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Create a value panel
	mockValuePanel := NewSecretsValue(cfg, vaultMgr, logger)
	mockValuePanel.Initialize()
	st.SetValuePanel(mockValuePanel)

	// Should not panic
	st.toggleValueMasking()
}

func TestSecretsTree_ToggleValueMasking_WithoutValuePanel(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Should not panic without value panel
	st.toggleValueMasking()
}

func TestSecretsTree_CopySecretValue_NoNode(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	// Should not panic with no node selected
	st.copySecretValue()
}

func TestSecretsTree_CreateSecret_NoNode(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	// Should not panic with no node selected
	st.createSecret()
}

func TestSecretsTree_EditSecret_NoNode(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	// Should not panic with no node selected
	st.editSecret()
}

func TestSecretsTree_DeleteSecret_NoNode(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	st.Initialize()

	// Should not panic with no node selected
	st.deleteSecret()
}

func TestSecretsTree_HandleNodeChanged_NilReference(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	node := tview.NewTreeNode("test")
	// No reference set

	// Should not panic
	st.handleNodeChanged(node)
}

func TestSecretsTree_HandleNodeSelection_NilReference(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	node := tview.NewTreeNode("test")
	// No reference set

	// Should not panic
	st.handleNodeSelection(node)
}

func TestSecretsTree_AllSettersCalled(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)

	// Call all setters
	st.SetSelectionHandler(func(*vault.SecretNode, string) {})
	st.SetRefreshHandler(func() {})
	st.SetModalHandler(func(tview.Primitive, bool) {})
	st.SetValuePanel(&SecretsValue{})
	st.SetDialogService(&mockDialogService{})

	assert.NotNil(t, st.selectionHandler)
	assert.NotNil(t, st.refreshHandler)
	assert.NotNil(t, st.modalHandler)
	assert.NotNil(t, st.valuePanel)
	assert.NotNil(t, st.dialogSvc)
}

func TestSecretsTree_InitializeCreatesEmptyTree(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	st := NewSecretsTree(cfg, vaultMgr, logger, app)
	err := st.Initialize()

	assert.NoError(t, err)

	// Root should be expanded
	assert.NotNil(t, st.rootNode)
	assert.True(t, st.rootNode.IsExpanded())
}
