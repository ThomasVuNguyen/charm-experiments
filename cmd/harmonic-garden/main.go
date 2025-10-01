package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

const (
	fps              = 60
	deltaTime        = 1.0 / fps
	maxTrail         = 42
	maxFollowers     = 30
	initialFollowers = 9
	minDamping       = 0.02
	maxDamping       = 3.2
	minFrequency     = 1.0
	maxFrequency     = 14.0
	statusLines      = 6
)

var (
	frameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	infoTitle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	infoValue    = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Padding(0, 1)
	bannerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	helpBoxStyle = lipgloss.NewStyle().Padding(1, 2).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("54"))
)

type autopScene int

const (
	sceneOrbit autopScene = iota
	sceneRose
	sceneCascade
	scenePulse
	sceneWander
)

type formationMode int

const (
	formationHalo formationMode = iota
	formationRibbon
	formationBloom
	formationHelix
)

type vector struct {
	x float64
	y float64
}

type cell struct {
	ch       rune
	fg, bg   string
	bold     bool
	priority int
}

type shaderFunc func(x, y, t float64, width, height int, theme moodTheme) (glyph rune, fg, bg string)

type keyMap struct {
	Quit            key.Binding
	ToggleMode      key.Binding
	CycleScene      key.Binding
	CycleFormation  key.Binding
	CycleMood       key.Binding
	AddFollower     key.Binding
	RemoveFollower  key.Binding
	IncreaseFreq    key.Binding
	DecreaseFreq    key.Binding
	IncreaseDamping key.Binding
	DecreaseDamping key.Binding
	MoveNorth       key.Binding
	MoveSouth       key.Binding
	MoveWest        key.Binding
	MoveEast        key.Binding
	ToggleHelp      key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Quit:            key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
		ToggleMode:      key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "auto/manual")),
		CycleScene:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next scene")),
		CycleFormation:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "next formation")),
		CycleMood:       key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "next mood")),
		AddFollower:     key.NewBinding(key.WithKeys("+", "=", "a"), key.WithHelp("+", "add muse")),
		RemoveFollower:  key.NewBinding(key.WithKeys("-", "_", "d"), key.WithHelp("-", "trim muse")),
		IncreaseFreq:    key.NewBinding(key.WithKeys("'", "]"), key.WithHelp("'", "freq +")),
		DecreaseFreq:    key.NewBinding(key.WithKeys(";", "["), key.WithHelp(";", "freq -")),
		IncreaseDamping: key.NewBinding(key.WithKeys(".", ">"), key.WithHelp(".", "damping +")),
		DecreaseDamping: key.NewBinding(key.WithKeys(",", "<"), key.WithHelp(",", "damping -")),
		MoveNorth:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "drift north")),
		MoveSouth:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "drift south")),
		MoveWest:        key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "drift west")),
		MoveEast:        key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "drift east")),
		ToggleHelp:      key.NewBinding(key.WithKeys("?", "/"), key.WithHelp("?", "toggle help")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ToggleMode, k.CycleScene, k.CycleFormation, k.CycleMood, k.AddFollower, k.RemoveFollower, k.ToggleHelp}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ToggleMode, k.CycleScene, k.CycleFormation, k.CycleMood},
		{k.IncreaseFreq, k.DecreaseFreq, k.IncreaseDamping, k.DecreaseDamping},
		{k.MoveNorth, k.MoveSouth, k.MoveWest, k.MoveEast},
		{k.AddFollower, k.RemoveFollower, k.ToggleHelp, k.Quit},
	}
}

type moodTheme struct {
	name         string
	description  string
	palette      []string
	background   string
	accent       string
	wispGlyphs   []rune
	trailGlyphs  []rune
	seedGlyph    rune
	seedInterval float64
	shader       shaderFunc
}

func (m moodTheme) colorAt(t float64) string {
	if len(m.palette) == 0 {
		return "#FFFFFF"
	}
	if len(m.palette) == 1 {
		return m.palette[0]
	}
	t = clamp(t, 0, 0.9999)
	scaled := t * float64(len(m.palette)-1)
	idx := int(scaled)
	frac := scaled - float64(idx)
	return blendHex(m.palette[idx], m.palette[idx+1], frac)
}

