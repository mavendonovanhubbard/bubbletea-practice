package main

import (
  "fmt"
  "time"
  "os"
  "strings"
  "net/http"
  "net/url"
  "context"
  "encoding/json"
  tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/textinput"
)

type Definitions struct {
	List []struct {
		Definition  string    `json:"definition"`
		Permalink   string    `json:"permalink"`
		ThumbsUp    int       `json:"thumbs_up"`
		Author      string    `json:"author"`
		Word        string    `json:"word"`
		Defid       int       `json:"defid"`
		CurrentVote string    `json:"current_vote"`
		WrittenOn   time.Time `json:"written_on"`
		Example     string    `json:"example"`
		ThumbsDown  int       `json:"thumbs_down"`
	} `json:"list"`
}

type SearchResults struct {
  Definitions Definitions 
  Error error
}

type Model struct {
  TextInput textinput.Model
  ModelSearchResults SearchResults
}

func queryAPI(searchTerm string) tea.Cmd {
  return func () tea.Msg {
    apiUrl := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%s", url.QueryEscape(searchTerm))
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl, nil)
    if err != nil {
      return SearchResults {
        Error: err,
      }
    }

    res, err := http.DefaultClient.Do(req)
    if err != nil {
      return SearchResults {
        Error: err,
      }
    }

    defer res.Body.Close()
   
    var definitions Definitions

    err = json.NewDecoder(res.Body).Decode(&definitions)
    if err != nil {
      return SearchResults {
        Error: err,
      }
    }

    return SearchResults {
      Definitions: definitions,
    }
  }
}

func NewModel() Model {
  ti := textinput.New()
	ti.Placeholder = "drip"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
  return Model{
    TextInput: ti,
  }
}

func (m Model)Init() tea.Cmd {
  return textinput.Blink
}

func (m Model)View() string {
  var sb strings.Builder
  sb.WriteString("Enter a slang term\n")
  sb.WriteString(m.TextInput.View())
  sb.WriteString("\n")

  if len(m.ModelSearchResults.Definitions.List) != 0 {
    firstDef := m.ModelSearchResults.Definitions.List[0]
    sb.WriteString("Word: ")
    sb.WriteString(firstDef.Word)
    sb.WriteString("\n")

    sb.WriteString("Definition: ")
    sb.WriteString(firstDef.Definition)
    sb.WriteString("\n")

    sb.WriteString("Example: ")
    sb.WriteString(firstDef.Example)
    sb.WriteString("\n")
  }

  if m.ModelSearchResults.Error != nil {
    sb.WriteString("Error: ")
    sb.WriteString(m.ModelSearchResults.Error.Error())
    sb.WriteString("\n")
  }

  return sb.String()
}

func (m Model)Update(msg tea.Msg) (tea.Model,tea.Cmd) {
  var cmd tea.Cmd
  switch msg := msg.(type) {
  case SearchResults:
    m.ModelSearchResults = msg
    return m, nil
  case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c":
      return m, tea.Quit
    case "enter":
      searchText := m.TextInput.Value()
      m.TextInput.Reset()
      return m, queryAPI(searchText)
    }
  }
  m.TextInput, cmd = m.TextInput.Update(msg)
  return m,cmd
}

func main(){
  p := tea.NewProgram(NewModel())
  if _, err := p.Run(); err != nil {
      fmt.Printf("Alas, there's been an error: %v", err)
      os.Exit(1)
  }
}
