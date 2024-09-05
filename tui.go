package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listHeight = 15

	progressPadding  = 2
	progressMaxWidth = 80
	textMaxWidth = 80
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("170"))

	focusedStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	noStyle       = lipgloss.NewStyle()
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
	helpStyle     = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list   list.Model
	choice string

	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode

	chosen   bool
	finished bool
	quitting bool
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	if !m.chosen {
		return UpdateList(msg, m)
	}
	if !m.finished {
		return UpdateInputs(msg, m)
	}

	return m, nil
}

func (m model) View() string {
	var s string
	if m.quitting {
		return "\n Exit"
	}
	if !m.chosen {
		s = listView(m)
	} else {
		s = inputsView(m)
	}
	if len(m.inputs) == 1 {
		prompt := fmt.Sprintf("Para obter a token, faca o login no site. Abra o modo de desenvolvedor, procure a aba \"Aplicativo\", na sessão \"Armazenamento\", Cookies e copie e cole aqui o valor dos cookies __RequestVerificationToken e .ASPXAUTH")
		prompt = Wordwrap(prompt, textMaxWidth)
		s = fmt.Sprintf("%s\n\n%s", prompt, s)
	}

	return mainStyle.Render("\n" + s + "\n\n")
}

func UpdateList(msg tea.Msg, m *model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				m.chosen = true

				if m.list.Index() == 0 {
					m.inputs = make([]textinput.Model, 2)
					var t textinput.Model

					for i := range m.inputs {
						t = textinput.New()
						t.Cursor.Style = cursorStyle
						t.CharLimit = 255

						switch i {
						case 0:
							t.Placeholder = "Username"
							t.Focus()
							t.PromptStyle = focusedStyle
							t.TextStyle = focusedStyle
						case 1:
							t.Placeholder = "Password"
							t.EchoMode = textinput.EchoPassword
							t.EchoCharacter = '*'
						}

						m.inputs[i] = t
					}
				} else if m.list.Index() == 1 {
					m.inputs = make([]textinput.Model, 3)
					var t textinput.Model

					for i := range m.inputs {
						t = textinput.New()
						t.Cursor.Style = cursorStyle
						t.CharLimit = 255

						switch i {
						case 0:
							t.Placeholder = "Header Request Verification Token"
							t.Focus()
							t.PromptStyle = focusedStyle
							t.TextStyle = focusedStyle
							t.EchoMode = textinput.EchoPassword
							t.EchoCharacter = '*'

						case 1:
							t.Placeholder = "Form Request Verification Token"
							t.Focus()
							t.PromptStyle = focusedStyle
							t.TextStyle = focusedStyle
							t.EchoMode = textinput.EchoPassword
							t.EchoCharacter = '*'

						case 2:
							t.Placeholder = ".ASPXAUTH"
							t.Focus()
							t.PromptStyle = focusedStyle
							t.TextStyle = focusedStyle
							t.EchoMode = textinput.EchoPassword
							t.EchoCharacter = '*'
						}

						m.inputs[i] = t
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func UpdateInputs(msg tea.Msg, m *model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}

				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func listView(m model) string {
	if m.quitting {
		return quitTextStyle.Render("Exit")
	}

	return "\n" + m.list.View()
}

func inputsView(m model) string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}

	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}
