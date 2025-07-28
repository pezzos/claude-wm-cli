package navigation

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MenuOption represents a single menu option
type MenuOption struct {
	ID          string // Unique identifier for the option
	Label       string // Display text for the option
	Description string // Optional detailed description
	Action      string // Action identifier to execute
	Enabled     bool   // Whether this option is currently available
}

// MenuResult represents the result of a menu interaction
type MenuResult struct {
	SelectedOption *MenuOption
	Action         string // Special actions like "back", "quit", "help"
	Input          string // Raw user input
}

// Menu represents an interactive menu system
type Menu struct {
	Title       string
	Options     []MenuOption
	ShowNumbers bool // Whether to show numbered options
	ShowHelp    bool // Whether to show help text
	AllowBack   bool // Whether to allow back navigation
	AllowQuit   bool // Whether to allow quit
}

// MenuDisplay handles the presentation and interaction of menus
type MenuDisplay struct {
	reader *bufio.Reader
}

// NewMenuDisplay creates a new menu display handler
func NewMenuDisplay() *MenuDisplay {
	return &MenuDisplay{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Show displays the menu and handles user interaction
func (md *MenuDisplay) Show(menu *Menu) (*MenuResult, error) {
	for {
		// Display the menu
		md.displayMenu(menu)

		// Get user input
		input, err := md.getUserInput()
		if err != nil {
			return nil, fmt.Errorf("failed to get user input: %w", err)
		}

		// Process the input
		result := md.processInput(menu, input)
		if result != nil {
			return result, nil
		}

		// If we reach here, input was invalid - show error and retry
		fmt.Println("\n❌ Invalid selection. Please try again.")
	}
}

// displayMenu renders the menu to the console
func (md *MenuDisplay) displayMenu(menu *Menu) {
	// Clear screen (optional - can be made configurable)
	// fmt.Print("\033[2J\033[H")

	// Display title
	if menu.Title != "" {
		fmt.Printf("\n═══ %s ═══\n\n", menu.Title)
	}

	// Display options
	if menu.ShowNumbers {
		for i, option := range menu.Options {
			if !option.Enabled {
				continue
			}

			fmt.Printf("  %d) %s", i+1, option.Label)
			if option.Description != "" {
				fmt.Printf(" - %s", option.Description)
			}
			fmt.Println()
		}
	} else {
		for _, option := range menu.Options {
			if !option.Enabled {
				continue
			}

			fmt.Printf("  • %s", option.Label)
			if option.Description != "" {
				fmt.Printf(" - %s", option.Description)
			}
			fmt.Println()
		}
	}

	// Display navigation options
	fmt.Println()
	var navOptions []string

	if menu.AllowBack {
		navOptions = append(navOptions, "b) Back")
	}
	if menu.AllowQuit {
		navOptions = append(navOptions, "q) Quit")
	}
	if menu.ShowHelp {
		navOptions = append(navOptions, "h) Help")
	}

	if len(navOptions) > 0 {
		fmt.Printf("  %s\n", strings.Join(navOptions, "  "))
	}

	fmt.Print("\nSelect an option: ")
}

// getUserInput reads user input from stdin
func (md *MenuDisplay) getUserInput() (string, error) {
	input, err := md.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// processInput processes user input and returns appropriate result
func (md *MenuDisplay) processInput(menu *Menu, input string) *MenuResult {
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle special shortcuts
	switch input {
	case "q", "quit", "exit":
		if menu.AllowQuit {
			return &MenuResult{Action: "quit", Input: input}
		}
	case "b", "back":
		if menu.AllowBack {
			return &MenuResult{Action: "back", Input: input}
		}
	case "h", "help":
		if menu.ShowHelp {
			return &MenuResult{Action: "help", Input: input}
		}
	case "":
		// Empty input - show error
		return nil
	}

	// Try to parse as number (for numbered menus)
	if menu.ShowNumbers {
		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= len(menu.Options) {
				option := &menu.Options[num-1]
				if option.Enabled {
					return &MenuResult{
						SelectedOption: option,
						Action:         option.Action,
						Input:          input,
					}
				}
			}
		}
	}

	// Try to match by option ID or label (case-insensitive)
	for i, option := range menu.Options {
		if !option.Enabled {
			continue
		}

		if strings.EqualFold(option.ID, input) ||
			strings.EqualFold(option.Label, input) {
			return &MenuResult{
				SelectedOption: &menu.Options[i],
				Action:         option.Action,
				Input:          input,
			}
		}
	}

	return nil // Invalid input
}

// ShowMessage displays a message to the user
func (md *MenuDisplay) ShowMessage(message string) {
	fmt.Printf("\n%s\n", message)
}

// ShowError displays an error message to the user
func (md *MenuDisplay) ShowError(message string) {
	fmt.Printf("\n❌ Error: %s\n", message)
}

// ShowSuccess displays a success message to the user
func (md *MenuDisplay) ShowSuccess(message string) {
	fmt.Printf("\n✅ %s\n", message)
}

// ShowWarning displays a warning message to the user
func (md *MenuDisplay) ShowWarning(message string) {
	fmt.Printf("\n⚠️  Warning: %s\n", message)
}

// Confirm asks the user for yes/no confirmation
func (md *MenuDisplay) Confirm(message string) (bool, error) {
	fmt.Printf("%s (y/N): ", message)

	input, err := md.getUserInput()
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}

// PromptString prompts the user for a string input
func (md *MenuDisplay) PromptString(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)
	return md.getUserInput()
}

// PromptStringWithDefault prompts for string input with a default value
func (md *MenuDisplay) PromptStringWithDefault(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := md.getUserInput()
	if err != nil {
		return "", err
	}

	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return input, nil
}

// WaitForKeyPress waits for the user to press any key
func (md *MenuDisplay) WaitForKeyPress(message string) error {
	if message == "" {
		message = "Press any key to continue..."
	}

	fmt.Printf("\n%s ", message)
	_, err := md.reader.ReadString('\n')
	return err
}

// MenuBuilder provides a fluent interface for building menus
type MenuBuilder struct {
	menu *Menu
}

// NewMenuBuilder creates a new menu builder
func NewMenuBuilder(title string) *MenuBuilder {
	return &MenuBuilder{
		menu: &Menu{
			Title:       title,
			Options:     []MenuOption{},
			ShowNumbers: true,
			ShowHelp:    true,
			AllowBack:   true,
			AllowQuit:   true,
		},
	}
}

// AddOption adds an option to the menu
func (mb *MenuBuilder) AddOption(id, label, description, action string) *MenuBuilder {
	mb.menu.Options = append(mb.menu.Options, MenuOption{
		ID:          id,
		Label:       label,
		Description: description,
		Action:      action,
		Enabled:     true,
	})
	return mb
}

// AddSeparator adds a visual separator (disabled option)
func (mb *MenuBuilder) AddSeparator() *MenuBuilder {
	mb.menu.Options = append(mb.menu.Options, MenuOption{
		Label:   "────────────────────────",
		Enabled: false,
	})
	return mb
}

// SetShowNumbers controls whether to show numbered options
func (mb *MenuBuilder) SetShowNumbers(show bool) *MenuBuilder {
	mb.menu.ShowNumbers = show
	return mb
}

// SetAllowBack controls whether to allow back navigation
func (mb *MenuBuilder) SetAllowBack(allow bool) *MenuBuilder {
	mb.menu.AllowBack = allow
	return mb
}

// SetAllowQuit controls whether to allow quit
func (mb *MenuBuilder) SetAllowQuit(allow bool) *MenuBuilder {
	mb.menu.AllowQuit = allow
	return mb
}

// SetShowHelp controls whether to show help option
func (mb *MenuBuilder) SetShowHelp(show bool) *MenuBuilder {
	mb.menu.ShowHelp = show
	return mb
}

// Build returns the constructed menu
func (mb *MenuBuilder) Build() *Menu {
	return mb.menu
}
