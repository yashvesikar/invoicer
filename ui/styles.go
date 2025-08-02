package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)
	
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)
	
	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)
	
	normalStyle = lipgloss.NewStyle()
	
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	
	cursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
	
	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	
	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)
	
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)
	
	listItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)
	
	selectedListItemStyle = lipgloss.NewStyle().
		PaddingLeft(0).
		Foreground(lipgloss.Color("205"))
	
	formLabelStyle = lipgloss.NewStyle().
		Width(15).
		Foreground(lipgloss.Color("241"))
	
	formInputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
	
	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder())
	
	tableCellStyle = lipgloss.NewStyle().
		Padding(0, 1)
	
	statusDraftStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))
	
	statusSentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))
	
	statusPaidStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))
	
	statusOverdueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))
)