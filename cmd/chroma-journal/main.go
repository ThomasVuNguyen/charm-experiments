package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spread struct {
	title   string
	mantra  string
	cards   []card
	footer  string
	palette palette
}

type card struct {
	heading string
	content string
}

type palette struct {
	background string
	accent     string
	washes     []string
	border     string
	shadow     string
}

type model struct {
	width     int
	height    int
	spreads   []spread
	index     int
	gridMode  bool
	glowPulse float64
}

var (
	docStyle = lipgloss.NewStyle().Align(lipgloss.Center).Padding(1, 0)
	frame    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

func newModel() model {
	rand.Seed(time.Now().UnixNano())
	spreads := []spread{
		{
			title:  "Neon Herbarium",
			mantra: "Catalog the light that grows between frequencies.",
			cards: []card{
				{"Synesthesia Blooms", "Drip phosphor onto sonic stems; map the smell of chords."},
				{"Chromatic Soil", "Layer VHS grain with kaleidoscopic mycelium for lo-fi texture."},
				{"Afterglow Ritual", "Steep pixels in tidepool gradients until dawn hums."},
			},
			footer: "Press space to toggle layout • ←/→ to change spread • r to reshuffle washes",
			palette: palette{
				background: "#120926",
				accent:     "#F7BAE8",
				washes:     []string{"#4F2EDB", "#BC4CF9", "#FF79F9", "#FFA8D9"},
				border:     "#7C3AED",
				shadow:     "#0B0212",
			},
		},
		{
			title:  "Sunprint Observatory",
			mantra: "Expose the paper to windborne echoes and catalog the silhouettes.",
			cards: []card{
				{"Solar Ink", "Blend citrus spectra with brass over a linen substrate."},
				{"Aurora Stitch", "Hand quilt the twilight with auric thread and compass hum."},
				{"Dust Glyphs", "Etch constellations into pollen motes floating through projector haze."},
			},
			footer: "Chroma tip: gradient headings borrow pigment from the current palette",
			palette: palette{
				background: "#0F2012",
				accent:     "#F1F79E",
				washes:     []string{"#3AA677", "#6ED682", "#F4F69E", "#FFCA7A"},
				border:     "#93E697",
				shadow:     "#07140B",
			},
		},
		{
			title:  "Signal Dream Log",
			mantra: "Transcribe the static that glows behind closed eyelids.",
			cards: []card{
				{"Carrier Wave", "Ride midnight FM into lucid sketches of forgotten signage."},
				{"Ghost Typo", "Let stray photons misprint the headline into poetic glitches."},
				{"Resonant Margin", "Highlight the silence between syllables with pearlescent ink."},
			},
			footer: "Press g to switch between gallery (grid) and column layouts",
			palette: palette{
				background: "#0A1824",
				accent:     "#9BD7FF",
				washes:     []string{"#1D64F2", "#3F8CFF", "#6AAFFF", "#B8E0FF"},
				border:     "#64A7FF",
				shadow:     "#04101B",
			},
		},
	}

	return model{
		spreads:  spreads,
		gridMode: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg { return pulseMsg{} })
}

type pulseMsg struct{}

type shuffleMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			m.index = (m.index - 1 + len(m.spreads)) % len(m.spreads)
		case "right", "l":
			m.index = (m.index + 1) % len(m.spreads)
		case "g", "G", "space":
			m.gridMode = !m.gridMode
		case "r":
			return m, shuffleCmd()
		}
		return m, nil
	case pulseMsg:
		m.glowPulse += 0.18
		return m, tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg { return pulseMsg{} })
	case shuffleMsg:
		m.shuffleWashes()
		return m, nil
	default:
		return m, nil
	}
}

func shuffleCmd() tea.Cmd {
	return func() tea.Msg {
		return shuffleMsg{}
	}
}

func (m *model) shuffleWashes() {
	current := &m.spreads[m.index]
	rand.Shuffle(len(current.palette.washes), func(i, j int) {
		current.palette.washes[i], current.palette.washes[j] = current.palette.washes[j], current.palette.washes[i]
	})
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "calibrating gradients..."
	}

	spread := m.spreads[m.index]
	bg := lipgloss.NewStyle().Background(lipgloss.Color(spread.palette.background)).Padding(1, 2)

	header := renderHeader(spread, m.glowPulse)
	body := m.renderCards(spread)
	footer := renderFooter(spread, m.index+1, len(m.spreads))

	content := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	return docStyle.Width(m.width).Render(bg.Width(m.width - 4).Render(content))
}

func (m model) renderCards(sp spread) string {
	cardStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(sp.palette.border)).Padding(1, 2).Margin(1, 1).Background(lipgloss.Color(sp.palette.shadow)).BorderBackground(lipgloss.Color(sp.palette.background))

	var rendered []string
	for i, c := range sp.cards {
		wash := sp.palette.washes[i%len(sp.palette.washes)]
		heading := gradientText(strings.ToUpper(c.heading), sp.palette.washes)
		body := lipgloss.NewStyle().Foreground(lipgloss.Color(sp.palette.accent)).Render(c.content)
		inner := lipgloss.JoinVertical(lipgloss.Left, heading, body)
		card := cardStyle.Background(lipgloss.Color(wash)).Render(inner)
		rendered = append(rendered, card)
	}

	if m.gridMode {
		// arrange cards in rows of two for wider canvases
		if m.width > 80 {
			rows := make([]string, 0)
			for i := 0; i < len(rendered); i += 2 {
				if i+1 < len(rendered) {
					rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rendered[i], rendered[i+1]))
				} else {
					rows = append(rows, rendered[i])
				}
			}
			return lipgloss.JoinVertical(lipgloss.Left, rows...)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

func renderHeader(sp spread, glow float64) string {
	title := gradientText(sp.title, sp.palette.washes)
	mantra := lipgloss.NewStyle().Foreground(lipgloss.Color(sp.palette.accent)).Italic(true).Render(sp.mantra)

	pulses := []string{"◐", "◓", "◑", "◒"}
	glyph := pulses[int(glow)%len(pulses)]
	halo := lipgloss.NewStyle().Foreground(lipgloss.Color(sp.palette.washes[int(glow)%len(sp.palette.washes)])).Render(glyph)

	header := lipgloss.JoinHorizontal(lipgloss.Bottom, halo, " ", title)
	return lipgloss.JoinVertical(lipgloss.Left, header, mantra)
}

func renderFooter(sp spread, index, total int) string {
	status := fmt.Sprintf("spread %d/%d", index, total)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(sp.palette.accent)).Bold(true)
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color(sp.palette.washes[len(sp.palette.washes)-1])).Render(sp.footer)
	bar := lipgloss.JoinHorizontal(lipgloss.Center, statusStyle.Render(status), "  ", footer)
	return lipgloss.NewStyle().MarginTop(1).Render(bar)
}

func gradientText(text string, colors []string) string {
	if len(colors) == 0 {
		return text
	}
	runes := []rune(text)
	segments := make([]string, len(runes))
	for i, r := range runes {
		color := colors[i%len(colors)]
		segments[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(r))
	}
	return strings.Join(segments, "")
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("error:", err)
	}
}