type sceneMeta struct {
	id          autopScene
	name        string
	description string
}

type formationMeta struct {
	id          formationMode
	name        string
	description string
}

type follower struct {
	order       int
	pos         vector
	vel         vector
	phase       float64
	speed       float64
	radius      float64
	offsetSeed  float64
	paletteSeed float64
	trace       []vector
	springX     harmonica.Spring
	springY     harmonica.Spring
}

type seed struct {
	projector *harmonica.Projectile
	pos       vector
	life      float64
	ttl       float64
	color     string
	glyph     rune
}

type model struct {
	width        int
	height       int
	canvasWidth  int
	canvasHeight int

	ready      bool
	autop      bool
	sceneIndex int
	formation  formationMode
	moodIndex  int

	freq    float64
	damping float64

	t         float64
	target    vector
	followers []*follower
	seeds     []*seed
	seedTimer float64
	rng       *rand.Rand

	keys       keyMap
	help       help.Model
	showHelp   bool
	styleCache map[string]lipgloss.Style
}

var (
	moods = []moodTheme{
		{
			name:         "Aurora Bloom",
			description:  "Iridescent dusk fields and electric petals",
			palette:      []string{"#3E1F65", "#5C3C99", "#8D73FF", "#FF8BD5", "#FFE8A3"},
			background:   "#0B0618",
			accent:       "#FFD8FD",
			wispGlyphs:   []rune{' ', ' ', '.', '`', '^'},
			trailGlyphs:  []rune{'.', '*', '+', 'o'},
			seedGlyph:    '*',
			seedInterval: 0.28,
		},
		{
			name:         "Cosmic Tie-Dye",
			description:  "Sunburst ripples and peace-wave whorls",
			palette:      []string{"#321040", "#7E1978", "#F54BA1", "#FFB94F", "#FFEFA9"},
			background:   "#150713",
			accent:       "#FFEFD2",
			wispGlyphs:   []rune{'~', '-', '.', '=', '*'},
			trailGlyphs:  []rune{'.', '*', 'o', '+'},
			seedGlyph:    '~',
			seedInterval: 0.26,
			shader:       tieDyeShader,
		},
		{
			name:         "Solar Garden",
			description:  "Heat shimmer blooms and molten ribbons",
			palette:      []string{"#251605", "#813D0B", "#D66B02", "#FFAF45", "#F9F871"},
			background:   "#120701",
			accent:       "#FFE9B0",
			wispGlyphs:   []rune{' ', '.', ',', '`', '"'},
			trailGlyphs:  []rune{'.', '+', '*', 'x'},
			seedGlyph:    '+',
			seedInterval: 0.35,
		},
		{
			name:         "Deep Current",
			description:  "Bioluminescent swirls in tidal night",
			palette:      []string{"#010D1B", "#014F86", "#0DA5C0", "#7EF2FF", "#F8FFF6"},
			background:   "#000407",
			accent:       "#B4F1FF",
			wispGlyphs:   []rune{' ', '.', '`', '~'},
			trailGlyphs:  []rune{'.', ':', '*', 'o'},
			seedGlyph:    '*',
			seedInterval: 0.24,
		},
	}

	scenes = []sceneMeta{
		{sceneOrbit, "Ellipse Drift", "Nested ellipses breathing in slow counterpoint"},
		{sceneRose, "Rose Bloom", "Five-petal harmonics unfurling and collapsing"},
		{sceneCascade, "Cascade", "Falling waterfall of envelopes and echoes"},
		{scenePulse, "Pulse Spiral", "Heartbeat spiral with luminous bursts"},
		{sceneWander, "Wander Field", "Noise-driven drift through latent space"},
	}

	formations = []formationMeta{
		{formationHalo, "Halo", "Radial orbits with delicate offsets"},
		{formationRibbon, "Ribbon", "Flowing comet tails weaving in stereo"},
		{formationBloom, "Bloom", "Petal clusters breathing with the beat"},
		{formationHelix, "Helix", "Twisted lattice rippling through depth"},
	}
)

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
	keys := newKeyMap()
	return model{
		freq:       7.2,
		damping:    0.22,
		autop:      true,
		sceneIndex: 0,
		formation:  formationHalo,
		moodIndex:  0,
		keys:       keys,
		help:       help.New(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		styleCache: make(map[string]lipgloss.Style),
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

type frameMsg time.Time

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recomputeCanvas()
		if !m.ready && m.canvasWidth > 0 && m.canvasHeight > 0 {
			m.target = vector{float64(m.canvasWidth) / 2, float64(m.canvasHeight) / 2}
			for len(m.followers) < initialFollowers {
				m.addFollower()
			}
			m.ready = true
		}
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	case frameMsg:
		if !m.ready {
			return m, tick()
		}
		m.t += deltaTime
		if m.autop {
			m.updateTarget()
		}
		m.updateFollowers()
		m.updateSeeds()
		return m, tick()
	default:
		return m, nil
	}
}

