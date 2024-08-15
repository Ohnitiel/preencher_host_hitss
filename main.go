package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 10

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("170"))
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
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

type Activities struct {
	Id_CapturaActividad string
	// Id_Actividad        string
	HorasCapturadas string
	Comentario      string
	HorasExtras     bool
	HorasNocturnas  bool
	Bloqueada       bool
}

type FillData struct {
	Id_Proyecto                int
	Id_Recurso                 string
	FechaDia                   string
	Comentario                 string
	Actividades                []Activities
	pantallaCaptura            bool
	Latitude                   float64
	Longitude                  float64
	__RequestVerificationToken string
}

func login() (string, error) {
	login_url := "https://host.globalhitss.com/"
	response, err := http.PostForm(
		login_url,
		url.Values{
			"UserName": {},
			"Password": {},
			"Language": {"pt"},
			"bandera":  {"1"},
		},
	)
	if err != nil {
		fmt.Println("Erro ao realizar o login.")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cookies := response.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "__RequestVerificationToken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("Não foi possível encontrar o token")
}

func fillHours(requestToken string, calendar Calendar) {
	var data FillData
	activities := make([]Activities, 1)

	activities[0] = Activities{
		Id_CapturaActividad: "0",
		// Id_Actividad:        "973214",
		HorasCapturadas: "8.0",
		Comentario:      "t",
		HorasExtras:     false,
		HorasNocturnas:  false,
		Bloqueada:       false,
	}

	url := "https://host.globalhitss.com/Horas/ActualizaActividades2"
	projectId := 73556
	ResourceId := "50900135"

	current_time := time.Now()
	startDate := time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		if !calendar[d] {
			continue
		}

		data = FillData{
			Id_Proyecto:                projectId,
			Id_Recurso:                 ResourceId,
			FechaDia:                   d.Format("2006-01-02"),
			Comentario:                 "",
			Actividades:                activities,
			pantallaCaptura:            true,
			Latitude:                   0,
			Longitude:                  0,
			__RequestVerificationToken: requestToken,
		}

		payload, _ := json.Marshal(data)

		_, err := http.Post(url, "application/json",
			bytes.NewBuffer(payload))
		if err != nil {
			fmt.Printf("Erro ao enviar os dados, dia %s: %s\n",
				d.Format("2006-01-02"), err.Error())
		}
	}
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		if m.list.Index() == 0 {
			return quitTextStyle.Render("Username password")
		} else if m.list.Index() == 1 {
			return quitTextStyle.Render("Token")
		}
	}
	if m.quitting {
		return quitTextStyle.Render("Exit")
	}

	return "\n" + m.list.View()
}

func main() {
	items := []list.Item{
		item("Login com usuário e senha"),
		item("Login com token"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Como deseja realizar o login?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Algo deu errado: %v\n", err)
		os.Exit(1)
	}
}

func main2() {

	cookie, err := login()
	if err != nil {
		fmt.Print(err)
		os.Exit(4)
	}

	fillHours(cookie, CalendarForYear(time.Now().Year()))
}
