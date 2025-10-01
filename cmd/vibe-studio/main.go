package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listMinWidth = 26
	footerHeight = 3
)

type infusionPhase int

const (
	phaseIdle infusionPhase = iota
	phaseInfusing
	phaseComplete
)

type vibeItem struct {
	title       string
	description string
	hue         string
	mood        string
}

func (v vibeItem) Title() string       { return v.title }
func (v vibeItem) Description() string { return v.description }
func (v vibeItem) FilterValue() string { return v.title }

type infusionTickMsg struct{}

type keyMap struct {
	Infuse     key.Binding
	Shuffle    key.Binding
	ToggleLog  key.Binding
	ToggleHelp key.Binding
	Quit       key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Infuse:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "infuse blend")),
		Shuffle:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "shuffle muse")),
		ToggleLog:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "toggle focus")),
		ToggleHelp: key.NewBinding(key.WithKeys("?", "/"), key.WithHelp("?", "toggle help")),
		Quit:       key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Infuse, k.Shuffle, k.ToggleLog, k.ToggleHelp, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Infuse, k.Shuffle, k.ToggleLog},
		{k.ToggleHelp, k.Quit},
	}
}

type model struct {
	width     int
	height    int
	columnGap int

	list     list.Model
	viewport viewport.Model
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     keyMap

	phase       infusionPhase
	progressVal float64
	focusOnList bool
	blends      []string
}

var (
	listStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170")).Padding(1, 1)
	viewportStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("99")).Padding(1, 2)
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true).Padding(0, 1)
	subtitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("189"))
	statusReady    = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	statusInfusing = lipgloss.NewStyle().Foreground(lipgloss.Color("218")).Bold(true)
	statusComplete = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
)

func newModel() model {
	rand.Seed(time.Now().UnixNano())

	items := []list.Item{
		vibeItem{"Lumen Nectar", "Citrine bloom, honeyed brass, a vertical sunrise.", "Pulse", "Radiant"},
		vibeItem{"Glacial Prism", "Iridescent gliss, crushed mint, lunar reflections.", "Prism", "Cool"},
		vibeItem{"Velvet Ember", "Cardamom ash riding velvet bass and wildfire.", "Glow", "Warm"},
		vibeItem{"Signal Bloom", "Neon rains, modem birdsong, ultraviolet pollen.", "Flux", "Electric"},
		vibeItem{"Amber River", "Resonant reeds over amber dusk, tide-swept rhythm.", "Drift", "Calm"},
		vibeItem{"Azure Temple", "Chorused whales, cobalt incense, open sky reverb.", "Wave", "Deep"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().Foreground(lipgloss.Color("213")).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().Foreground(lipgloss.Color("182"))

	l := list.New(items, delegate, 30, 14)
	l.Title = "Muse Palette"
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

	prog := progress.New(progress.WithDefaultGradient())

	vp := viewport.New(60, 16)
	vp.SetContent(introCopy())

	keys := newKeyMap()

	return model{
		columnGap:   2,
		list:        l,
		viewport:    vp,
		spinner:     sp,
		progress:    prog,
		help:        help.New(),
		keys:        keys,
		phase:       phaseIdle,
		progressVal: 0,
		focusOnList: true,
	}
}

func introCopy() string {
	return "Welcome to Vibe Studio. Select a muse from the left, infuse it, and let the console sketch a synesthetic blend."
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tea.Tick(time.Millisecond*150, func(time.Time) tea.Msg { return infusionTickMsg{} }))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.list.SetSize(m.listWidth(), m.listHeight())
		m.viewport.Width = m.viewportWidth()
		m.viewport.Height = m.viewportHeight()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Shuffle):
			m.shuffle()
		case key.Matches(msg, m.keys.ToggleLog):
			m.focusOnList = !m.focusOnList
		case key.Matches(msg, m.keys.Infuse):
			if m.phase != phaseInfusing {
				cmds = append(cmds, m.beginInfusion(), infusionTickCmd(), m.spinner.Tick)
			}
		}

		if m.focusOnList {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			if m.phase != phaseInfusing {
				m.viewport.SetContent(m.describeCurrent())
			}
		} else {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case infusionTickMsg:
		if m.phase == phaseInfusing {
			m.progressVal += 0.07 + rand.Float64()*0.05
			if m.progressVal >= 1 {
				m.progressVal = 1
				m.phase = phaseComplete
				cmds = append(cmds, m.progress.SetPercent(1))
				blend := m.generateBlend()
				m.blends = append([]string{blend}, m.blends...)
				if len(m.blends) > 8 {
					m.blends = m.blends[:8]
				}
				m.viewport.SetContent(m.renderLog())
			} else {
				cmds = append(cmds, m.progress.SetPercent(m.progressVal), infusionTickCmd(), m.spinner.Tick)
			}
		}
		return m, tea.Batch(cmds...)

	default:
		if msg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		if !m.focusOnList {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.focusOnList && m.phase != phaseInfusing {
			m.viewport.SetContent(m.describeCurrent())
		}
		return m, tea.Batch(cmds...)
	}
}

