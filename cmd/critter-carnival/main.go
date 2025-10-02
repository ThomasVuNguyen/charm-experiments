package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fps            = 32
	deltaTime      = 1.0 / fps
	statusHeight   = 6
	minStageHeight = 16
)

type spriteFrame []string

type pixelPalette map[rune]string

type styleKey struct {
	fg   string
	bg   string
	bold bool
}

type cell struct {
	ch   rune
	fg   string
	bg   string
	bold bool
}

type model struct {
	width       int
	height      int
	ready       bool
	rng         *rand.Rand
	t           float64
	sprite      []spriteFrame
	palette     pixelPalette
	frame       int
	frameTimer  float64
	frameSpeed  float64
	hoverRadius float64
	hoverSpeed  float64
	colorPulse  float64

	styleCache map[styleKey]lipgloss.Style
	backdrop   background
}

type background struct {
	skyPalette    []string
	groundPalette []string
	overlayColors []string
	horizon       float64
}

type frameMsg time.Time

func main() {
	rand.Seed(time.Now().UnixNano())
	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func newModel() model {
	return model{
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		sprite:      foxSpriteFrames,
		palette:     foxPalette,
		frameSpeed:  0.15,
		hoverRadius: 6,
		hoverSpeed:  0.45,
		colorPulse:  0.35,
		styleCache:  make(map[styleKey]lipgloss.Style),
		backdrop: background{
			skyPalette:    []string{"#040726", "#101d46", "#283a7a", "#4c5bbb"},
			groundPalette: []string{"#0c1f1d", "#123530", "#1c4f46", "#2a6f62"},
			overlayColors: []string{"#94f7d1", "#5ce1ff", "#c084fc"},
			horizon:       0.58,
		},
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 0 && m.width > 0 {
			m.ready = true
		}
		return m, nil
	case frameMsg:
		if !m.ready {
			return m, tick()
		}
		m.t += deltaTime
		m.frameTimer += deltaTime
		if len(m.sprite) > 0 && m.frameSpeed > 0 && m.frameTimer >= m.frameSpeed {
			m.frameTimer = math.Mod(m.frameTimer, m.frameSpeed)
			m.frame = (m.frame + 1) % len(m.sprite)
		}
		return m, tick()
	default:
		return m, nil
	}
}

func (m model) View() string {
	if !m.ready {
		return "stitching constellations..."
	}

	stageWidth, stageHeight := m.stageSize()
	canvas := newCanvas(stageWidth, stageHeight)

	m.drawBackdrop(canvas)
	m.drawAurora(canvas)
	m.drawFireflies(canvas)
	m.drawSprite(canvas)

	stage := renderCanvas(canvas, m.styleCache)

	info := renderStatus(m.width, foxPalette, m.t)

	var b strings.Builder
	b.WriteString(stage)
	b.WriteByte('\n')
	b.WriteString(info)
	return b.String()
}

func (m model) stageSize() (int, int) {
	stageHeight := m.height - statusHeight
	if stageHeight < minStageHeight {
		stageHeight = max(4, m.height-2)
	}
	return m.width, stageHeight
}

func (m model) drawSprite(canvas [][]cell) {
	if len(m.sprite) == 0 {
		return
	}
	frame := m.sprite[m.frame%len(m.sprite)]
	stageHeight := len(canvas)
	if stageHeight == 0 {
		return
	}
	stageWidth := len(canvas[0])

	centerX := float64(stageWidth) * 0.5
	centerY := float64(stageHeight) * 0.44

	orbit := m.hoverRadius
	if orbit > float64(stageWidth)/3 {
		orbit = float64(stageWidth) / 3
	}

	x := centerX + math.Cos(m.t*m.hoverSpeed)*orbit
	y := centerY + math.Sin(m.t*m.hoverSpeed*0.72)*orbit*0.55

	brightness := 0.5 + 0.5*math.Sin(m.t*m.colorPulse)
	tint := blendHex("#f472b6", "#94f7d1", brightness)

	paintFrame(canvas, int(math.Round(x))-len(frame[0])/2, int(math.Round(y))-len(frame)/2, frame, tint, m.palette)
}

func (m model) drawBackdrop(canvas [][]cell) {
	if len(canvas) == 0 {
		return
	}
	stageHeight := len(canvas)
	stageWidth := len(canvas[0])
	horizon := clampFloat(m.backdrop.horizon, 0.2, 0.9)
	horizonRow := clampInt(int(float64(stageHeight)*horizon), 1, stageHeight-2)

	for y := 0; y < stageHeight; y++ {
		var palette []string
		var t float64
		if y <= horizonRow {
			palette = m.backdrop.skyPalette
			t = float64(y) / float64(max(1, horizonRow))
		} else {
			palette = m.backdrop.groundPalette
			denom := stageHeight - horizonRow
			if denom <= 1 {
				t = 0
			} else {
				t = float64(y-horizonRow) / float64(denom-1)
			}
		}
		color := gradientColor(palette, t)
		for x := 0; x < stageWidth; x++ {
			canvas[y][x].bg = color
			if canvas[y][x].ch == 0 {
				canvas[y][x].ch = ' '
			}
		}
	}

	accent := blendHex("#94f7d1", "#38bdf8", 0.4)
	row := canvas[horizonRow]
	for x := 0; x < stageWidth; x += 2 {
		if row[x].ch == ' ' {
			row[x].ch = '_'
			row[x].fg = accent
		}
	}
}

func (m model) drawAurora(canvas [][]cell) {
	if len(canvas) == 0 || len(canvas[0]) == 0 {
		return
	}
	stageHeight := len(canvas)
	stageWidth := len(canvas[0])
	base := clampInt(int(float64(stageHeight)*m.backdrop.horizon), 2, stageHeight-3)

	for i, color := range m.backdrop.overlayColors {
		amplitude := float64(stageHeight) * (0.04 + 0.025*float64(i))
		wavelength := 10 + i*6
		thickness := 1 + i
		speed := 0.7 + 0.25*float64(i)
		for x := 0; x < stageWidth; x++ {
			if (x+i)%2 != 0 {
				continue
			}
			wave := math.Sin((float64(x)/float64(wavelength))*2*math.Pi + m.t*speed + float64(i))
			centerY := base - i*2 + int(math.Round(wave*amplitude))
			for t := -thickness; t <= thickness; t++ {
				y := clampInt(centerY+t, 0, stageHeight-1)
				cell := canvas[y][x]
				if cell.ch != ' ' {
					continue
				}
				cell.ch = '~'
				cell.fg = blendHex(color, "#ffffff", 0.15*float64(thickness-t+1))
				if i == 0 {
					cell.bold = true
				}
				canvas[y][x] = cell
			}
		}
	}
}

func (m model) drawFireflies(canvas [][]cell) {
	if len(canvas) == 0 {
		return
	}
	stageHeight := len(canvas)
	stageWidth := len(canvas[0])
	density := stageWidth / 18
	if density < 8 {
		density = 8
	}
	for i := 0; i < density; i++ {
		phase := float64(i) / float64(density)
		x := int(float64(stageWidth) * phase)
		y := int(float64(stageHeight)*0.2 + math.Sin(m.t*0.6+phase*math.Pi*2)*3)
		if y < 0 || y >= stageHeight {
			continue
		}
		glow := 0.5 + 0.5*math.Sin(m.t*3+phase*6)
		color := blendHex("#fef3c7", "#a855f7", glow)
		canvas[y][x] = cell{ch: '•', fg: color, bg: canvas[y][x].bg}
	}
}

func renderStatus(width int, palette pixelPalette, t float64) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	glow := 0.5 + 0.5*math.Sin(t*0.9)
	accent := blendHex(palette['5'], palette['2'], glow)
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(accent))

	lines := []string{
		titleStyle.Render("Celestial Familiar"),
		accentStyle.Render("A single fox spirits through aurora lullabies"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Render("Use Ctrl+C or q to leave the dream"),
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func newCanvas(width, height int) [][]cell {
	if width <= 0 || height <= 0 {
		return [][]cell{}
	}
	canvas := make([][]cell, height)
	for y := range canvas {
		row := make([]cell, width)
		for x := range row {
			row[x] = cell{ch: ' '}
		}
		canvas[y] = row
	}
	return canvas
}

func paintFrame(canvas [][]cell, startX, startY int, frame spriteFrame, fallback string, palette pixelPalette) {
	height := len(frame)
	if height == 0 {
		return
	}
	stageHeight := len(canvas)
	if stageHeight == 0 {
		return
	}
	stageWidth := len(canvas[0])
	for dy, line := range frame {
		y := startY + dy
		if y < 0 || y >= stageHeight {
			continue
		}
		for dx, r := range line {
			if r == ' ' {
				continue
			}
			x := startX + dx
			if x < 0 || x >= stageWidth {
				continue
			}
			existing := canvas[y][x]
			if palette != nil {
				if col, ok := palette[r]; ok {
					existing.bg = col
					existing.fg = ""
					existing.bold = false
					existing.ch = ' '
					canvas[y][x] = existing
					continue
				}
			}
			existing.ch = r
			existing.fg = fallback
			existing.bold = false
			canvas[y][x] = existing
		}
	}
}

func renderCanvas(canvas [][]cell, cache map[styleKey]lipgloss.Style) string {
	if len(canvas) == 0 {
		return ""
	}
	var b strings.Builder
	for y, row := range canvas {
		if len(row) == 0 {
			if y < len(canvas)-1 {
				b.WriteByte('\n')
			}
			continue
		}
		currentFg := row[0].fg
		currentBg := row[0].bg
		currentBold := row[0].bold
		var segment strings.Builder
		for _, c := range row {
			if c.fg != currentFg || c.bg != currentBg || c.bold != currentBold {
				b.WriteString(applyStyle(segment.String(), currentFg, currentBg, currentBold, cache))
				segment.Reset()
				currentFg = c.fg
				currentBg = c.bg
				currentBold = c.bold
			}
			if c.ch == 0 {
				segment.WriteByte(' ')
			} else {
				segment.WriteRune(c.ch)
			}
		}
		b.WriteString(applyStyle(segment.String(), currentFg, currentBg, currentBold, cache))
		if y < len(canvas)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func applyStyle(text string, fg string, bg string, bold bool, cache map[styleKey]lipgloss.Style) string {
	if text == "" {
		return ""
	}
	key := styleKey{fg: fg, bg: bg, bold: bold}
	style, ok := cache[key]
	if !ok {
		style = lipgloss.NewStyle()
		if fg != "" {
			style = style.Foreground(lipgloss.Color(fg))
		}
		if bg != "" {
			style = style.Background(lipgloss.Color(bg))
		}
		if bold {
			style = style.Bold(true)
		}
		cache[key] = style
	}
	return style.Render(text)
}

func gradientColor(palette []string, t float64) string {
	if len(palette) == 0 {
		return ""
	}
	if len(palette) == 1 {
		return palette[0]
	}
	t = clampFloat(t, 0, 0.9999)
	scaled := t * float64(len(palette)-1)
	idx := int(scaled)
	frac := scaled - float64(idx)
	nextIdx := idx + 1
	if nextIdx >= len(palette) {
		nextIdx = len(palette) - 1
	}
	return blendHex(palette[idx], palette[nextIdx], frac)
}

func blendHex(a, b string, t float64) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	r1, g1, b1 := hexToRGB(a)
	r2, g2, b2 := hexToRGB(b)
	t = clampFloat(t, 0, 1)
	r := int(math.Round(float64(r1) + (float64(r2)-float64(r1))*t))
	g := int(math.Round(float64(g1) + (float64(g2)-float64(g1))*t))
	bl := int(math.Round(float64(b1) + (float64(b2)-float64(b1))*t))
	return rgbToHex(r, g, bl)
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 255, 255, 255
	}
	r := int((val >> 16) & 0xFF)
	g := int((val >> 8) & 0xFF)
	b := int(val & 0xFF)
	return r, g, b
}

func rgbToHex(r, g, b int) string {
	r = clampInt(r, 0, 255)
	g = clampInt(g, 0, 255)
	b = clampInt(b, 0, 255)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	foxPalette = pixelPalette{
		'1': "#f97316",
		'2': "#fb923c",
		'3': "#1f2937",
		'4': "#fef3c7",
		'5': "#f472b6",
	}

	foxSpriteFrames = []spriteFrame{
		{
			"    222222    ",
			"  222222222   ",
			" 22221111222  ",
			"2221333313222 ",
			" 22134444122  ",
			"  22111152    ",
			"    21152     ",
		},
		{
			"    222222    ",
			"  222222222   ",
			" 22221111222  ",
			"2221333313222 ",
			" 22134444122  ",
			"   2211152    ",
			"    211552    ",
		},
	}
)