func (m model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.ToggleMode):
		m.autop = !m.autop
	case key.Matches(msg, m.keys.CycleScene):
		m.sceneIndex = (m.sceneIndex + 1) % len(scenes)
	case key.Matches(msg, m.keys.CycleFormation):
		idx := (indexOfFormation(m.formation) + 1) % len(formations)
		m.formation = formations[idx].id
	case key.Matches(msg, m.keys.CycleMood):
		m.moodIndex = (m.moodIndex + 1) % len(moods)
	case key.Matches(msg, m.keys.AddFollower):
		m.addFollower()
	case key.Matches(msg, m.keys.RemoveFollower):
		m.removeFollower()
	case key.Matches(msg, m.keys.IncreaseFreq):
		m.adjustFrequency(0.35)
	case key.Matches(msg, m.keys.DecreaseFreq):
		m.adjustFrequency(-0.35)
	case key.Matches(msg, m.keys.IncreaseDamping):
		m.adjustDamping(0.05)
	case key.Matches(msg, m.keys.DecreaseDamping):
		m.adjustDamping(-0.05)
	case key.Matches(msg, m.keys.MoveNorth):
		m.manualTarget(0, -1)
	case key.Matches(msg, m.keys.MoveSouth):
		m.manualTarget(0, 1)
	case key.Matches(msg, m.keys.MoveWest):
		m.manualTarget(-1, 0)
	case key.Matches(msg, m.keys.MoveEast):
		m.manualTarget(1, 0)
	case key.Matches(msg, m.keys.ToggleHelp):
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func indexOfFormation(f formationMode) int {
	for i, meta := range formations {
		if meta.id == f {
			return i
		}
	}
	return 0
}

func (m *model) manualTarget(dx, dy float64) {
	m.autop = false
	m.target.x += dx
	m.target.y += dy
	m.clampTarget()
}

func (m *model) recomputeCanvas() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	m.canvasWidth = m.width
	m.canvasHeight = m.height - statusLines
	if m.canvasHeight < 10 {
		m.canvasHeight = max(m.height-2, 3)
	}
	m.clampTarget()
}

func (m *model) clampTarget() {
	if m.canvasWidth <= 0 || m.canvasHeight <= 0 {
		return
	}
	if m.target.x < 0 {
		m.target.x = 0
	}
	if m.target.x > float64(m.canvasWidth-1) {
		m.target.x = float64(m.canvasWidth - 1)
	}
	if m.target.y < 0 {
		m.target.y = 0
	}
	if m.target.y > float64(m.canvasHeight-1) {
		m.target.y = float64(m.canvasHeight - 1)
	}
}

func (m *model) currentMood() moodTheme {
	return moods[m.moodIndex]
}

