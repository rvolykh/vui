package keyboard

import (
	"github.com/gdamore/tcell/v2"
)

// ShortcutHandler is a function that handles a keyboard shortcut
type ShortcutHandler func() bool // returns true if event was handled

// ShortcutContext represents different contexts where shortcuts may be active
type ShortcutContext string

const (
	// ContextGlobal shortcuts are active everywhere
	ContextGlobal ShortcutContext = "global"
	// ContextMain shortcuts are active in the main view (not in forms/modals)
	ContextMain ShortcutContext = "main"
	// ContextForm shortcuts are active in forms
	ContextForm ShortcutContext = "form"
	// ContextProfiles shortcuts are active in the profiles screen
	ContextProfiles ShortcutContext = "profiles"
)

// Shortcut represents a keyboard shortcut
type Shortcut struct {
	Key         tcell.Key
	Rune        rune
	Modifiers   tcell.ModMask
	Context     ShortcutContext
	Description string
	Handler     ShortcutHandler
}

// ShortcutRegistry manages keyboard shortcuts
type ShortcutRegistry struct {
	shortcuts      []Shortcut
	currentContext ShortcutContext
	hasActiveModal func() bool
}

// NewShortcutRegistry creates a new shortcut registry
func NewShortcutRegistry() *ShortcutRegistry {
	return &ShortcutRegistry{
		shortcuts:      []Shortcut{},
		currentContext: ContextMain,
		hasActiveModal: func() bool { return false },
	}
}

// SetModalChecker sets the function to check if a modal is active
func (sr *ShortcutRegistry) SetModalChecker(checker func() bool) {
	sr.hasActiveModal = checker
}

// SetContext sets the current context
func (sr *ShortcutRegistry) SetContext(context ShortcutContext) {
	sr.currentContext = context
}

// GetContext returns the current context
func (sr *ShortcutRegistry) GetContext() ShortcutContext {
	return sr.currentContext
}

// Register registers a new shortcut
func (sr *ShortcutRegistry) Register(shortcut Shortcut) {
	sr.shortcuts = append(sr.shortcuts, shortcut)
}

// RegisterKey registers a shortcut for a special key (like F1, Enter, etc.)
func (sr *ShortcutRegistry) RegisterKey(key tcell.Key, context ShortcutContext, description string, handler ShortcutHandler) {
	sr.Register(Shortcut{
		Key:         key,
		Context:     context,
		Description: description,
		Handler:     handler,
	})
}

// RegisterRune registers a shortcut for a rune (like 'h', 'q', etc.)
func (sr *ShortcutRegistry) RegisterRune(r rune, context ShortcutContext, description string, handler ShortcutHandler) {
	sr.Register(Shortcut{
		Key:         tcell.KeyRune,
		Rune:        r,
		Context:     context,
		Description: description,
		Handler:     handler,
	})
}

// RegisterKeyWithModifiers registers a shortcut with modifiers (like Ctrl+D)
func (sr *ShortcutRegistry) RegisterKeyWithModifiers(key tcell.Key, modifiers tcell.ModMask, context ShortcutContext, description string, handler ShortcutHandler) {
	sr.Register(Shortcut{
		Key:         key,
		Modifiers:   modifiers,
		Context:     context,
		Description: description,
		Handler:     handler,
	})
}

// Handle processes a keyboard event and calls the appropriate handler
func (sr *ShortcutRegistry) Handle(event *tcell.EventKey) bool {
	// Determine effective context - if modal is active, only global shortcuts work
	effectiveContext := sr.currentContext
	if sr.hasActiveModal != nil && sr.hasActiveModal() {
		effectiveContext = ContextForm // Forms include modals
	}

	for _, shortcut := range sr.shortcuts {
		// Check if shortcut matches the event
		if !sr.matchesEvent(shortcut, event) {
			continue
		}

		// Check if shortcut is active in current context
		if !sr.isActiveInContext(shortcut, effectiveContext) {
			continue
		}

		// Call the handler
		if shortcut.Handler != nil {
			if shortcut.Handler() {
				return true
			}
		}
	}

	return false
}

// matchesEvent checks if a shortcut matches a keyboard event
func (sr *ShortcutRegistry) matchesEvent(shortcut Shortcut, event *tcell.EventKey) bool {
	// Check modifiers
	if shortcut.Modifiers != 0 && event.Modifiers() != shortcut.Modifiers {
		return false
	}

	// Check special key
	if shortcut.Key != tcell.KeyRune {
		return event.Key() == shortcut.Key
	}

	// Check rune
	return event.Key() == tcell.KeyRune && event.Rune() == shortcut.Rune
}

// isActiveInContext checks if a shortcut is active in the given context
func (sr *ShortcutRegistry) isActiveInContext(shortcut Shortcut, currentContext ShortcutContext) bool {
	// Global shortcuts are always active
	if shortcut.Context == ContextGlobal {
		return true
	}

	// Otherwise, context must match
	return shortcut.Context == currentContext
}

// GetShortcuts returns all registered shortcuts for a given context
func (sr *ShortcutRegistry) GetShortcuts(context ShortcutContext) []Shortcut {
	var result []Shortcut
	for _, shortcut := range sr.shortcuts {
		if shortcut.Context == context || shortcut.Context == ContextGlobal {
			result = append(result, shortcut)
		}
	}
	return result
}

// Clear removes all registered shortcuts
func (sr *ShortcutRegistry) Clear() {
	sr.shortcuts = []Shortcut{}
}