func (m *model) beginInfusion() tea.Cmd {
	m.phase = phaseInfusing
	m.progressVal = 0
	return m.progress.SetPercent(0)
}

func (m *model) shuffle() {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}
	m.list.Select(rand.Intn(len(items)))
	if m.phase != phaseInfusing {
		m.viewport.SetContent(m.describeCurrent())
	}
}

func (m model) selectedVibe() vibeItem {
	item, ok := m.list.SelectedItem().(vibeItem)
	if !ok {
		return vibeItem{}
	}
	return item
}

func (m *model) updateLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	if m.viewport.Width == 0 {
		m.viewport.Width = m.viewportWidth()
	}
	if m.viewport.Height == 0 {
		m.viewport.Height = m.viewportHeight()
	}
	if len(m.blends) == 0 {
		m.viewport.SetContent(m.describeCurrent())
	}
}

func (m model) listWidth() int {
	w := m.width / 3
	if w < listMinWidth {
		w = listMinWidth
	}
	if w > m.width-20 {
		w = m.width - 20
	}
	if w < listMinWidth {
		w = listMinWidth
	}
	return w
}

func (m model) listHeight() int {
	return max(6, m.height-footerHeight-3)
}

func (m model) viewportWidth() int {
	return max(20, m.width-m.listWidth()-m.columnGap-4)
}

func (m model) viewportHeight() int {
	return max(6, m.height-footerHeight-3)
}

func (m *model) generateBlend() string {
	vibe := m.selectedVibe()
	textures := []string{"silk", "ember", "crystal", "pulse", "mist", "grain"}
	motions := []string{"spirals", "fractals", "drift lines", "pulse waves", "auroras", "migrations"}
	accents := []string{"echo guitar", "modular bloom", "holographic choir", "percussion dust", "analog haze", "quantum chime"}
	texture := textures[rand.Intn(len(textures))]
	motion := motions[rand.Intn(len(motions))]
	accent := accents[rand.Intn(len(accents))]

	stamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("[%s] %s // hue:%s mood:%s // texture:%s // motion:%s // accent:%s", stamp, vibe.title, vibe.hue, vibe.mood, texture, motion, accent)
}

func (m *model) renderLog() string {
	if len(m.blends) == 0 {
		return m.describeCurrent()
	}
	return strings.Join(m.blends, "\n")
}

func (m model) describeCurrent() string {
	vibe := m.selectedVibe()
	if vibe.title == "" {
		return introCopy()
	}
	summary := fmt.Sprintf("Current Muse: %s\nHue: %s\nMood: %s\n\n%s", vibe.title, vibe.hue, vibe.mood, vibe.description)
	return summary
}

func (m model) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		titleStyle.Render("Vibe Studio"),
		subtitleStyle.Render(" – generative blend atelier"),
	)

	status := m.renderStatus()

	listBoxStyle := listStyle.Copy()
	logBoxStyle := viewportStyle.Copy()
	if m.focusOnList {
		listBoxStyle = listBoxStyle.BorderForeground(lipgloss.Color("213"))
		logBoxStyle = logBoxStyle.BorderForeground(lipgloss.Color("59"))
	} else {
		listBoxStyle = listBoxStyle.BorderForeground(lipgloss.Color("59"))
		logBoxStyle = logBoxStyle.BorderForeground(lipgloss.Color("213")).Bold(true)
	}

	listView := listBoxStyle.Width(m.listWidth()).Height(m.listHeight() + 2).Render(m.list.View())
	logView := logBoxStyle.Width(m.viewportWidth()).Height(m.viewportHeight() + 2).Render(m.viewport.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, strings.Repeat(" ", m.columnGap), logView)

	footer := m.progress.View()
	helper := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		status,
		body,
		footer,
		helper,
	)
}

func (m model) renderStatus() string {
	var state string
	switch m.phase {
	case phaseInfusing:
		state = statusInfusing.Render(m.spinner.View() + " infusing...")
	case phaseComplete:
		state = statusComplete.Render("blend bottled")
	default:
		state = statusReady.Render("ready")
	}

	vibe := m.selectedVibe()
	info := infoStyle.Render(fmt.Sprintf(" muse: %s", vibe.title))

	return lipgloss.JoinHorizontal(lipgloss.Top, state, "  ", info)
}

func infusionTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg {
		return infusionTickMsg{}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