func (m *model) updateTarget() {
	if m.canvasWidth == 0 || m.canvasHeight == 0 {
		return
	}
	scene := scenes[m.sceneIndex].id
	w := float64(m.canvasWidth)
	h := float64(m.canvasHeight)
	cx := w / 2
	cy := h / 2

	switch scene {
	case sceneOrbit:
		a := w * 0.35
		b := h * 0.28
		speed := 0.55
		m.target.x = cx + math.Cos(m.t*speed)*a + math.Cos(m.t*0.9)*w*0.05
		m.target.y = cy + math.Sin(m.t*speed*1.2)*b + math.Sin(m.t*0.77)*h*0.04
	case sceneRose:
		k := 5.0
		theta := m.t * 0.8
		radius := (0.4 + 0.15*math.Sin(m.t*0.6)) * math.Sin(k*theta)
		r := radius * w
		m.target.x = cx + r*math.Cos(theta)
		m.target.y = cy + r*math.Sin(theta)
	case sceneCascade:
		slow := math.Sin(m.t * 0.3)
		sway := math.Sin(m.t * 1.8)
		drift := math.Sin(m.t*0.5 + sway*0.4)
		m.target.x = cx + drift*w*0.25
		m.target.y = cy + ((1+slow)/2)*h*0.35 + math.Sin(m.t*1.2)*h*0.06
	case scenePulse:
		theta := m.t * 1.3
		pulse := (math.Sin(m.t*2.4) + 1) / 2
		radius := w * (0.18 + 0.28*pulse)
		m.target.x = cx + radius*math.Cos(theta)
		m.target.y = cy + radius*0.7*math.Sin(theta*1.4)
	case sceneWander:
		n1 := perlin2(m.t*0.15, 0.0)
		n2 := perlin2(0.0, m.t*0.12+3.7)
		m.target.x = cx + n1*w*0.4
		m.target.y = cy + n2*h*0.35
	}
	m.clampTarget()
}

func (m *model) updateFollowers() {
	mood := m.currentMood()
	count := len(m.followers)
	if count == 0 {
		return
	}
	stageW := float64(m.canvasWidth)
	stageH := float64(m.canvasHeight)
	for _, f := range m.followers {
		f.step(m.target, m.formation, stageW, stageH, m.t, deltaTime, count, mood)
	}
}

func (m *model) updateSeeds() {
	mood := m.currentMood()
	stageW := float64(m.canvasWidth)
	stageH := float64(m.canvasHeight)
	if stageW == 0 || stageH == 0 {
		return
	}
	m.seedTimer += deltaTime
	if m.seedTimer >= mood.seedInterval {
		m.emitSeed(mood)
		m.seedTimer = math.Mod(m.seedTimer, mood.seedInterval)
	}

	alive := m.seeds[:0]
	for _, s := range m.seeds {
		pos := s.projector.Update()
		s.pos = vector{pos.X, pos.Y}
		s.life += deltaTime
		if s.life >= s.ttl {
			continue
		}
		if s.pos.x < -2 || s.pos.y < -2 || s.pos.x > stageW+2 || s.pos.y > stageH+2 {
			continue
		}
		alive = append(alive, s)
	}
	m.seeds = alive
}

func (m *model) emitSeed(theme moodTheme) {
	if m.canvasWidth == 0 || m.canvasHeight == 0 {
		return
	}
	velocity := harmonica.Vector{
		X: (m.rng.Float64()*2 - 1) * 14,
		Y: -6 - m.rng.Float64()*6,
	}
	start := harmonica.Point{X: m.target.x, Y: m.target.y}
	projectile := harmonica.NewProjectile(harmonica.FPS(fps), start, velocity, harmonica.Vector{X: 0, Y: 18})
	ttl := 1.4 + m.rng.Float64()*0.9
	hue := theme.colorAt(m.rng.Float64())
	m.seeds = append(m.seeds, &seed{
		projector: projectile,
		ttl:       ttl,
		glyph:     theme.seedGlyph,
		color:     hue,
	})
}

func (m *model) adjustDamping(delta float64) {
	m.damping = clamp(m.damping+delta, minDamping, maxDamping)
	m.retuneFollowers()
}

func (m *model) adjustFrequency(delta float64) {
	m.freq = clamp(m.freq+delta, minFrequency, maxFrequency)
	m.retuneFollowers()
}

func (m *model) addFollower() {
	if len(m.followers) >= maxFollowers {
		return
	}
	follower := newFollower(len(m.followers), m.freq, m.damping, m.rng)
	follower.pos = m.target
	follower.trace = append(follower.trace, m.target)
	m.followers = append(m.followers, follower)
}

func (m *model) removeFollower() {
	if len(m.followers) <= 3 {
		return
	}
	m.followers = m.followers[:len(m.followers)-1]
}

func (m *model) retuneFollowers() {
	for _, f := range m.followers {
		f.springX = harmonica.NewSpring(harmonica.FPS(fps), m.freq, m.damping)
		f.springY = harmonica.NewSpring(harmonica.FPS(fps), m.freq, m.damping)
	}
}

