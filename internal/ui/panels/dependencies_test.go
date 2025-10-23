package panels

// Mock dialog service for testing
type mockDialogService struct{}

func (m *mockDialogService) ShowInfo(title, message string, callback func()) {}
func (m *mockDialogService) ShowError(message string, callback func())       {}
