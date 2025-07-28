package navigation

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMenuOption_Basic(t *testing.T) {
	option := MenuOption{
		ID:          "test",
		Label:       "Test Option",
		Description: "A test option",
		Action:      "test_action",
		Enabled:     true,
	}

	assert.Equal(t, "test", option.ID)
	assert.Equal(t, "Test Option", option.Label)
	assert.Equal(t, "A test option", option.Description)
	assert.Equal(t, "test_action", option.Action)
	assert.True(t, option.Enabled)
}

func TestMenuBuilder_Basic(t *testing.T) {
	menu := NewMenuBuilder("Test Menu").
		AddOption("opt1", "Option 1", "First option", "action1").
		AddOption("opt2", "Option 2", "Second option", "action2").
		SetShowNumbers(true).
		SetAllowBack(true).
		SetAllowQuit(true).
		Build()

	assert.Equal(t, "Test Menu", menu.Title)
	assert.Len(t, menu.Options, 2)
	assert.True(t, menu.ShowNumbers)
	assert.True(t, menu.AllowBack)
	assert.True(t, menu.AllowQuit)

	// Check first option
	assert.Equal(t, "opt1", menu.Options[0].ID)
	assert.Equal(t, "Option 1", menu.Options[0].Label)
	assert.Equal(t, "action1", menu.Options[0].Action)
	assert.True(t, menu.Options[0].Enabled)
}

func TestMenuBuilder_WithSeparator(t *testing.T) {
	menu := NewMenuBuilder("Test Menu").
		AddOption("opt1", "Option 1", "", "action1").
		AddSeparator().
		AddOption("opt2", "Option 2", "", "action2").
		Build()

	assert.Len(t, menu.Options, 3)
	assert.False(t, menu.Options[1].Enabled) // Separator should be disabled
	assert.Contains(t, menu.Options[1].Label, "──")
}

func TestMenuBuilder_Configuration(t *testing.T) {
	menu := NewMenuBuilder("Test").
		SetShowNumbers(false).
		SetAllowBack(false).
		SetAllowQuit(false).
		SetShowHelp(false).
		Build()

	assert.False(t, menu.ShowNumbers)
	assert.False(t, menu.AllowBack)
	assert.False(t, menu.AllowQuit)
	assert.False(t, menu.ShowHelp)
}

func TestMenuDisplay_ProcessInput_Numbers(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		ShowNumbers: true,
		Options: []MenuOption{
			{ID: "opt1", Label: "Option 1", Action: "action1", Enabled: true},
			{ID: "opt2", Label: "Option 2", Action: "action2", Enabled: true},
		},
	}

	// Test valid number selection
	result := display.processInput(menu, "1")
	require.NotNil(t, result)
	assert.Equal(t, "opt1", result.SelectedOption.ID)
	assert.Equal(t, "action1", result.Action)

	result = display.processInput(menu, "2")
	require.NotNil(t, result)
	assert.Equal(t, "opt2", result.SelectedOption.ID)
	assert.Equal(t, "action2", result.Action)

	// Test invalid number selection
	result = display.processInput(menu, "0")
	assert.Nil(t, result)
	result = display.processInput(menu, "3")
	assert.Nil(t, result)
	result = display.processInput(menu, "99")
	assert.Nil(t, result)
}

func TestMenuDisplay_ProcessInput_Shortcuts(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		AllowBack: true,
		AllowQuit: true,
		ShowHelp:  true,
	}

	// Test quit shortcuts
	testCases := []struct {
		input    string
		expected string
	}{
		{"q", "quit"},
		{"quit", "quit"},
		{"exit", "quit"},
		{"Q", "quit"}, // Case insensitive
		{"b", "back"},
		{"back", "back"},
		{"B", "back"}, // Case insensitive
		{"h", "help"},
		{"help", "help"},
		{"H", "help"}, // Case insensitive
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := display.processInput(menu, tc.input)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, result.Action)
		})
	}
}

func TestMenuDisplay_ProcessInput_DisabledShortcuts(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		AllowBack: false,
		AllowQuit: false,
		ShowHelp:  false,
	}

	// Test that disabled shortcuts don't work
	shortcuts := []string{"q", "quit", "b", "back", "h", "help"}
	for _, shortcut := range shortcuts {
		result := display.processInput(menu, shortcut)
		assert.Nil(t, result, "Shortcut %s should be disabled", shortcut)
	}
}