func newFollower(order int, freq, damping float64, rng *rand.Rand) *follower {
	baseRadius := 5.0 + float64(order)*1.35
	radius := baseRadius * (0.7 + rng.Float64()*0.6)
	speed := 0.3 + rng.Float64()*0.6 + float64(order)*0.03
	phase := rng.Float64() * 2 * math.Pi
	paletteSeed := rng.Float64()
	offsetSeed := rng.Float64()
	return &follower{
		order:       order,
		radius:      radius,
		speed:       speed,
		phase:       phase,
		paletteSeed: paletteSeed,
		offsetSeed:  offsetSeed,
		trace:       make([]vector, 0, maxTrail),
		springX:     harmonica.NewSpring(harmonica.FPS(fps), freq, damping),
		springY:     harmonica.NewSpring(harmonica.FPS(fps), freq, damping),
	}
}

func (f *follower) step(target vector, formation formationMode, stageW, stageH, t, dt float64, count int, mood moodTheme) {
	if count < 1 {
		count = 1
	}
	var offset vector
	switch formation {
	case formationHalo:
		f.phase = math.Mod(f.phase+f.speed*dt, 2*math.Pi)
		ellipse := 0.55 + 0.25*math.Sin(t*0.8+float64(f.order)*0.3)
		offset.x = math.Cos(f.phase) * f.radius * ellipse
		offset.y = math.Sin(f.phase) * f.radius * 0.6 * ellipse
	case formationRibbon:
		wave := math.Sin(t*1.4 + float64(f.order)*0.7)
		offset.x = -float64(f.order) * (1.9 + 0.4*math.Sin(t*0.6))
		offset.y = wave * stageH * 0.09
		f.phase = math.Mod(f.phase+f.speed*dt*0.6, 2*math.Pi)
		offset.x += math.Cos(f.phase+wave) * 2.4
	case formationBloom:
		petalCount := 3 + (f.order % 5)
		bloom := (math.Sin(t*0.7+float64(petalCount)) + 1) / 2
		radius := f.radius * (0.6 + 0.5*bloom)
		f.phase = math.Mod(f.phase+f.speed*dt*1.2, 2*math.Pi)
		offset.x = math.Cos(f.phase*float64(petalCount)) * radius
		offset.y = math.Sin(f.phase*float64(petalCount)) * radius * 0.6
	case formationHelix:
		depth := (float64(f.order) / float64(count-1)) - 0.5
		helixRadius := stageW * 0.16
		offset.x = math.Sin(t*0.9+depth*math.Pi*2) * helixRadius
		offset.y = depth*stageH*0.6 + math.Cos(t*1.6+depth*4)*4
		f.phase = math.Mod(f.phase+f.speed*dt, 2*math.Pi)
		offset.x += math.Cos(f.phase+depth*6) * 3
	}

	targetX := clamp(target.x+offset.x, 0, stageW-1)
	targetY := clamp(target.y+offset.y, 0, stageH-1)

	f.pos.x, f.vel.x = f.springX.Update(f.pos.x, f.vel.x, targetX)
	f.pos.y, f.vel.y = f.springY.Update(f.pos.y, f.vel.y, targetY)

	f.pos.x = clamp(f.pos.x, 0, stageW-1)
	f.pos.y = clamp(f.pos.y, 0, stageH-1)

	f.trace = append(f.trace, f.pos)
	if len(f.trace) > maxTrail {
		f.trace = f.trace[len(f.trace)-maxTrail:]
	}
}

func (m model) View() string {
	if !m.ready {
		return "harmonica is tuning resonances..."
	}

	mood := m.currentMood()
	canvas := m.prepareCanvas(mood)
	m.paintTrails(canvas, mood)
	m.paintSeeds(canvas, mood)
	m.paintTarget(canvas, mood)

	var builder strings.Builder
	for y := 0; y < m.canvasHeight; y++ {
		row := canvas[y]
		for x := 0; x < m.canvasWidth; x++ {
			builder.WriteString(m.renderCell(row[x]))
		}
		builder.WriteRune('\n')
	}

	builder.WriteString(m.renderFooter())
	if m.showHelp {
		helper := m.help
		helper.ShowAll = true
		builder.WriteString("\n")
		builder.WriteString(helpBoxStyle.Render(helper.View(m.keys)))
	}
	return builder.String()
}

