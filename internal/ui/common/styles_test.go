package common

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	assert.Equal(t, tcell.ColorBlack, theme.Background)
	assert.Equal(t, tcell.ColorDarkSlateGray, theme.ContrastBackground)
	assert.Equal(t, tcell.ColorDarkGray, theme.MoreContrastBackground)
	assert.Equal(t, tcell.ColorDarkCyan, theme.Border)
	assert.Equal(t, tcell.ColorWhite, theme.Title)
	assert.Equal(t, tcell.ColorDarkCyan, theme.Graphics)
	assert.Equal(t, tcell.ColorWhite, theme.PrimaryText)
	assert.Equal(t, tcell.ColorLightGray, theme.SecondaryText)
	assert.Equal(t, tcell.ColorGray, theme.TertiaryText)
	assert.Equal(t, tcell.ColorBlack, theme.InverseText)
	assert.Equal(t, tcell.ColorDarkGray, theme.ContrastSecondaryText)
}

func TestInitializeTheme(t *testing.T) {
	// Store original values
	origBackground := tview.Styles.PrimitiveBackgroundColor
	origBorder := tview.Styles.BorderColor

	// Initialize theme
	InitializeTheme()

	// Verify theme was applied
	assert.Equal(t, tcell.ColorBlack, tview.Styles.PrimitiveBackgroundColor)
	assert.Equal(t, tcell.ColorDarkCyan, tview.Styles.BorderColor)
	assert.Equal(t, tcell.ColorWhite, tview.Styles.TitleColor)

	// Restore original values
	tview.Styles.PrimitiveBackgroundColor = origBackground
	tview.Styles.BorderColor = origBorder
}

func TestGetFormStyle(t *testing.T) {
	style := GetFormStyle()

	assert.Equal(t, tcell.ColorDarkSlateGray, style.FieldBackground)
	assert.Equal(t, tcell.ColorWhite, style.FieldText)
	assert.Equal(t, tcell.ColorLightGray, style.Label)
}

func TestGetButtonStyle(t *testing.T) {
	style := GetButtonStyle()

	assert.Equal(t, tcell.ColorDarkGray, style.Background)
	assert.Equal(t, tcell.ColorWhite, style.Text)
	assert.NotNil(t, style.ActivatedStyle)
}

func TestGetDeleteButtonStyle(t *testing.T) {
	style := GetDeleteButtonStyle()

	assert.Equal(t, tcell.ColorDarkGray, style.Background)
	assert.Equal(t, tcell.ColorWhite, style.Text)
	assert.NotNil(t, style.ActivatedStyle)
}

func TestApplyFormStyle(t *testing.T) {
	form := tview.NewForm()
	ApplyFormStyle(form)

	// We can't directly test the applied colors on the form,
	// but we can verify the function doesn't panic
	assert.NotNil(t, form)
}

func TestApplyButtonStyle(t *testing.T) {
	form := tview.NewForm()
	ApplyButtonStyle(form)

	// We can't directly test the applied colors on the form,
	// but we can verify the function doesn't panic
	assert.NotNil(t, form)
}

func TestApplyModalStyle(t *testing.T) {
	modal := tview.NewModal()
	ApplyModalStyle(modal)

	// We can't directly test the applied colors on the modal,
	// but we can verify the function doesn't panic
	assert.NotNil(t, modal)
}
