package keyboard

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewShortcutRegistry(t *testing.T) {
	registry := NewShortcutRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, ContextMain, registry.currentContext)
	assert.Empty(t, registry.shortcuts)
}

func TestShortcutRegistry_SetContext(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.SetContext(ContextForm)
	assert.Equal(t, ContextForm, registry.GetContext())

	registry.SetContext(ContextProfiles)
	assert.Equal(t, ContextProfiles, registry.GetContext())
}

func TestShortcutRegistry_RegisterKey(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterKey(tcell.KeyF1, ContextGlobal, "Help", func() bool {
		return true
	})

	assert.Len(t, registry.shortcuts, 1)
	assert.Equal(t, tcell.KeyF1, registry.shortcuts[0].Key)
	assert.Equal(t, ContextGlobal, registry.shortcuts[0].Context)
	assert.Equal(t, "Help", registry.shortcuts[0].Description)
}

func TestShortcutRegistry_RegisterRune(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterRune('h', ContextMain, "Help", func() bool {
		return true
	})

	assert.Len(t, registry.shortcuts, 1)
	assert.Equal(t, 'h', registry.shortcuts[0].Rune)
	assert.Equal(t, ContextMain, registry.shortcuts[0].Context)
}

func TestShortcutRegistry_RegisterKeyWithModifiers(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterKeyWithModifiers(tcell.KeyCtrlD, tcell.ModCtrl, ContextMain, "Delete", func() bool {
		return true
	})

	assert.Len(t, registry.shortcuts, 1)
	assert.Equal(t, tcell.KeyCtrlD, registry.shortcuts[0].Key)
	assert.Equal(t, tcell.ModCtrl, registry.shortcuts[0].Modifiers)
}

func TestShortcutRegistry_Handle_SpecialKey(t *testing.T) {
	registry := NewShortcutRegistry()
	called := false

	registry.RegisterKey(tcell.KeyF1, ContextGlobal, "Help", func() bool {
		called = true
		return true
	})

	event := tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone)
	handled := registry.Handle(event)

	assert.True(t, handled)
	assert.True(t, called)
}

func TestShortcutRegistry_Handle_Rune(t *testing.T) {
	registry := NewShortcutRegistry()
	called := false

	registry.RegisterRune('h', ContextMain, "Help", func() bool {
		called = true
		return true
	})

	event := tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)
	handled := registry.Handle(event)

	assert.True(t, handled)
	assert.True(t, called)
}

func TestShortcutRegistry_Handle_ContextMismatch(t *testing.T) {
	registry := NewShortcutRegistry()
	registry.SetContext(ContextProfiles)
	called := false

	// Register shortcut for main context
	registry.RegisterRune('h', ContextMain, "Help", func() bool {
		called = true
		return true
	})

	// Try to trigger it in profiles context
	event := tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)
	handled := registry.Handle(event)

	assert.False(t, handled)
	assert.False(t, called)
}

func TestShortcutRegistry_Handle_GlobalShortcut(t *testing.T) {
	registry := NewShortcutRegistry()
	registry.SetContext(ContextForm) // Different context
	called := false

	// Global shortcuts work in any context
	registry.RegisterKey(tcell.KeyCtrlC, ContextGlobal, "Quit", func() bool {
		called = true
		return true
	})

	event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)
	handled := registry.Handle(event)

	assert.True(t, handled)
	assert.True(t, called)
}

func TestShortcutRegistry_Handle_WithModal(t *testing.T) {
	registry := NewShortcutRegistry()
	registry.SetContext(ContextMain)
	registry.SetModalChecker(func() bool { return true }) // Modal is active

	mainCalled := false
	globalCalled := false

	// Register main context shortcut
	registry.RegisterRune('h', ContextMain, "Help", func() bool {
		mainCalled = true
		return true
	})

	// Register global shortcut
	registry.RegisterKey(tcell.KeyCtrlC, ContextGlobal, "Quit", func() bool {
		globalCalled = true
		return true
	})

	// Main context shortcut should not work when modal is active
	event := tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)
	handled := registry.Handle(event)
	assert.False(t, handled)
	assert.False(t, mainCalled)

	// Global shortcut should still work
	event = tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)
	handled = registry.Handle(event)
	assert.True(t, handled)
	assert.True(t, globalCalled)
}

func TestShortcutRegistry_GetShortcuts(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterKey(tcell.KeyF1, ContextGlobal, "Help", func() bool { return true })
	registry.RegisterRune('h', ContextMain, "Help Alt", func() bool { return true })
	registry.RegisterRune('r', ContextMain, "Refresh", func() bool { return true })
	registry.RegisterRune('n', ContextProfiles, "New", func() bool { return true })

	// Get main shortcuts (should include global)
	mainShortcuts := registry.GetShortcuts(ContextMain)
	assert.Len(t, mainShortcuts, 3) // F1 (global), h, r

	// Get profiles shortcuts (should include global)
	profilesShortcuts := registry.GetShortcuts(ContextProfiles)
	assert.Len(t, profilesShortcuts, 2) // F1 (global), n
}

func TestShortcutRegistry_Clear(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterKey(tcell.KeyF1, ContextGlobal, "Help", func() bool { return true })
	registry.RegisterRune('h', ContextMain, "Help", func() bool { return true })

	assert.Len(t, registry.shortcuts, 2)

	registry.Clear()
	assert.Empty(t, registry.shortcuts)
}

func TestShortcutRegistry_Handle_HandlerReturnsFalse(t *testing.T) {
	registry := NewShortcutRegistry()

	registry.RegisterRune('h', ContextMain, "Help", func() bool {
		return false // Handler indicates event was not fully handled
	})

	event := tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)
	handled := registry.Handle(event)

	assert.False(t, handled) // Event should propagate
}