func (m *model) prepareCanvas(theme moodTheme) [][]cell {
	canvas := make([][]cell, m.canvasHeight)
	for y := 0; y < m.canvasHeight; y++ {
		row := make([]cell, m.canvasWidth)
		for x := 0; x < m.canvasWidth; x++ {
			wav := math.Sin(float64(x)*0.11+m.t*0.35) + math.Cos(float64(y)*0.09-m.t*0.21+float64(x)*0.03)
			intensity := (wav + 2) / 4
			glyph := theme.wispGlyphs[int(intensity*float64(len(theme.wispGlyphs)))%len(theme.wispGlyphs)]
			fg := theme.colorAt(0.15 + intensity*0.35)
			bg := theme.background
			if theme.shader != nil {
				sGlyph, sFG, sBG := theme.shader(float64(x), float64(y), m.t, m.canvasWidth, m.canvasHeight, theme)
				if sGlyph != 0 {
					glyph = sGlyph
				}
				if sFG != "" {
					fg = sFG
				}
				if sBG != "" {
					bg = sBG
				}
			}
			row[x] = cell{
				ch:       glyph,
				fg:       fg,
				bg:       bg,
				bold:     false,
				priority: 0,
			}
		}
		canvas[y] = row
	}
	return canvas
}

func (m *model) paintTrails(canvas [][]cell, theme moodTheme) {
	for _, f := range m.followers {
		trailLen := len(f.trace)
		if trailLen == 0 {
			continue
		}
		for i := 0; i < trailLen; i++ {
			p := f.trace[i]
			x := int(math.Round(p.x))
			y := int(math.Round(p.y))
			if x < 0 || y < 0 || x >= m.canvasWidth || y >= m.canvasHeight {
				continue
			}
			ratio := float64(i+1) / float64(trailLen)
			strength := math.Pow(ratio, 1.3)
			colorMix := clamp(strength*0.8+f.paletteSeed*0.3, 0, 1)
			fg := theme.colorAt(colorMix)
			glyph := theme.trailGlyphs[min(int(strength*float64(len(theme.trailGlyphs))), len(theme.trailGlyphs)-1)]
			priority := 1
			if i == trailLen-1 {
				glyph = '@'
				priority = 3
			}
			bg := canvas[y][x].bg
			cell := cell{ch: glyph, fg: fg, bg: bg, bold: i >= trailLen-2, priority: priority}
			if cell.priority >= canvas[y][x].priority {
				canvas[y][x] = cell
			}
		}
	}
}

func (m *model) paintSeeds(canvas [][]cell, theme moodTheme) {
	for _, s := range m.seeds {
		x := int(math.Round(s.pos.x))
		y := int(math.Round(s.pos.y))
		if x < 0 || y < 0 || x >= m.canvasWidth || y >= m.canvasHeight {
			continue
		}
		glow := clamp(1-(s.life/s.ttl), 0, 1)
		fg := blendHex(s.color, theme.colorAt(0.98), 1-(glow*0.65))
		bg := canvas[y][x].bg
		cell := cell{ch: s.glyph, fg: fg, bg: bg, bold: true, priority: 4}
		if cell.priority >= canvas[y][x].priority {
			canvas[y][x] = cell
		}
	}
}

func (m *model) paintTarget(canvas [][]cell, theme moodTheme) {
	tx := int(math.Round(m.target.x))
	ty := int(math.Round(m.target.y))
	if tx < 0 || ty < 0 || tx >= m.canvasWidth || ty >= m.canvasHeight {
		return
	}
	bg := canvas[ty][tx].bg
	cell := cell{ch: '#', fg: theme.accent, bg: bg, bold: true, priority: 5}
	canvas[ty][tx] = cell
}

