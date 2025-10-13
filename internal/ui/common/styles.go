package common

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ThemeColors holds the color scheme for the application
type ThemeColors struct {
	Background             tcell.Color
	ContrastBackground     tcell.Color
	MoreContrastBackground tcell.Color
	Border                 tcell.Color
	Title                  tcell.Color
	Graphics               tcell.Color
	PrimaryText            tcell.Color
	SecondaryText          tcell.Color
	TertiaryText           tcell.Color
	InverseText            tcell.Color
	ContrastSecondaryText  tcell.Color
}

// DefaultTheme returns the default dark theme colors
func DefaultTheme() ThemeColors {
	return ThemeColors{
		Background:             tcell.ColorBlack,
		ContrastBackground:     tcell.ColorDarkSlateGray,
		MoreContrastBackground: tcell.ColorDarkGray,
		Border:                 tcell.ColorDarkCyan,
		Title:                  tcell.ColorWhite,
		Graphics:               tcell.ColorDarkCyan,
		PrimaryText:            tcell.ColorWhite,
		SecondaryText:          tcell.ColorLightGray,
		TertiaryText:           tcell.ColorGray,
		InverseText:            tcell.ColorBlack,
		ContrastSecondaryText:  tcell.ColorDarkGray,
	}
}

// InitializeTheme applies the theme colors to tview
func InitializeTheme() {
	theme := DefaultTheme()
	tview.Styles.PrimitiveBackgroundColor = theme.Background
	tview.Styles.ContrastBackgroundColor = theme.ContrastBackground
	tview.Styles.MoreContrastBackgroundColor = theme.MoreContrastBackground
	tview.Styles.BorderColor = theme.Border
	tview.Styles.TitleColor = theme.Title
	tview.Styles.GraphicsColor = theme.Graphics
	tview.Styles.PrimaryTextColor = theme.PrimaryText
	tview.Styles.SecondaryTextColor = theme.SecondaryText
	tview.Styles.TertiaryTextColor = theme.TertiaryText
	tview.Styles.InverseTextColor = theme.InverseText
	tview.Styles.ContrastSecondaryTextColor = theme.ContrastSecondaryText
}

// FormStyle holds styling configuration for forms
type FormStyle struct {
	FieldBackground tcell.Color
	FieldText       tcell.Color
	Label           tcell.Color
}

// GetFormStyle returns the standard form styling
func GetFormStyle() FormStyle {
	return FormStyle{
		FieldBackground: tcell.ColorDarkSlateGray,
		FieldText:       tcell.ColorWhite,
		Label:           tcell.ColorLightGray,
	}
}

// ButtonStyle holds styling configuration for buttons
type ButtonStyle struct {
	Background     tcell.Color
	Text           tcell.Color
	ActivatedStyle tcell.Style
}

// GetButtonStyle returns the standard button styling
func GetButtonStyle() ButtonStyle {
	return ButtonStyle{
		Background: tcell.ColorDarkGray,
		Text:       tcell.ColorWhite,
		ActivatedStyle: tcell.StyleDefault.
			Background(tcell.ColorDarkCyan).
			Foreground(tcell.ColorBlack),
	}
}

// GetDeleteButtonStyle returns styling for destructive delete buttons
func GetDeleteButtonStyle() ButtonStyle {
	return ButtonStyle{
		Background: tcell.ColorDarkGray,
		Text:       tcell.ColorWhite,
		ActivatedStyle: tcell.StyleDefault.
			Background(tcell.ColorRed).
			Foreground(tcell.ColorWhite).
			Bold(true).
			Reverse(true),
	}
}

// ApplyFormStyle applies standard styling to a form
func ApplyFormStyle(form *tview.Form) {
	style := GetFormStyle()
	form.SetFieldBackgroundColor(style.FieldBackground).
		SetFieldTextColor(style.FieldText).
		SetLabelColor(style.Label)
}

// ApplyButtonStyle applies standard styling to form buttons
func ApplyButtonStyle(form *tview.Form) {
	style := GetButtonStyle()
	form.SetButtonBackgroundColor(style.Background).
		SetButtonTextColor(style.Text).
		SetButtonActivatedStyle(style.ActivatedStyle)
}

// ApplyModalStyle applies standard styling to a modal
func ApplyModalStyle(modal *tview.Modal) {
	style := GetButtonStyle()
	modal.SetButtonBackgroundColor(style.Background).
		SetButtonTextColor(style.Text).
		SetButtonActivatedStyle(style.ActivatedStyle)
}
