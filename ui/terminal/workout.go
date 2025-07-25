package terminal

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"fmt"
	"os"
	"seneschal/config"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
	"github.com/muesli/termenv"
)

// ANSIBlack
// ANSIRed
// ANSIGreen
// ANSIYellow
// ANSIBlue
// ANSIMagenta
// ANSICyan
// ANSIWhite
// ANSIBrightBlack
// ANSIBrightRed
// ANSIBrightGreen
// ANSIBrightYellow
// ANSIBrightBlue
// ANSIBrightMagenta
// ANSIBrightCyan
// ANSIBrightWhite

var (
	todoStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIYellow))
	finishStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIGreen))
	breakStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightGreen))
	workStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIMagenta))
)

type tickMsg time.Time

//go:generate stringer -type=WorkoutStatus
type WorkoutStatus int

const (
	WorkoutStatus_Break WorkoutStatus = iota
	WorkoutStatus_Work
	WorkoutStatus_Fin
)

type model struct {
	wc         *config.WorkoutConfig
	status     WorkoutStatus
	curItem    *config.WorkoutItem
	curItemIdx int
	breaking   bool
	countDown  int
	curRepeat  int
}

func initialModel(wc *config.WorkoutConfig) model {
	m := model{
		wc:         wc,
		status:     WorkoutStatus_Break,
		curItem:    &config.WorkoutItem{},
		curItemIdx: 0,
		breaking:   false,
		countDown:  0,
		curRepeat:  0,
	}
	if len(wc.ItemList) == 0 {
		m.status = WorkoutStatus_Fin
	} else {
		m.curItem = m.wc.ItemList[m.curItemIdx]
		m.countDown = m.curItem.Target
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.curItem != nil && m.curItem.Type == config.WorkoutType_Count && !m.breaking {
				if m.curItem.Break == 0 {
					m.curRepeat++
					if m.curRepeat == m.curItem.Repeat {
						m.curItemIdx++
						m.curRepeat = 0
						if m.curItemIdx == len(m.wc.ItemList) {
							m.status = WorkoutStatus_Fin
							m.curItem = nil
						} else {
							m.breaking = false
							m.curItem = m.wc.ItemList[m.curItemIdx]
							m.countDown = m.curItem.Target
						}
					}
				} else {
					m.breaking = true
					m.countDown = m.curItem.Break
				}
			}
		case " ":
			switch m.status {
			case WorkoutStatus_Break:
				m.status = WorkoutStatus_Work
			case WorkoutStatus_Work:
				m.status = WorkoutStatus_Break
			case WorkoutStatus_Fin:
			}
		}
	case tickMsg:
		if m.status != WorkoutStatus_Work {
			return m, tick()
		}
		if m.curItem == nil || (m.curItem.Type == config.WorkoutType_Count && !m.breaking) {
			return m, tick()
		}
		m.countDown--
		if m.breaking {
			if m.countDown == 0 {
				m.curRepeat++
				m.breaking = false
				if m.curRepeat == m.curItem.Repeat {
					m.curRepeat = 0
					m.curItemIdx++
					if m.curItemIdx == len(m.wc.ItemList) {
						m.status = WorkoutStatus_Fin
						m.curItem = nil
					} else {
						m.curItem = m.wc.ItemList[m.curItemIdx]
						m.countDown = m.curItem.Target
					}
				} else {
					m.countDown = m.curItem.Target
				}
			}
		} else {
			if m.countDown == 0 {
				if m.curItem.Break == 0 {
					m.breaking = false
					m.curRepeat++
					if m.curRepeat == m.curItem.Repeat {
						m.curItemIdx++
						if m.curItemIdx == len(m.wc.ItemList) {
							m.status = WorkoutStatus_Fin
							m.curItem = nil
						} else {
							m.curItem = m.wc.ItemList[m.curItemIdx]
						}
					}
					m.countDown = m.curItem.Target
				} else {
					m.breaking = true
					m.countDown = m.curItem.Break
				}

			}
		}
		return m, tick()

	}

	return m, nil
}

func (m model) View() string {
	s := ""
	if m.status == WorkoutStatus_Fin {
		f := figure.NewFigure("FIN", "", true)
		s += finishStyle.Render(f.String()) + "\n\n"
	} else if m.curItem != nil {
		f := figure.NewFigure(fmt.Sprintf("%3d", m.countDown), "", true)
		s += todoStyle.Render(m.curItem.Name) + "\t"
		if m.breaking {
			s += "breaking\n"
			s += breakStyle.Render(f.String())
		} else {
			s += "working\n"
			s += workStyle.Render(f.String())
		}
		s += "\n\n"
	}
	s += m.itemListView() + "\n\n"
	return s
}

func (m *model) itemListView() string {
	var viewLineList []string
	completed := true
	for idx, item := range m.wc.ItemList {
		if idx == m.curItemIdx {
			completed = false
		}
		viewLineList = append(viewLineList, m.itemView(item, completed, m.curRepeat))
	}
	return strings.Join(viewLineList, "\n")

}

func (m *model) itemView(item *config.WorkoutItem, completed bool, repeate int) string {
	style := todoStyle
	if completed {
		repeate = item.Repeat
		style = finishStyle
	}
	return style.Render(fmt.Sprintf("%v: duartion: %ds break: %ds repeat: %d/%d",
		item.Name, item.Target, item.Break, repeate, item.Repeat))
}

func RunWithWorkoutConfig(wc *config.WorkoutConfig) {
	p := tea.NewProgram(initialModel(wc))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