func tieDyeShader(x, y, t float64, width, height int, theme moodTheme) (rune, string, string) {
	if width == 0 || height == 0 {
		return 0, "", ""
	}
	cx := float64(width-1) / 2
	cy := float64(height-1) / 2
	dx := (x - cx) / float64(width)
	dy := (y - cy) / float64(height)
	radius := math.Sqrt(dx*dx + dy*dy)
	angle := math.Atan2(dy, dx)
	swirl := radius*18 + angle*6 - t*1.4
	wave := (math.Sin(swirl) + 1) / 2
	petals := math.Sin(angle*8 + t*0.9)
	mix := math.Mod(wave*0.7+radius*0.5+petals*0.2, 1)
	if mix < 0 {
		mix += 1
	}
	fg := theme.colorAt(mix)
	bgMix := clamp(wave*0.6+0.2, 0, 1)
	bg := blendHex(fg, theme.background, 1-bgMix)
	var glyph rune
	if len(theme.wispGlyphs) > 0 {
		idx := int(math.Mod(math.Abs(petals)*float64(len(theme.wispGlyphs)), float64(len(theme.wispGlyphs))))
		glyph = theme.wispGlyphs[idx]
	}
	if glyph == 0 {
		glyph = '~'
	}
	return glyph, fg, bg
}

func (m *model) renderFooter() string {
	scene := scenes[m.sceneIndex]
	formation := formations[indexOfFormation(m.formation)]
	mood := m.currentMood()

	bits := []string{
		fmt.Sprintf("%s %s", infoTitle.Render("scene"), infoValue.Render(scene.name)),
		fmt.Sprintf("%s %s", infoTitle.Render("formation"), infoValue.Render(formation.name)),
		fmt.Sprintf("%s %s", infoTitle.Render("mood"), infoValue.Render(mood.name)),
		fmt.Sprintf("%s %s", infoTitle.Render("mode"), infoValue.Render(modeLabel(m.autop))),
		fmt.Sprintf("%s %.2f", infoTitle.Render("freq"), m.freq),
		fmt.Sprintf("%s %.2f", infoTitle.Render("damping"), m.damping),
		fmt.Sprintf("%s %d", infoTitle.Render("muses"), len(m.followers)),
	}

	footer := statusStyle.Render(strings.Join(bits, "  "))
	short := m.help.ShortHelpView(m.keys.ShortHelp())
	banner := bannerStyle.Render("harmonic garden")

	lines := []string{
		frameStyle.Render(strings.Repeat("─", max(m.canvasWidth, len(footer)))),
		lipgloss.JoinHorizontal(lipgloss.Left, banner, "  ", scene.description),
		footer,
		short,
	}
	return strings.Join(lines, "\n")
}

func (m *model) renderCell(c cell) string {
	key := fmt.Sprintf("%s|%s|%t", c.fg, c.bg, c.bold)
	style, ok := m.styleCache[key]
	if !ok {
		style = lipgloss.NewStyle()
		if c.fg != "" {
			style = style.Foreground(lipgloss.Color(c.fg))
		}
		if c.bg != "" {
			style = style.Background(lipgloss.Color(c.bg))
		}
		if c.bold {
			style = style.Bold(true)
		}
		m.styleCache[key] = style
	}
	return style.Render(string(c.ch))
}

func modeLabel(autop bool) string {
	if autop {
		return "auto"
	}
	return "manual"
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func blendHex(a, b string, t float64) string {
	r1, g1, b1 := hexToRGB(a)
	r2, g2, b2 := hexToRGB(b)
	mix := func(v1, v2 int) int {
		return int(math.Round(float64(v1) + (float64(v2)-float64(v1))*t))
	}
	return fmt.Sprintf("#%02X%02X%02X", mix(r1, r2), mix(g1, g2), mix(b1, b2))
}

func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func perlin2(x, y float64) float64 {
	// lightweight value noise for organic drift
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf := x - float64(xi)
	yf := y - float64(yi)

	topRight := hash2(xi+1, yi+1)
	topLeft := hash2(xi, yi+1)
	bottomRight := hash2(xi+1, yi)
	bottomLeft := hash2(xi, yi)

	u := fade(xf)
	v := fade(yf)

	lerpTop := lerp(topLeft, topRight, u)
	lerpBottom := lerp(bottomLeft, bottomRight, u)
	return lerp(lerpBottom, lerpTop, v)
}

func hash2(x, y int) float64 {
	n := x*374761393 + y*668265263
	n = (n ^ (n >> 13)) * 1274126177
	n = n ^ (n >> 16)
	return float64(n%1024)/512 - 1
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
