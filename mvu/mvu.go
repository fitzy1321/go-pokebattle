package mvu

import (
	"fmt"
	"go-pokebattle/sqlmodels"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/key"
	"gorm.io/gorm"
)

func InitModel(db *gorm.DB) (*AppModel, error) {
	var saveFileStarts []saveFileStart
	result := db.Model(&sqlmodels.UserSave{}).Select("id", "name").Scan(&saveFileStarts)
	if result.Error != nil {
		return nil, fmt.Errorf("There was a problem loading save files: %+v\n", result.Error)
	}

	return &AppModel{
		width: 0, height: 0,
		internal: &internalAppState{
			DB:          db,
			saveFiles:   saveFileStarts,
			currentFile: nil,
		},
		viewState: titleView,
	}, nil
}

type keyMap struct {
	Enter key.Binding
	// Back  key.Binding
	Quit key.Binding
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	// Back: key.NewBinding(
	// 	key.WithKeys("esc", "backspace"),
	// 	key.WithHelp("esc", "back"),
	// ),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "ctrl+d"),
		key.WithHelp("q", "quit"),
	),
}

type viewState int

const (
	titleView viewState = iota
	saveFileOrNewView
)

type saveFileStart struct {
	ID   uint
	Name string
}
type internalAppState struct {
	DB          *gorm.DB
	saveFiles   []saveFileStart
	currentFile *sqlmodels.UserSave
}

// fields here should be related to view, render, update states
type AppModel struct {
	width, height uint
	internal      *internalAppState
	viewState     viewState
}

type titleTickDone struct{}

func (m AppModel) Init() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return titleTickDone{}
	})
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t := msg.(type) {
	case titleTickDone:
		if m.viewState == titleView {
			m.viewState = saveFileOrNewView
			return m, nil
		}
	// case tea.WindowSizeMsg:
	// 	m.tuiWidth, m.tuiHeight = uint(t.Width), uint(t.Height)
	case tea.KeyMsg:
		switch {
		// case key.Matches(t, keys.Enter):
		case key.Matches(t, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AppModel) View() tea.View {
	var view tea.View
	if m.viewState == titleView {
		view = tea.NewView("Welcome to Pokebattle TUI!")
	}
	if m.viewState == saveFileOrNewView {
		view = tea.NewView("I have no idea what I'm doing ...\n  press q or ctrl+c to quit, I guess ...\n")
	}

	return view
}
