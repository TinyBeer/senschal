package terminal

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"fmt"
	"os"
	"seneschal/config"
	"seneschal/ui/component"
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
	selectedWorkConfigStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIYellow)).Background(lipgloss.ANSIColor(termenv.ANSIBrightBlue))
	todoStyle               = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIYellow))
	finishStyle             = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIGreen))
	breakStyle              = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightGreen))
	workStyle               = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIMagenta))
)

type tickMsg time.Time

//go:generate stringer -type=WorkoutStatus -linecomment
type WorkoutStatus int

const (
	WorkoutStatus_Inactive WorkoutStatus = iota // INACTIVE
	WorkoutStatus_Active                        // ACTIVE
	WorkoutStatus_Fin                           // FIN
)

type model struct {
	wcList     []*config.WorkoutConfig
	wc         *config.WorkoutConfig
	status     WorkoutStatus
	curItem    *config.WorkoutItem
	curItemIdx int
	breaking   bool
	countDown  int
	curRepeat  int
	frame      int
}

func initialModel(wcList []*config.WorkoutConfig, wc *config.WorkoutConfig) model {
	m := model{
		wcList:     wcList,
		wc:         wc,
		status:     WorkoutStatus_Inactive,
		curItem:    nil,
		curItemIdx: 0,
		breaking:   false,
		countDown:  0,
		curRepeat:  0,
	}
	if wc == nil || len(wc.ItemList) == 0 {
		m.status = WorkoutStatus_Fin
	} else {
		m.curItem = m.wc.ItemList[m.curItemIdx]
		if m.curItem.Type == config.WorkoutType_Duration {
			m.countDown = m.curItem.Target
		} else {
			m.status = WorkoutStatus_Active
		}

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
			case WorkoutStatus_Inactive:
				m.status = WorkoutStatus_Active
			case WorkoutStatus_Active:
				m.status = WorkoutStatus_Inactive
			case WorkoutStatus_Fin:
			}
		}
	case tickMsg:
		m.frame++
		if m.status != WorkoutStatus_Active {
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
						if m.curItem.Type == config.WorkoutType_Duration {
							m.countDown = m.curItem.Target
						}
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
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			if m.status == WorkoutStatus_Active && m.curItem != nil && m.curItem.Type == config.WorkoutType_Count && !m.breaking {
				m.countDown++
				if m.countDown == m.curItem.Target {
					if m.curItem.Break == 0 {
						m.curRepeat++
						m.countDown = 0
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
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := ""
	if m.status == WorkoutStatus_Fin {
		f := figure.NewFigure(m.status.String(), "", true)
		s += finishStyle.Render(f.String()) + "\n\n"
	} else if m.curItem != nil {
		s += component.StyleActiveItem.Render(m.curItem.Name) + "\t"
		if m.breaking {
			f := figure.NewFigure(fmt.Sprintf("%3d -- %3d", m.curItem.Break, m.countDown), "", true)
			s += breakStyle.Render(m.status.String() + "\n" + f.String())
		} else {
			f := figure.NewFigure(fmt.Sprintf("%3d -- %3d", m.curItem.Target, m.countDown), "", true)
			s += workStyle.Render(m.status.String() + "\n" + f.String())
		}
		s += "\n\n"
	}

	return s + m.workoutListInfo()
}

func (m model) workoutListInfo() string {
	workoutContainer := component.NewBox(component.Direction_V)
	for _, wc := range m.wcList {
		if wc == m.wc {
			workoutContainer.AddSub(component.NewInlineTextWithStyle(10, "->"+wc.Name, component.StyleSelectedWorkConfig))
		} else {
			workoutContainer.AddSub(component.NewInlineText(10, wc.Name))
		}

	}

	itemContainre := component.NewBox(component.Direction_V)
	if m.wc != nil {
		for idx, item := range m.wc.ItemList {
			itemInfo := component.NewBox(component.Direction_H)
			var style lipgloss.Style
			var repeatInfo string
			switch {
			case m.curItemIdx > idx:
				style = component.StyleCompletedItem
				repeatInfo = fmt.Sprintf("%d/%d", item.Repeat, item.Repeat)
			case m.curItemIdx == idx:
				style = component.StyleActiveItem
				repeatInfo = fmt.Sprintf("%d/%d", m.curRepeat, item.Repeat)
			default:
				repeatInfo = fmt.Sprintf("%d/%d", 0, item.Repeat)
			}

			var targetInfo string
			switch item.Type {
			case config.WorkoutType_Count:
				targetInfo = fmt.Sprintf("%s: %d", item.Type.String(), item.Target)
			case config.WorkoutType_Duration:
				targetInfo = fmt.Sprintf("%s: %ds", item.Type.String(), item.Target)
			}

			itemInfo.AddSub(component.NewInlineTextWithStyle(20, item.Name, style))
			itemInfo.AddSub(component.NewInlineTextWithStyle(16, targetInfo, style))
			itemInfo.AddSub(component.NewInlineTextWithStyle(16, "Repeat: "+repeatInfo, style))
			itemContainre.AddSub(itemInfo)
		}

	}

	v := component.NewBox(component.Direction_H)
	v.AddSub(workoutContainer)
	v.AddSub(itemContainre)

	return strings.Join(component.JoinStyleStringMatrix(v.GetCurrentContent(m.frame)), "\n")
}

func Workout(wcList []*config.WorkoutConfig, wc *config.WorkoutConfig) {
	fmt.Print("\033[H\033[2J")
	p := tea.NewProgram(initialModel(wcList, wc), tea.WithMouseAllMotion())
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
