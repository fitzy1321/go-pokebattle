package mvu

import (
	"fmt"
	"time"

	"pogomon/store"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"gorm.io/gorm"
)

// fields here should be related to view, render, update states
type (
	// * Top Level Bubbletea Model
	AppModel struct {
		width, height int
		viewState     viewState
		pokemonList   list.Model
		internalAppState
	}

	internalAppState struct {
		DB          *gorm.DB
		currentFile *store.UserSave
		saveFiles   []saveFileStart
		pokedex     []store.Pokemon
	}

	saveFileStart struct {
		ID   uint
		Name string
	}

	keyMap struct {
		Enter key.Binding
		// Back  key.Binding
		Quit key.Binding
	}

	tickMsg          struct{}
	titleTickDoneMsg struct{}
)

func NewAppModel(db *gorm.DB) (*AppModel, error) {
	var saveFileStarts []saveFileStart
	result := db.Model(&store.UserSave{}).Select("id", "name").Scan(&saveFileStarts)
	if result.Error != nil {
		return nil, fmt.Errorf("There was a problem loading save files: %+v\n", result.Error)
	}

	pokedex, err := store.GetPokemon(db)
	if err != nil {
		return nil, err
	}
	var pitems []list.Item = make([]list.Item, len(pokedex))
	for i, p := range pokedex {
		pitems[i] = p
	}

	return &AppModel{
		width: 0, height: 0,
		viewState:   titleView,
		pokemonList: list.New(pitems, list.NewDefaultDelegate(), 0, 0),
		internalAppState: internalAppState{
			DB:          db,
			saveFiles:   saveFileStarts,
			currentFile: nil,
			pokedex:     pokedex,
		},
	}, nil
}

// Gets called once "on start", I think
func (m AppModel) Init() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return titleTickDoneMsg{}
	})
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case titleTickDoneMsg:
		if m.viewState == titleView {
			m.viewState = pokemonListView
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.pokemonList.SetSize(msg.Width, msg.Height)
		var cmd tea.Cmd
		m.pokemonList, cmd = m.pokemonList.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch {
		// case key.Matches(t, keys.Enter):
		case key.Matches(msg, keys.Quit, keys.Enter):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AppModel) View() tea.View {
	var view tea.View
	if m.viewState == titleView {
		view = tea.NewView("Welcome to Pokebattle TUI!\nPress ctrl+q to quit")
	}
	if m.viewState == saveFileOrNewView {
		view = tea.NewView("I have no idea what I'm doing ...\n  press q or ctrl+c to quit, I guess ...\n")
	}

	if m.viewState == pokemonListView {
		view = tea.NewView(m.pokemonList.View())
	}

	return view
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
	pokemonListView
)
