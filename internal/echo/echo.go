package echo

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for beautiful CLI output
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#02BA84"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))
)

// Header prints a styled header message
func Header(message string) {
	fmt.Println(headerStyle.Render(" " + message + " "))
}

// Success prints a styled success message
func Success(message string) {
	fmt.Println(successStyle.Render(message))
}

// Successf prints a styled success message with formatting
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}

// Error prints a styled error message
func Error(message string) {
	fmt.Println(errorStyle.Render(message))
}

// Errorf prints a styled error message with formatting
func Errorf(format string, args ...interface{}) {
	fmt.Println(errorStyle.Render(fmt.Sprintf(format, args...)))
}

// Info prints a styled info message
func Info(message string) {
	fmt.Println(infoStyle.Render(message))
}

// Infof prints a styled info message with formatting
func Infof(format string, args ...interface{}) {
	fmt.Println(infoStyle.Render(fmt.Sprintf(format, args...)))
}

// HeaderStyle returns the header style for use in longer messages
func HeaderStyle() lipgloss.Style {
	return headerStyle
}

// SuccessStyle returns the success style
func SuccessStyle() lipgloss.Style {
	return successStyle
}

// ErrorStyle returns the error style
func ErrorStyle() lipgloss.Style {
	return errorStyle
}

// InfoStyle returns the info style
func InfoStyle() lipgloss.Style {
	return infoStyle
}