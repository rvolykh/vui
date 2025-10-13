package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModalBuilder(t *testing.T) {
	builder := NewModalBuilder()
	assert.NotNil(t, builder)
	assert.Empty(t, builder.buttons)
	assert.Empty(t, builder.text)
	assert.False(t, builder.deleteStyle)
}

func TestModalBuilder_SetText(t *testing.T) {
	builder := NewModalBuilder().SetText("Test message")
	assert.Equal(t, "Test message", builder.text)
}

func TestModalBuilder_AddButton(t *testing.T) {
	builder := NewModalBuilder().
		AddButton("OK").
		AddButton("Cancel")

	assert.Len(t, builder.buttons, 2)
	assert.Equal(t, "OK", builder.buttons[0])
	assert.Equal(t, "Cancel", builder.buttons[1])
}

func TestModalBuilder_AddButtons(t *testing.T) {
	builder := NewModalBuilder().
		AddButtons("OK", "Cancel", "Help")

	assert.Len(t, builder.buttons, 3)
	assert.Equal(t, []string{"OK", "Cancel", "Help"}, builder.buttons)
}

func TestModalBuilder_SetDoneFunc(t *testing.T) {
	called := false
	builder := NewModalBuilder().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			called = true
		})

	assert.NotNil(t, builder.doneFunc)

	// Call the function to verify it works
	builder.doneFunc(0, "OK")
	assert.True(t, called)
}

func TestModalBuilder_UseDeleteStyle(t *testing.T) {
	builder := NewModalBuilder().UseDeleteStyle()
	assert.True(t, builder.deleteStyle)
}

func TestModalBuilder_Build(t *testing.T) {
	modal := NewModalBuilder().
		SetText("Test message").
		AddButtons("OK", "Cancel").
		Build()

	assert.NotNil(t, modal)
}

func TestModalBuilder_BuildWithDeleteStyle(t *testing.T) {
	modal := NewModalBuilder().
		SetText("Delete confirmation").
		AddButtons("Cancel", "Delete").
		UseDeleteStyle().
		Build()

	assert.NotNil(t, modal)
}

func TestErrorModal(t *testing.T) {
	modal := ErrorModal("Test error", func() {
		// Callback function
	})

	assert.NotNil(t, modal)
}

func TestConfirmationModal(t *testing.T) {
	modal := ConfirmationModal(
		"Are you sure?",
		func() { /* confirm callback */ },
		func() { /* cancel callback */ },
	)

	assert.NotNil(t, modal)
}

func TestDeleteConfirmationModal(t *testing.T) {
	modal := DeleteConfirmationModal(
		"test-item",
		func() { /* delete callback */ },
		func() { /* cancel callback */ },
	)

	assert.NotNil(t, modal)
}

func TestInfoModal(t *testing.T) {
	modal := InfoModal("Information", "This is a test", func() {
		// Callback function
	})

	assert.NotNil(t, modal)
}

func TestModalBuilder_Chaining(t *testing.T) {
	// Test that method chaining works correctly
	modal := NewModalBuilder().
		SetText("Chained modal").
		AddButton("First").
		AddButton("Second").
		AddButtons("Third", "Fourth").
		UseDeleteStyle().
		Build()

	assert.NotNil(t, modal)
}