func TestMenuDisplay_ProcessInput_ByID(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		ShowNumbers: false,
		Options: []MenuOption{
			{ID: "start", Label: "Start Project", Action: "start_action", Enabled: true},
			{ID: "status", Label: "Check Status", Action: "status_action", Enabled: true},
		},
	}

	// Test selection by ID
	result := display.processInput(menu, "start")
	require.NotNil(t, result)
	assert.Equal(t, "start", result.SelectedOption.ID)
	assert.Equal(t, "start_action", result.Action)

	// Test case insensitive
	result = display.processInput(menu, "STATUS")
	require.NotNil(t, result)
	assert.Equal(t, "status", result.SelectedOption.ID)
	assert.Equal(t, "status_action", result.Action)
}

func TestMenuDisplay_ProcessInput_ByLabel(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		Options: []MenuOption{
			{ID: "opt1", Label: "Start", Action: "start_action", Enabled: true},
			{ID: "opt2", Label: "Exit", Action: "exit_action", Enabled: true},
		},
	}

	// Test selection by label
	result := display.processInput(menu, "Start")
	require.NotNil(t, result)
	assert.Equal(t, "opt1", result.SelectedOption.ID)

	// Test case insensitive
	result = display.processInput(menu, "exit")
	require.NotNil(t, result)
	assert.Equal(t, "opt2", result.SelectedOption.ID)
}

func TestMenuDisplay_ProcessInput_DisabledOptions(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		ShowNumbers: true,
		Options: []MenuOption{
			{ID: "opt1", Label: "Enabled", Action: "action1", Enabled: true},
			{ID: "opt2", Label: "Disabled", Action: "action2", Enabled: false},
		},
	}

	// Test that enabled option works
	result := display.processInput(menu, "1")
	require.NotNil(t, result)
	assert.Equal(t, "opt1", result.SelectedOption.ID)

	// Test that disabled option doesn't work
	result = display.processInput(menu, "2")
	assert.Nil(t, result)

	// Test that disabled option doesn't work by ID either
	result = display.processInput(menu, "opt2")
	assert.Nil(t, result)
}

func TestMenuDisplay_ProcessInput_EmptyInput(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		Options: []MenuOption{
			{ID: "opt1", Label: "Option 1", Action: "action1", Enabled: true},
		},
	}

	// Empty input should return nil
	result := display.processInput(menu, "")
	assert.Nil(t, result)

	result = display.processInput(menu, "   ")
	assert.Nil(t, result)
}

func TestMenuDisplay_ProcessInput_InvalidInput(t *testing.T) {
	display := NewMenuDisplay()
	menu := &Menu{
		ShowNumbers: true,
		Options: []MenuOption{
			{ID: "opt1", Label: "Option 1", Action: "action1", Enabled: true},
		},
	}

	// Various invalid inputs
	invalidInputs := []string{
		"invalid",
		"999",
		"abc",
		"!@#",
		"option99",
	}

	for _, input := range invalidInputs {
		result := display.processInput(menu, input)
		assert.Nil(t, result, "Input %s should be invalid", input)
	}
}

// Test helper functions
func TestMenuDisplay_HelperFunctions(t *testing.T) {
	display := NewMenuDisplay()

	// Test message functions (these mainly test that they don't panic)
	display.ShowMessage("Test message")
	display.ShowError("Test error")
	display.ShowSuccess("Test success")
	display.ShowWarning("Test warning")
}

func TestMenuDisplay_WithMockInput(t *testing.T) {
	// Create a string reader for mock input
	input := "1\n"
	reader := bufio.NewReader(strings.NewReader(input))
	display := &MenuDisplay{reader: reader}

	menu := NewMenuBuilder("Test Menu").
		AddOption("opt1", "Option 1", "", "action1").
		Build()

	// Test that input selects option 1
	result, err := display.Show(menu)
	require.NoError(t, err)
	assert.Equal(t, "opt1", result.SelectedOption.ID)
}

func TestMenuDisplay_Confirm(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"n", false},
		{"N", false},
		{"no", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input + "\n"))
			display := &MenuDisplay{reader: reader}

			result, err := display.Confirm("Test question")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMenuDisplay_PromptString(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("test input\n"))
	display := &MenuDisplay{reader: reader}

	result, err := display.PromptString("Enter something")
	require.NoError(t, err)
	assert.Equal(t, "test input", result)
}

func TestMenuDisplay_PromptStringWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "use input",
			input:        "user input",
			defaultValue: "default",
			expected:     "user input",
		},
		{
			name:         "use default",
			input:        "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty input no default",
			input:        "",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input + "\n"))
			display := &MenuDisplay{reader: reader}

			result, err := display.PromptStringWithDefault("Enter something", tt.defaultValue)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
