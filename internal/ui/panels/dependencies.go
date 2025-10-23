package panels

// dialogService is the interface for showing dialogs
type dialogService interface {
	ShowInfo(title, message string, callback func())
}
