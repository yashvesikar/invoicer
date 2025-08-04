package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/models"
)

type settingsMode int

const (
	settingsModeEdit settingsMode = iota
	settingsModeSuccess
)

type SettingsModel struct {
	storage    models.Storage
	config     *config.Config
	setupModel tea.Model
	mode       settingsMode
	message    string
	err        error
}

func NewSettingsModel(storage models.Storage, cfg *config.Config) SettingsModel {
	setupModel := config.NewSettingsEditor(cfg)
	return SettingsModel{
		storage:    storage,
		config:     cfg,
		setupModel: setupModel,
		mode:       settingsModeEdit,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return m.setupModel.Init()
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case settingsModeEdit:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Check if user wants to go back to menu
			if msg.String() == "esc" {
				return NewMainMenuModel(m.storage, m.config), nil
			}
		case config.SettingsSavedMsg:
			// Settings were saved successfully
			m.mode = settingsModeSuccess
			m.message = "Settings saved successfully!"
			return m, tea.Tick(settingsSuccessTimeout, func(t time.Time) tea.Msg {
				return returnToMenuMsg{}
			})
		case config.SettingsCancelledMsg:
			// User cancelled settings edit
			return NewMainMenuModel(m.storage, m.config), nil
		case config.SettingsErrorMsg:
			// Error saving settings
			m.err = msg.Err
			return m, nil
		}

		// Pass through to setup model
		var cmd tea.Cmd
		m.setupModel, cmd = m.setupModel.Update(msg)
		return m, cmd

	case settingsModeSuccess:
		switch msg.(type) {
		case returnToMenuMsg:
			return NewMainMenuModel(m.storage, m.config), nil
		case tea.KeyMsg:
			// Any key press returns to menu
			return NewMainMenuModel(m.storage, m.config), nil
		}
	}

	return m, nil
}

func (m SettingsModel) View() string {
	switch m.mode {
	case settingsModeEdit:
		s := m.setupModel.View()
		if m.err != nil {
			s += "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		}
		return s
	case settingsModeSuccess:
		return appStyle.Render(
			titleStyle.Render("Settings") + "\n\n" +
				successStyle.Render(m.message) + "\n\n" +
				helpStyle.Render("Press any key to return to menu"),
		)
	}
	return ""
}

type returnToMenuMsg struct{}

var settingsSuccessTimeout = time.Second * 2