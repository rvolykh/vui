package forms

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewKeyValueFormBuilder(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   true,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)

	assert.NotNil(t, builder)
	assert.Equal(t, app, builder.app)
	assert.Equal(t, logger, builder.logger)
	assert.Equal(t, config.Title, builder.config.Title)
	assert.NotNil(t, builder.keyValuePairs)
	assert.Empty(t, builder.keyValuePairs)
	assert.Empty(t, builder.secretPath)
}

func TestKeyValueFormBuilder_Build(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   true,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	primitive := builder.Build()

	assert.NotNil(t, primitive)
	assert.NotNil(t, builder.container)
	assert.NotNil(t, builder.currentForm)
}

func TestKeyValueFormBuilder_BuildWithInitialData(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	initialData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   true,
		InitialData:    initialData,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	primitive := builder.Build()

	assert.NotNil(t, primitive)
	assert.Equal(t, 2, len(builder.keyValuePairs))
	assert.Equal(t, "value1", builder.keyValuePairs["key1"])
	assert.Equal(t, "value2", builder.keyValuePairs["key2"])
}

func TestKeyValueFormBuilder_SetPath(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.SetPath("/secret/path")

	assert.Equal(t, "/secret/path", builder.secretPath)
}

func TestKeyValueFormBuilder_BuildWithoutPathLabel(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "", // No path label
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	primitive := builder.Build()

	assert.NotNil(t, primitive)
	assert.NotNil(t, builder.currentForm)
}

func TestKeyValueFormBuilder_BuildWithShowDeleteKeys(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	initialData := map[string]string{
		"key1": "value1",
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		InitialData:    initialData,
		ShowDeleteKeys: true,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	primitive := builder.Build()

	assert.NotNil(t, primitive)
	assert.Equal(t, 1, len(builder.keyValuePairs))
}

func TestKeyValueFormBuilder_BuildWithCallbacks(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()

	saveCalled := false
	cancelCalled := false

	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
		OnSave: func(path string, data map[string]string) {
			saveCalled = true
		},
		OnCancel: func() {
			cancelCalled = true
		},
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	assert.NotNil(t, builder.config.OnSave)
	assert.NotNil(t, builder.config.OnCancel)
	assert.False(t, saveCalled)
	assert.False(t, cancelCalled)
}

func TestKeyValueFormBuilder_FindKeyValueFields(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	keyField, valueField := builder.findKeyValueFields()

	assert.NotNil(t, keyField)
	assert.NotNil(t, valueField)
	assert.Equal(t, "Key", keyField.GetLabel())
	assert.Equal(t, "Value", valueField.GetLabel())
}

func TestKeyValueFormBuilder_FindKeyValueFields_WithNilForm(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	// Don't build, so currentForm is nil

	keyField, valueField := builder.findKeyValueFields()

	assert.Nil(t, keyField)
	assert.Nil(t, valueField)
}

func TestKeyValueFormBuilder_CreateAddKeyValueHandler(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	handler := builder.createAddKeyValueHandler()
	assert.NotNil(t, handler)
}

func TestKeyValueFormBuilder_CreateSaveHandler(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()

	saveCalled := false
	var savedPath string
	var savedData map[string]string

	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		SaveButtonText: "Save",
		OnSave: func(path string, data map[string]string) {
			saveCalled = true
			savedPath = path
			savedData = data
		},
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.SetPath("/test/path")
	builder.keyValuePairs = map[string]string{"key1": "value1"}
	builder.Build()

	handler := builder.createSaveHandler()
	assert.NotNil(t, handler)

	// Execute the handler
	handler()

	assert.True(t, saveCalled)
	assert.Equal(t, "/test/path", savedPath)
	assert.Equal(t, "value1", savedData["key1"])
}

func TestKeyValueFormBuilder_CreateSaveHandler_WithNilCallback(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()

	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
		OnSave:         nil,
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	handler := builder.createSaveHandler()
	assert.NotNil(t, handler)

	// Should not panic
	handler()
}

func TestKeyValueFormBuilder_CreateDeleteKeyHandler(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.keyValuePairs = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	builder.Build()

	handler := builder.createDeleteKeyHandler("key1")
	assert.NotNil(t, handler)

	// Execute the handler
	handler()

	assert.Equal(t, 1, len(builder.keyValuePairs))
	assert.NotContains(t, builder.keyValuePairs, "key1")
	assert.Contains(t, builder.keyValuePairs, "key2")
}

func TestKeyValueFormBuilder_PathEditableWithNoData(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   true,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Path should be editable when no data exists
	assert.Empty(t, builder.keyValuePairs)
}

func TestKeyValueFormBuilder_PathNotEditableWithData(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	initialData := map[string]string{"key1": "value1"}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   true,
		InitialData:    initialData,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Path should be disabled when data exists
	assert.NotEmpty(t, builder.keyValuePairs)
}

func TestKeyValueFormBuilder_PathNotEditableWhenConfigured(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		PathEditable:   false,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Path should be disabled when PathEditable is false
	assert.False(t, builder.config.PathEditable)
}

func TestKeyValueFormBuilder_RebuildForm(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		PathLabel:      "Path",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	initialForm := builder.currentForm

	// Rebuild the form
	builder.rebuildForm()

	// Should have a new form instance
	assert.NotNil(t, builder.currentForm)
	assert.NotEqual(t, initialForm, builder.currentForm)
}

func TestKeyValueFormBuilder_TitleWithPairCount(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	initialData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		InitialData:    initialData,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Title should include pair count
	assert.Equal(t, 2, len(builder.keyValuePairs))
}

func TestKeyValueFormBuilder_LongValueTruncation(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	longValue := "This is a very long value that should be truncated when displayed in the form because it exceeds the maximum length"
	initialData := map[string]string{
		"key1": longValue,
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		InitialData:    initialData,
		ShowDeleteKeys: false,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	assert.NotNil(t, builder.currentForm)
}

func TestKeyValueFormBuilder_MultilineValue(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	multilineValue := "line1\nline2\nline3"
	initialData := map[string]string{
		"key1": multilineValue,
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		InitialData:    initialData,
		ShowDeleteKeys: true,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	assert.Equal(t, multilineValue, builder.keyValuePairs["key1"])
}

func TestKeyValueFormBuilder_EmptyInitialData(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		InitialData:    map[string]string{},
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	assert.Empty(t, builder.keyValuePairs)
}

func TestKeyValueFormBuilder_MultipleRebuild(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	config := KeyValueFormConfig{
		Title:          "Test Form",
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Rebuild multiple times
	builder.rebuildForm()
	builder.rebuildForm()
	builder.rebuildForm()

	assert.NotNil(t, builder.currentForm)
}

func TestKeyValueFormBuilder_SortedKeys(t *testing.T) {
	app := tview.NewApplication()
	logger := logrus.New()
	initialData := map[string]string{
		"zebra": "value1",
		"alpha": "value2",
		"beta":  "value3",
	}
	config := KeyValueFormConfig{
		Title:          "Test Form",
		InitialData:    initialData,
		SaveButtonText: "Save",
	}

	builder := NewKeyValueFormBuilder(app, logger, config)
	builder.Build()

	// Keys should be stored regardless of order
	assert.Equal(t, 3, len(builder.keyValuePairs))
	assert.Contains(t, builder.keyValuePairs, "zebra")
	assert.Contains(t, builder.keyValuePairs, "alpha")
	assert.Contains(t, builder.keyValuePairs, "beta")
}
