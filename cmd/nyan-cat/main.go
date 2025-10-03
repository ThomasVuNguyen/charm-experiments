package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg struct{}

type star struct {
	x, y   int
	bright bool
}

type pixel struct {
	char  string
	color lipgloss.Color
}

type page int

const (
	pageJellyfishHorse page = iota
	pageCactusOctopus
	pageClockworkButterfly
	pageGlowmushroomSloth
	pageCrystalSpider
	pageNoodleWhale
	pageEyestalkTurtle
	pageFeatherFish
	pageGeometricBee
	pageVoidSquid
	pagePrismaticWorm
	pageTentacleTree
	pageFloatingBrain
	totalPages
)

type model struct {
	width       int
	height      int
	frame       int
	creatureX   int
	creatureY   int
	currentPage page
	harmonics   []harmonicWave
	flowField   []flowParticle
	energyOrbs  []energyOrb
	spirals     []spiral
	pulses      []pulse
	wisps       []wisp
	time        float64
	moodIndex   int
	formField   []formParticle
}

type harmonicWave struct {
	x, y      float64
	frequency float64
	amplitude float64
	phase     float64
	color     lipgloss.Color
}

type flowParticle struct {
	x, y       float64
	vx, vy     float64
	life       float64
	intensity  float64
	color      lipgloss.Color
	glyph      rune
}

type energyOrb struct {
	x, y      float64
	radius    float64
	pulse     float64
	color     lipgloss.Color
	intensity float64
}

type spiral struct {
	centerX, centerY float64
	angle           float64
	radius          float64
	growth          float64
	color           lipgloss.Color
}

type pulse struct {
	x, y       float64
	radius     float64
	expansion  float64
	intensity  float64
	color      lipgloss.Color
}

type wisp struct {
	x, y     float64
	trail    []wispPoint
	velocity float64
	color    lipgloss.Color
	glow     float64
}

type wispPoint struct {
	x, y     float64
	alpha    float64
}

type formParticle struct {
	x, y      float64
	velocity  float64
	size      float64
	rotation  float64
	color     lipgloss.Color
	form      int
}

const pixelBlock = "█"

type moodTheme struct {
	name        string
	palette     []lipgloss.Color
	background  lipgloss.Color
	accent      lipgloss.Color
	glyphs      []rune
	intensity   float64
}

var (
	bizarreMoods = []moodTheme{
		{
			name:       "Cosmic Mutation",
			palette:    []lipgloss.Color{lipgloss.Color("93"), lipgloss.Color("129"), lipgloss.Color("201"), lipgloss.Color("213")},
			background: lipgloss.Color("0"),
			accent:     lipgloss.Color("15"),
			glyphs:     []rune{'∞', '◊', '⟡', '⧨', '⬢'},
			intensity:  0.8,
		},
		{
			name:       "Acid Dream",
			palette:    []lipgloss.Color{lipgloss.Color("196"), lipgloss.Color("226"), lipgloss.Color("46"), lipgloss.Color("51")},
			background: lipgloss.Color("22"),
			accent:     lipgloss.Color("15"),
			glyphs:     []rune{'~', '≈', '∿', '◯', '◉'},
			intensity:  1.2,
		},
		{
			name:       "Void Ripple",
			palette:    []lipgloss.Color{lipgloss.Color("240"), lipgloss.Color("244"), lipgloss.Color("248"), lipgloss.Color("15")},
			background: lipgloss.Color("0"),
			accent:     lipgloss.Color("93"),
			glyphs:     []rune{'·', '⋅', '∘', '○', '●'},
			intensity:  0.6,
		},
		{
			name:       "Neural Bloom",
			palette:    []lipgloss.Color{lipgloss.Color("21"), lipgloss.Color("39"), lipgloss.Color("45"), lipgloss.Color("87")},
			background: lipgloss.Color("17"),
			accent:     lipgloss.Color("201"),
			glyphs:     []rune{'※', '⚡', '✦', '❋', '✧'},
			intensity:  0.9,
		},
	}

	weirdColors = []lipgloss.Color{
		lipgloss.Color("93"),  // Mystic purple
		lipgloss.Color("201"), // Neon pink
		lipgloss.Color("226"), // Electric yellow
		lipgloss.Color("51"),  // Cyber cyan
		lipgloss.Color("196"), // Acid red
		lipgloss.Color("129"), // Void magenta
		lipgloss.Color("46"),  // Alien green
		lipgloss.Color("21"),  // Deep space blue
	}

	// Bizarre Creature pixel art
	// Legend: each creature uses unique color codes for surreal anatomy
	// . = transparent, K = black outline, W = white
	
	// Jellyfish-Horse: floating equine with trailing tentacles
	jellyfishHorseFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", ".", ".", "K", "K", "K", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", "K", "J", "J", "J", "J", "J", "K", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", "K", "J", "J", "T", "T", "T", "T", "J", "J", "K", ".", ".", ".", "."},
			{".", ".", "K", "K", "J", "J", "T", "T", "W", "K", "T", "T", "J", "J", "K", "K", ".", "."},
			{".", "K", "J", "J", "J", "T", "T", "T", "T", "T", "T", "T", "T", "J", "J", "J", "K", "."},
			{"K", "J", "J", "J", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "J", "J", "J", "K"},
			{"T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T", "T"},
			{"T", ".", "T", ".", "T", ".", "T", ".", "T", ".", "T", ".", "T", ".", "T", ".", "T", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{"T", ".", ".", "T", ".", ".", "T", ".", ".", "T", ".", ".", "T", ".", ".", "T", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", "T", ".", ".", "T", ".", ".", "T", ".", ".", "T", ".", ".", "T", ".", ".", "T", "."},
		},
	}

	// Cactus-Octopus: spiny succulent with writhing arms
	cactusOctopusFrames = [][][]string{
		{
			{".", ".", ".", ".", "C", ".", ".", ".", "C", ".", ".", ".", "C", ".", ".", "."},
			{".", ".", ".", "C", "C", "C", ".", "C", "C", "C", ".", "C", "C", "C", ".", "."},
			{".", ".", "C", "C", "S", "C", "C", "C", "S", "C", "C", "C", "S", "C", "C", "."},
			{".", "C", "C", "S", "S", "S", "C", "S", "S", "S", "C", "S", "S", "S", "C", "C"},
			{"C", "C", "S", "S", "W", "S", "S", "S", "K", "S", "S", "S", "W", "S", "S", "C"},
			{"C", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "C"},
			{"C", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "C"},
			{".", "C", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "S", "C", "."},
			{".", ".", "C", "C", "S", "S", "S", "S", "S", "S", "S", "S", "C", "C", ".", "."},
			{".", ".", ".", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", ".", ".", "."},
		},
	}

	// Clockwork Butterfly: mechanical wings with gears
	clockworkButterflyFrames = [][][]string{
		{
			{".", "M", "G", "M", ".", ".", ".", ".", ".", ".", "M", "G", "M", "."},
			{"M", "G", "M", "G", "M", ".", ".", ".", ".", "M", "G", "M", "G", "M"},
			{"G", "M", "G", "M", "G", "M", "K", "K", "M", "G", "M", "G", "M", "G"},
			{"M", "G", "M", "G", "M", "G", "K", "K", "G", "M", "G", "M", "G", "M"},
			{"G", "M", "G", "M", "G", "M", "K", "K", "M", "G", "M", "G", "M", "G"},
			{"M", "G", "M", "G", "M", ".", "K", "K", ".", "M", "G", "M", "G", "M"},
			{".", "M", "G", "M", ".", ".", ".", ".", ".", ".", "M", "G", "M", "."},
		},
	}

	// Space Whale pixel art (24x10)
	spaceWhaleFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", "K", "K", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "K", "K", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", "K", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "K", ".", ".", ".", ".", ".", "."},
			{".", "K", "N", "N", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "N", "N", "K", ".", ".", ".", ".", "."},
			{".", "K", "N", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "N", "K", ".", ".", ".", ".", "."},
			{".", "K", "N", "N", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "L", "N", "N", "K", ".", ".", ".", ".", "."},
			{".", ".", "K", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "K", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", "K", "K", "N", "N", "N", "N", "N", "N", "N", "N", "N", "N", "K", "K", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", ".", ".", ".", ".", ".", ".", ".", ".", "."},
		},
	}

	// Cyber Cat pixel art (20x12)
	cyberCatFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", "K", ".", ".", ".", ".", ".", ".", "K", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", "K", "C", "K", ".", ".", ".", ".", "K", "C", "K", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", "K", "C", "C", "C", "K", "K", "K", "K", "C", "C", "C", "K", ".", ".", ".", "."},
			{".", ".", ".", "K", "C", "C", "M", "C", "C", "C", "C", "C", "C", "M", "C", "C", "K", ".", ".", "."},
			{".", ".", "K", "C", "C", "C", "C", "C", "C", "K", "K", "C", "C", "C", "C", "C", "C", "K", ".", "."},
			{".", ".", "K", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "K", ".", "."},
			{".", ".", ".", "K", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "K", ".", ".", "."},
			{".", ".", ".", ".", "K", "K", "C", "C", "C", "C", "C", "C", "C", "C", "K", "K", ".", ".", ".", "."},
			{".", ".", ".", "K", ".", ".", "K", ".", ".", ".", ".", ".", ".", "K", ".", ".", "K", ".", ".", "."},
			{".", ".", ".", "K", ".", ".", "K", ".", ".", ".", ".", ".", ".", "K", ".", ".", "K", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
		},
	}

	// Add remaining creatures (truncated for space - continuing with just a few more)
	forestFoxFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", "K", ".", ".", ".", "K", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", "K", "A", "K", ".", "K", "A", "K", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", "K", "A", "A", "A", "K", "A", "A", "A", "K", ".", ".", ".", "."},
			{".", ".", ".", ".", "K", "A", "A", "A", "A", "A", "A", "A", "A", "A", "K", ".", ".", "."},
			{".", ".", ".", "K", "A", "A", "W", "K", "A", "A", "A", "K", "W", "A", "A", "K", ".", "."},
			{".", ".", "K", "A", "A", "A", "A", "A", "A", "K", "A", "A", "A", "A", "A", "A", "K", "."},
			{".", ".", "K", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "K", "."},
			{".", ".", ".", "K", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "A", "K", ".", "."},
			{".", ".", ".", ".", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
		},
	}

	pixelColors = map[string]lipgloss.Color{
		"K": lipgloss.Color("0"),   // Black outline
		"W": lipgloss.Color("15"),  // White
		"J": lipgloss.Color("93"),  // Jellyfish purple
		"T": lipgloss.Color("51"),  // Translucent cyan (tentacles)
		"C": lipgloss.Color("46"),  // Cactus green
		"S": lipgloss.Color("226"), // Spines yellow
		"M": lipgloss.Color("244"), // Metal gray (clockwork)
		"G": lipgloss.Color("208"), // Gear bronze
		"B": lipgloss.Color("130"), // Brown fur
		"F": lipgloss.Color("196"), // Fire red
		"E": lipgloss.Color("21"),  // Electric blue
		"V": lipgloss.Color("129"), // Void purple
		"N": lipgloss.Color("201"), // Neon pink
		"Y": lipgloss.Color("226"), // Glowing yellow
		"R": lipgloss.Color("196"), // Crimson red
		"P": lipgloss.Color("93"),  // Prismatic purple
		"A": lipgloss.Color("39"),  // Aqua blue
		"O": lipgloss.Color("208"), // Orange
		"I": lipgloss.Color("87"),  // Iridescent silver
		"U": lipgloss.Color("99"),  // Ultraviolet
		"X": lipgloss.Color("160"), // Exotic red
		"Z": lipgloss.Color("240"), // Zinc gray
		"Q": lipgloss.Color("45"),  // Quantum cyan
		"H": lipgloss.Color("220"), // Holographic gold
		"L": lipgloss.Color("82"),  // Luminous green
		"D": lipgloss.Color("88"),  // Deep crimson
	}

	// Corgi pixel art (18x10)
	corgiFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", "K", "K", "K", "K", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", "K", "K", "K", "O", "O", "O", "O", "K", "K", "K", ".", ".", ".", "."},
			{".", ".", ".", "K", "O", "O", "O", "O", "O", "O", "O", "O", "O", "O", "K", ".", ".", "."},
			{".", ".", "K", "O", "W", "K", "O", "O", "O", "O", "O", "O", "K", "W", "O", "K", ".", "."},
			{".", ".", "K", "O", "O", "O", "O", "O", "K", "K", "O", "O", "O", "O", "O", "K", ".", "."},
			{".", ".", "K", "O", "O", "O", "O", "O", "O", "O", "O", "O", "O", "O", "O", "K", ".", "."},
			{".", ".", ".", "K", "O", "O", "O", "O", "O", "O", "O", "O", "O", "O", "K", ".", ".", "."},
			{".", ".", ".", ".", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", ".", ".", ".", "."},
			{".", ".", ".", "K", ".", ".", "K", ".", ".", ".", ".", "K", ".", ".", "K", ".", ".", "."},
			{".", ".", ".", "K", ".", ".", "K", ".", ".", ".", ".", "K", ".", ".", "K", ".", ".", "."},
		},
	}

	// Dragon pixel art (22x12)
	dragonFrames = [][][]string{
		{
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", "K", "K", "K", "K", ".", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", "K", "D", "D", "D", "D", "K", ".", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", "K", "D", "D", "F", "K", "F", "D", "K", ".", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", "K", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", ".", ".", ".", "."},
			{".", ".", ".", "K", "K", "K", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", ".", ".", "."},
			{".", ".", "K", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", ".", "."},
			{".", "K", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", "."},
			{".", "K", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", "."},
			{".", ".", "K", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "D", "K", ".", ".", ".", "."},
			{".", ".", ".", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", "K", ".", ".", ".", ".", "."},
			{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."},
		},
	}
)

func newModel() model {
	rand.Seed(time.Now().UnixNano())

	m := model{
		width:       80,
		height:      24,
		creatureX:   30,
		creatureY:   10,
		currentPage: pageJellyfishHorse,
		time:        0,
		moodIndex:   0,
	}

	// Initialize harmonic waves for flowing backgrounds
	for i := 0; i < 12; i++ {
		m.harmonics = append(m.harmonics, harmonicWave{
			x:         rand.Float64() * 80,
			y:         rand.Float64() * 24,
			frequency: 0.5 + rand.Float64()*3,
			amplitude: 2 + rand.Float64()*4,
			phase:     rand.Float64() * 6.28,
			color:     weirdColors[rand.Intn(len(weirdColors))],
		})
	}

	// Initialize flow field particles
	for i := 0; i < 60; i++ {
		m.flowField = append(m.flowField, flowParticle{
			x:         rand.Float64() * 80,
			y:         rand.Float64() * 24,
			vx:        (rand.Float64() - 0.5) * 2,
			vy:        (rand.Float64() - 0.5) * 2,
			life:      1.0,
			intensity: rand.Float64(),
			color:     weirdColors[rand.Intn(len(weirdColors))],
			glyph:     bizarreMoods[0].glyphs[rand.Intn(len(bizarreMoods[0].glyphs))],
		})
	}

	// Initialize energy orbs
	for i := 0; i < 8; i++ {
		m.energyOrbs = append(m.energyOrbs, energyOrb{
			x:         rand.Float64() * 80,
			y:         rand.Float64() * 24,
			radius:    1 + rand.Float64()*3,
			pulse:     rand.Float64() * 6.28,
			color:     weirdColors[rand.Intn(len(weirdColors))],
			intensity: rand.Float64(),
		})
	}

	// Initialize spirals
	for i := 0; i < 5; i++ {
		m.spirals = append(m.spirals, spiral{
			centerX: rand.Float64() * 80,
			centerY: rand.Float64() * 24,
			angle:   0,
			radius:  0,
			growth:  0.1 + rand.Float64()*0.2,
			color:   weirdColors[rand.Intn(len(weirdColors))],
		})
	}

	// Initialize pulses
	for i := 0; i < 6; i++ {
		m.pulses = append(m.pulses, pulse{
			x:         rand.Float64() * 80,
			y:         rand.Float64() * 24,
			radius:    0,
			expansion: 0.2 + rand.Float64()*0.3,
			intensity: 1.0,
			color:     weirdColors[rand.Intn(len(weirdColors))],
		})
	}

	// Initialize wisps
	for i := 0; i < 10; i++ {
		m.wisps = append(m.wisps, wisp{
			x:        rand.Float64() * 80,
			y:        rand.Float64() * 24,
			trail:    make([]wispPoint, 8),
			velocity: 0.5 + rand.Float64(),
			color:    weirdColors[rand.Intn(len(weirdColors))],
			glow:     rand.Float64(),
		})
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Recalculate creature position
		m.creatureX = m.width / 3
		m.creatureY = m.height / 2
		return m, nil

	case tickMsg:
		m.frame++
		m.time += 0.016 // 60fps delta time

		// Move creature in complex harmonic patterns
		m.creatureY = m.height/2 + int(4*math.Sin(m.time*1.5)) + int(2*math.Cos(m.time*2.3))
		m.creatureX = m.width/3 + int(2*math.Sin(m.time*0.8))

		// Update harmonic wave system (inspired by harmonic-garden)
		for i := range m.harmonics {
			m.harmonics[i].phase += m.harmonics[i].frequency * 0.02
			m.harmonics[i].x += math.Sin(m.harmonics[i].phase) * 0.5
			m.harmonics[i].y += math.Cos(m.harmonics[i].phase*1.3) * 0.3
			
			// Wrap around edges
			if m.harmonics[i].x < 0 {
				m.harmonics[i].x = float64(m.width)
			} else if m.harmonics[i].x > float64(m.width) {
				m.harmonics[i].x = 0
			}
			if m.harmonics[i].y < 0 {
				m.harmonics[i].y = float64(m.height)
			} else if m.harmonics[i].y > float64(m.height) {
				m.harmonics[i].y = 0
			}
		}

		// Update flow field particles
		for i := range m.flowField {
			// Apply flow field forces
			angle := math.Sin(m.flowField[i].x*0.1) + math.Cos(m.flowField[i].y*0.1) + m.time*0.5
			m.flowField[i].vx += math.Cos(angle) * 0.1
			m.flowField[i].vy += math.Sin(angle) * 0.1
			
			// Apply damping
			m.flowField[i].vx *= 0.98
			m.flowField[i].vy *= 0.98
			
			// Update position
			m.flowField[i].x += m.flowField[i].vx
			m.flowField[i].y += m.flowField[i].vy
			
			// Wrap around
			if m.flowField[i].x < 0 {
				m.flowField[i].x = float64(m.width)
			} else if m.flowField[i].x > float64(m.width) {
				m.flowField[i].x = 0
			}
			if m.flowField[i].y < 0 {
				m.flowField[i].y = float64(m.height)
			} else if m.flowField[i].y > float64(m.height) {
				m.flowField[i].y = 0
			}
			
			// Update intensity
			m.flowField[i].intensity = 0.3 + 0.7*math.Abs(math.Sin(m.time*2+float64(i)*0.5))
		}

		// Update energy orbs
		for i := range m.energyOrbs {
			m.energyOrbs[i].pulse += 0.1
			m.energyOrbs[i].intensity = 0.5 + 0.5*math.Sin(m.energyOrbs[i].pulse)
			m.energyOrbs[i].radius = 1 + 2*m.energyOrbs[i].intensity
		}

		// Update spirals
		for i := range m.spirals {
			m.spirals[i].angle += m.spirals[i].growth
			m.spirals[i].radius += m.spirals[i].growth * 0.5
			if m.spirals[i].radius > 20 {
				m.spirals[i].radius = 0
				m.spirals[i].centerX = rand.Float64() * float64(m.width)
				m.spirals[i].centerY = rand.Float64() * float64(m.height)
			}
		}

		// Update pulses
		for i := range m.pulses {
			m.pulses[i].radius += m.pulses[i].expansion
			m.pulses[i].intensity = math.Max(0, 1.0-m.pulses[i].radius/15.0)
			if m.pulses[i].radius > 15 {
				m.pulses[i].radius = 0
				m.pulses[i].x = rand.Float64() * float64(m.width)
				m.pulses[i].y = rand.Float64() * float64(m.height)
				m.pulses[i].intensity = 1.0
			}
		}

		// Update wisps
		for i := range m.wisps {
			// Add current position to trail
			copy(m.wisps[i].trail[1:], m.wisps[i].trail[0:len(m.wisps[i].trail)-1])
			m.wisps[i].trail[0] = wispPoint{
				x:     m.wisps[i].x,
				y:     m.wisps[i].y,
				alpha: 1.0,
			}
			
			// Update trail alpha
			for j := range m.wisps[i].trail {
				m.wisps[i].trail[j].alpha *= 0.9
			}
			
			// Move wisp
			angle := m.time*0.5 + float64(i)*0.8
			m.wisps[i].x += math.Cos(angle) * m.wisps[i].velocity
			m.wisps[i].y += math.Sin(angle*1.3) * m.wisps[i].velocity * 0.7
			
			// Wrap around
			if m.wisps[i].x < 0 {
				m.wisps[i].x = float64(m.width)
			} else if m.wisps[i].x > float64(m.width) {
				m.wisps[i].x = 0
			}
			if m.wisps[i].y < 0 {
				m.wisps[i].y = float64(m.height)
			} else if m.wisps[i].y > float64(m.height) {
				m.wisps[i].y = 0
			}
		}

		// Cycle through mood themes periodically
		if m.frame%600 == 0 {
			m.moodIndex = (m.moodIndex + 1) % len(bizarreMoods)
		}

		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			m.currentPage = (m.currentPage - 1 + totalPages) % totalPages
	case "right", "l", "tab":
			m.currentPage = (m.currentPage + 1) % totalPages
		case "1":
			m.currentPage = pageJellyfishHorse
		case "2":
			m.currentPage = pageCactusOctopus
		case "3":
			m.currentPage = pageClockworkButterfly
		case "4":
			m.currentPage = pageGlowmushroomSloth
		case "5":
			m.currentPage = pageCrystalSpider
		case "6":
			m.currentPage = pageNoodleWhale
		case "7":
			m.currentPage = pageEyestalkTurtle
		case "8":
			m.currentPage = pageFeatherFish
		case "9":
			m.currentPage = pageGeometricBee
		case "0":
			m.currentPage = pageVoidSquid
		case "m":
			m.moodIndex = (m.moodIndex + 1) % len(bizarreMoods)
		}
	}

	return m, nil
}

func (m model) View() string {
	// Create pixel grid 
	grid := make([][]pixel, m.height)
	for i := range grid {
		grid[i] = make([]pixel, m.width)
		for j := range grid[i] {
			grid[i][j] = pixel{char: " ", color: lipgloss.Color("0")}
		}
	}

	// Draw creature-specific background and creature
	switch m.currentPage {
	case pageJellyfishHorse:
		m.drawUnderwaterBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0])
	case pageCactusOctopus:
		m.drawDesertBackground(grid)
		m.drawBizarreCreature(grid, cactusOctopusFrames[0])
	case pageClockworkButterfly:
		m.drawMechanicalBackground(grid)
		m.drawBizarreCreature(grid, clockworkButterflyFrames[0])
	case pageGlowmushroomSloth:
		m.drawBioluminescentBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageCrystalSpider:
		m.drawCrystallineBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageNoodleWhale:
		m.drawNoodleBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageEyestalkTurtle:
		m.drawPsychedelicBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageFeatherFish:
		m.drawAerialBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageGeometricBee:
		m.drawGeometricBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageVoidSquid:
		m.drawVoidBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pagePrismaticWorm:
		m.drawPrismaticBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageTentacleTree:
		m.drawForestBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	case pageFloatingBrain:
		m.drawMentalBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0]) // placeholder
	default:
		m.drawHarmonicBackground(grid)
		m.drawBizarreCreature(grid, jellyfishHorseFrames[0])
	}

	// Convert grid to styled string
	var output strings.Builder
	for y, row := range grid {
		for _, p := range row {
			if p.char == " " {
				output.WriteString(" ")
			} else {
				style := lipgloss.NewStyle().Foreground(p.color)
				output.WriteString(style.Render(p.char))
			}
		}
		if y < len(grid)-1 {
			output.WriteString("\n")
		}
	}

	// Add page indicator and controls
	pageNames := []string{"Jellyfish Horse", "Cactus Octopus", "Clockwork Butterfly", "Glowmushroom Sloth", "Crystal Spider", "Noodle Whale", "Eyestalk Turtle", "Feather Fish", "Geometric Bee", "Void Squid", "Prismatic Worm", "Tentacle Tree", "Floating Brain"}
	currentPageName := pageNames[m.currentPage]
	currentMood := bizarreMoods[m.moodIndex]
	
	pageIndicator := lipgloss.NewStyle().
		Foreground(currentMood.accent).Bold(true).
		Render(fmt.Sprintf("[%d/%d] %s", int(m.currentPage)+1, int(totalPages), currentPageName))
	
	moodIndicator := lipgloss.NewStyle().
		Foreground(currentMood.palette[0]).
		Render(fmt.Sprintf("Mood: %s", currentMood.name))
	
	controls := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("← → or h/l to switch • 1-9,0 for direct • m for mood • q to quit")

	footer := lipgloss.JoinVertical(lipgloss.Left, "", pageIndicator, moodIndicator, controls)

	return output.String() + footer
}

func (m model) drawHarmonicBackground(grid [][]pixel) {
	mood := bizarreMoods[m.moodIndex]
	
	// Fill background with spaces (empty ASCII background)
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: mood.background}
		}
	}
	
	// Generate flowing ASCII field using sine waves and flow fields
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			// Create layered noise field for ASCII selection
			noise1 := math.Sin(float64(x)*0.1 + m.time*0.8) * math.Cos(float64(y)*0.15 + m.time*0.6)
			noise2 := math.Sin(float64(x)*0.2 + float64(y)*0.1 + m.time*1.2) * 0.7
			noise3 := math.Cos(float64(x)*0.05 + float64(y)*0.08 + m.time*0.4) * 0.5
			
			combined := (noise1 + noise2 + noise3) / 3.0
			
			// Map noise to ASCII characters and intensity
			intensity := math.Abs(combined)
			
			if intensity > 0.2 {
				// Select ASCII character based on noise value and mood
				var char string
				var color lipgloss.Color
				
				if intensity < 0.4 {
					// Subtle background characters
					chars := []string{".", "·", ":", ";", "'", "`"}
					char = chars[int(math.Abs(combined*100))%len(chars)]
					color = mood.palette[0]
				} else if intensity < 0.6 {
					// Medium intensity - flowing patterns
					chars := []string{"~", "-", "=", "≈", "∼", "⌒", "∿"}
					char = chars[int(math.Abs(combined*100))%len(chars)]
					color = mood.palette[1]
				} else if intensity < 0.8 {
					// Higher intensity - complex patterns
					chars := []string{"∞", "◊", "⟡", "⧨", "⬢", "◯", "◈"}
					char = chars[int(math.Abs(combined*100))%len(chars)]
					color = mood.palette[2]
				} else {
					// Highest intensity - accent characters
					char = string(mood.glyphs[int(math.Abs(combined*100))%len(mood.glyphs)])
					color = mood.accent
				}
				
				grid[y][x] = pixel{char: char, color: color}
			}
		}
	}
	
	// Draw harmonic waves as ASCII trails
	for _, wave := range m.harmonics {
		for t := 0.0; t < 6.28; t += 0.3 {
			x := wave.x + wave.amplitude*math.Cos(t+wave.phase)
			y := wave.y + wave.amplitude*math.Sin(t*wave.frequency+wave.phase)
			
			ix, iy := int(x), int(y)
			if ix >= 0 && ix < m.width && iy >= 0 && iy < m.height {
				intensity := math.Abs(math.Sin(t + wave.phase))
				if intensity > 0.4 {
					// Use flowing wave characters
					waveChars := []string{"~", "≈", "∿", "⌒", "∼"}
					char := waveChars[int(t*5)%len(waveChars)]
					grid[iy][ix] = pixel{char: char, color: wave.color}
				}
			}
		}
	}
	
	// Draw flow field particles as drifting ASCII
	for _, p := range m.flowField {
		x, y := int(p.x), int(p.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height && p.intensity > 0.3 {
			grid[y][x] = pixel{char: string(p.glyph), color: p.color}
		}
	}
	
	// Draw energy orbs as ASCII mandalas
	for _, orb := range m.energyOrbs {
		cx, cy := int(orb.x), int(orb.y)
		r := int(orb.radius * orb.intensity)
		
		if r > 0 {
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					distance := math.Sqrt(float64(dx*dx + dy*dy))
					if distance <= float64(r) {
						x, y := cx+dx, cy+dy
						if x >= 0 && x < m.width && y >= 0 && y < m.height {
							alpha := 1.0 - distance/float64(r)
							if alpha > 0.6 {
								// Inner core
								grid[y][x] = pixel{char: "●", color: orb.color}
							} else if alpha > 0.3 {
								// Middle ring
								grid[y][x] = pixel{char: "○", color: orb.color}
							} else if alpha > 0.1 {
								// Outer ring
								grid[y][x] = pixel{char: "·", color: orb.color}
							}
						}
					}
				}
			}
		}
	}
	
	// Draw spirals as ASCII curves
	for _, spiral := range m.spirals {
		for i := 0.0; i < spiral.radius && i < 20; i += 0.8 {
			angle := spiral.angle + i*0.3
			x := spiral.centerX + i*math.Cos(angle)
			y := spiral.centerY + i*math.Sin(angle)
			
			ix, iy := int(x), int(y)
			if ix >= 0 && ix < m.width && iy >= 0 && iy < m.height {
				// Choose spiral character based on angle
				spiralChars := []string{"◦", "⋄", "◊", "⬢", "⬟", "⟡"}
				char := spiralChars[int(i)%len(spiralChars)]
				grid[iy][ix] = pixel{char: char, color: spiral.color}
			}
		}
	}
	
	// Draw pulses as expanding ASCII rings
	for _, pulse := range m.pulses {
		if pulse.intensity > 0.1 {
			cx, cy := int(pulse.x), int(pulse.y)
			r := int(pulse.radius)
			
			if r > 0 && r < 20 {
				// Draw ASCII circle
				for angle := 0.0; angle < 6.28; angle += 0.4 {
					x := cx + int(float64(r)*math.Cos(angle))
					y := cy + int(float64(r)*math.Sin(angle))
					
					if x >= 0 && x < m.width && y >= 0 && y < m.height {
						// Choose ring character based on pulse intensity
						if pulse.intensity > 0.7 {
							grid[y][x] = pixel{char: "◉", color: pulse.color}
						} else if pulse.intensity > 0.4 {
							grid[y][x] = pixel{char: "◯", color: pulse.color}
						} else {
							grid[y][x] = pixel{char: "○", color: pulse.color}
						}
					}
				}
			}
		}
	}
	
	// Draw wisps as flowing ASCII trails
	for _, wisp := range m.wisps {
		for i, point := range wisp.trail {
			if point.alpha > 0.1 {
				x, y := int(point.x), int(point.y)
				if x >= 0 && x < m.width && y >= 0 && y < m.height {
					// Use different trail characters based on position in trail
					trailChars := []string{"·", ":", "∶", "⁚", "‥", "…", "⋯", "⋱"}
					if i < len(trailChars) {
						grid[y][x] = pixel{char: trailChars[i], color: wisp.color}
					}
				}
			}
		}
	}
}

func (m model) drawBizarreCreature(grid [][]pixel, creatureFrame [][]string) {
	for y, row := range creatureFrame {
		for x, pix := range row {
			py := m.creatureY - len(creatureFrame)/2 + y
			px := m.creatureX + x*2 // Double width for better visibility
			
			if py >= 0 && py < m.height && px >= 0 && px < m.width && pix != "." {
				if color, exists := pixelColors[pix]; exists {
					grid[py][px] = pixel{char: pixelBlock, color: color}
					if px+1 < m.width {
						grid[py][px+1] = pixel{char: pixelBlock, color: color}
					}
				}
			}
		}
	}
}

// Underwater background for Jellyfish Horse - flowing currents and bubbles
func (m model) drawUnderwaterBackground(grid [][]pixel) {
	// Deep ocean blue base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("17")}
		}
	}
	
	// Flowing water currents
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			currentX := math.Sin(float64(x)*0.1 + m.time*1.5) * math.Cos(float64(y)*0.2 + m.time*0.8)
			currentY := math.Cos(float64(x)*0.15 + m.time*1.2) * math.Sin(float64(y)*0.1 + m.time*1.0)
			intensity := math.Abs(currentX + currentY)
			
			if intensity > 0.3 {
				if intensity < 0.5 {
					grid[y][x] = pixel{char: "~", color: lipgloss.Color("39")}
				} else if intensity < 0.7 {
					grid[y][x] = pixel{char: "≈", color: lipgloss.Color("45")}
				} else {
					grid[y][x] = pixel{char: "∿", color: lipgloss.Color("51")}
				}
			}
		}
	}
	
	// Floating bubbles
	for _, orb := range m.energyOrbs {
		x, y := int(orb.x), int(orb.y-m.time*10) // Bubbles rise
		if y < 0 { y += m.height } // Wrap around
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			bubbleChars := []string{"○", "◯", "◦", "∘"}
			char := bubbleChars[int(orb.pulse*4)%len(bubbleChars)]
			grid[y][x] = pixel{char: char, color: lipgloss.Color("87")}
		}
	}
}

// Desert background for Cactus Octopus - sand patterns and heat waves
func (m model) drawDesertBackground(grid [][]pixel) {
	// Sandy base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("58")}
		}
	}
	
	// Sand dune patterns
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			dune := math.Sin(float64(x)*0.2 + m.time*0.3) * math.Cos(float64(y)*0.1)
			wind := math.Sin(float64(x)*0.05 + float64(y)*0.1 + m.time*2.0) * 0.8
			combined := math.Abs(dune + wind)
			
			if combined > 0.3 {
				if combined < 0.5 {
					grid[y][x] = pixel{char: ".", color: lipgloss.Color("220")}
				} else if combined < 0.7 {
					grid[y][x] = pixel{char: ":", color: lipgloss.Color("226")}
				} else {
					grid[y][x] = pixel{char: "∴", color: lipgloss.Color("208")}
				}
			}
		}
	}
	
	// Heat shimmers
	for _, wave := range m.harmonics {
		for t := 0.0; t < 6.28; t += 0.4 {
			x := wave.x + wave.amplitude*math.Sin(t+wave.phase*2)
			y := wave.y + wave.amplitude*0.3*math.Cos(t*2+wave.phase)
			ix, iy := int(x), int(y)
			if ix >= 0 && ix < m.width && iy >= 0 && iy < m.height {
				shimmers := []string{"'", "`", "\"", "'"}
				char := shimmers[int(t*2)%len(shimmers)]
				grid[iy][ix] = pixel{char: char, color: lipgloss.Color("226")}
			}
		}
	}
}

// Mechanical background for Clockwork Butterfly - gears and steam
func (m model) drawMechanicalBackground(grid [][]pixel) {
	// Dark metal base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("240")}
		}
	}
	
	// Gear mechanisms
	for y := 0; y < m.height; y += 6 {
		for x := 0; x < m.width; x += 8 {
			if rand.Float64() > 0.4 {
				rotation := m.time * 2.0
				gearChars := []string{"⚙", "⊕", "⊗", "⊙"}
				char := gearChars[int(rotation*4)%len(gearChars)]
				if x < m.width && y < m.height {
					grid[y][x] = pixel{char: char, color: lipgloss.Color("244")}
				}
			}
		}
	}
	
	// Steam pipes and pressure
	for _, pulse := range m.pulses {
		x, y := int(pulse.x), int(pulse.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			steamChars := []string{"│", "┃", "║", "▓"}
			char := steamChars[int(pulse.intensity*4)%len(steamChars)]
			grid[y][x] = pixel{char: char, color: lipgloss.Color("248")}
		}
	}
	
	// Rivets and bolts
	for x := 3; x < m.width; x += 7 {
		for y := 2; y < m.height; y += 5 {
			if x < m.width && y < m.height {
				grid[y][x] = pixel{char: "•", color: lipgloss.Color("242")}
			}
		}
	}
}

// Bioluminescent background for Glowmushroom Sloth
func (m model) drawBioluminescentBackground(grid [][]pixel) {
	// Dark forest base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("22")}
		}
	}
	
	// Glowing spores
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			spore := math.Sin(float64(x)*0.3 + m.time*3.0) * math.Cos(float64(y)*0.2 + m.time*2.5)
			glow := math.Abs(spore)
			
			if glow > 0.4 {
				if glow < 0.6 {
					grid[y][x] = pixel{char: "·", color: lipgloss.Color("46")}
				} else if glow < 0.8 {
					grid[y][x] = pixel{char: "∘", color: lipgloss.Color("82")}
				} else {
					grid[y][x] = pixel{char: "◉", color: lipgloss.Color("120")}
				}
			}
		}
	}
	
	// Mushroom caps
	for _, orb := range m.energyOrbs {
		x, y := int(orb.x), int(orb.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			grid[y][x] = pixel{char: "🍄", color: lipgloss.Color("206")}
		}
	}
}

// Crystalline background for Crystal Spider
func (m model) drawCrystallineBackground(grid [][]pixel) {
	// Crystal cave base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("54")}
		}
	}
	
	// Crystal formations
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			crystal := math.Sin(float64(x)*0.4 + m.time*0.5) * math.Cos(float64(y)*0.3 + m.time*0.8)
			refraction := math.Abs(crystal)
			
			if refraction > 0.3 {
				crystalChars := []string{"◊", "⬟", "⟡", "◈", "⬢", "⌬"}
				char := crystalChars[int(refraction*20)%len(crystalChars)]
				colors := []lipgloss.Color{lipgloss.Color("129"), lipgloss.Color("93"), lipgloss.Color("201")}
				color := colors[int(refraction*10)%len(colors)]
				grid[y][x] = pixel{char: char, color: color}
			}
		}
	}
}

// Flowing noodle background for Noodle Whale
func (m model) drawNoodleBackground(grid [][]pixel) {
	// Brothy base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("178")}
		}
	}
	
	// Flowing noodles
	for _, wisp := range m.wisps {
		for i, point := range wisp.trail {
			x, y := int(point.x), int(point.y)
			if x >= 0 && x < m.width && y >= 0 && y < m.height && i < 6 {
				noodleChars := []string{"∿", "～", "〜", "⌇", "≋", "∽"}
				grid[y][x] = pixel{char: noodleChars[i%len(noodleChars)], color: lipgloss.Color("226")}
			}
		}
	}
	
	// Steam bubbles
	for _, orb := range m.energyOrbs {
		x, y := int(orb.x), int(orb.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			grid[y][x] = pixel{char: "○", color: lipgloss.Color("15")}
		}
	}
}

// Psychedelic background for Eyestalk Turtle
func (m model) drawPsychedelicBackground(grid [][]pixel) {
	// Trippy base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("53")}
		}
	}
	
	// Kaleidoscope patterns
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			pattern := math.Sin(float64(x)*0.2 + m.time*4.0) * math.Cos(float64(y)*0.2 + m.time*3.0)
			spiral := math.Sin(math.Sqrt(float64(x*x+y*y))*0.3 + m.time*2.0)
			combined := math.Abs(pattern + spiral)
			
			if combined > 0.3 {
				psychChars := []string{"◐", "◑", "◒", "◓", "●", "○", "◉"}
				char := psychChars[int(combined*20)%len(psychChars)]
				colors := []lipgloss.Color{lipgloss.Color("201"), lipgloss.Color("93"), lipgloss.Color("129"), lipgloss.Color("219")}
				color := colors[int(combined*15)%len(colors)]
				grid[y][x] = pixel{char: char, color: color}
			}
		}
	}
}

// Aerial background for Feather Fish
func (m model) drawAerialBackground(grid [][]pixel) {
	// Sky base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("39")}
		}
	}
	
	// Wind currents
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			wind := math.Sin(float64(x)*0.1 + m.time*2.0) * math.Cos(float64(y)*0.05 + m.time*1.5)
			if math.Abs(wind) > 0.4 {
				windChars := []string{"~", "≈", "⌇", "∼"}
				char := windChars[int(math.Abs(wind)*10)%len(windChars)]
				grid[y][x] = pixel{char: char, color: lipgloss.Color("87")}
			}
		}
	}
	
	// Floating feathers
	for _, p := range m.flowField {
		x, y := int(p.x), int(p.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height && p.intensity > 0.5 {
			grid[y][x] = pixel{char: "❋", color: lipgloss.Color("15")}
		}
	}
}

// Geometric background for Geometric Bee
func (m model) drawGeometricBackground(grid [][]pixel) {
	// Grid base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("0")}
		}
	}
	
	// Hexagonal honeycomb
	for y := 0; y < m.height; y += 4 {
		for x := 0; x < m.width; x += 6 {
			if x < m.width && y < m.height {
				pulse := math.Sin(m.time*3.0 + float64(x+y)*0.1)
				if pulse > 0 {
					grid[y][x] = pixel{char: "⬢", color: lipgloss.Color("226")}
				}
			}
		}
	}
	
	// Geometric lines
	for _, pulse := range m.pulses {
		x, y := int(pulse.x), int(pulse.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			lineChars := []string{"─", "│", "╱", "╲"}
			char := lineChars[int(pulse.intensity*4)%len(lineChars)]
			grid[y][x] = pixel{char: char, color: lipgloss.Color("46")}
		}
	}
}

// Void background for Void Squid
func (m model) drawVoidBackground(grid [][]pixel) {
	// Pure darkness
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("0")}
		}
	}
	
	// Sparse void particles
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			void := math.Sin(float64(x)*0.01 + m.time*0.1) * math.Cos(float64(y)*0.01 + m.time*0.2)
			if math.Abs(void) > 0.8 {
				voidChars := []string{".", "·", "∘", "○"}
				char := voidChars[int(math.Abs(void)*10)%len(voidChars)]
				grid[y][x] = pixel{char: char, color: lipgloss.Color("240")}
			}
		}
	}
}

// Prismatic background for Prismatic Worm
func (m model) drawPrismaticBackground(grid [][]pixel) {
	// Rainbow refraction base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("0")}
		}
	}
	
	// Light spectrum
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			spectrum := math.Sin(float64(x)*0.3 + m.time*2.0) * math.Cos(float64(y)*0.2 + m.time*1.5)
			refraction := math.Abs(spectrum)
			
			if refraction > 0.3 {
				spectrumChars := []string{"▀", "▄", "█", "░", "▒", "▓"}
				char := spectrumChars[int(refraction*20)%len(spectrumChars)]
				rainbowColors := []lipgloss.Color{
					lipgloss.Color("196"), lipgloss.Color("208"), lipgloss.Color("226"),
					lipgloss.Color("46"), lipgloss.Color("51"), lipgloss.Color("21"), lipgloss.Color("93"),
				}
				color := rainbowColors[int(refraction*30)%len(rainbowColors)]
				grid[y][x] = pixel{char: char, color: color}
			}
		}
	}
}

// Forest background for Tentacle Tree
func (m model) drawForestBackground(grid [][]pixel) {
	// Dark forest floor
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("22")}
		}
	}
	
	// Creeping vines
	for _, wisp := range m.wisps {
		for i, point := range wisp.trail {
			x, y := int(point.x), int(point.y)
			if x >= 0 && x < m.width && y >= 0 && y < m.height && i < 5 {
				vineChars := []string{"│", "┃", "║", "╎", "╏"}
				grid[y][x] = pixel{char: vineChars[i%len(vineChars)], color: lipgloss.Color("28")}
			}
		}
	}
	
	// Swaying branches
	for y := 0; y < m.height; y += 3 {
		sway := math.Sin(m.time*1.5 + float64(y)*0.2)
		x := int(float64(m.width/4) + sway*3)
		if x >= 0 && x < m.width {
			grid[y][x] = pixel{char: "┬", color: lipgloss.Color("130")}
		}
	}
}

// Mental/neural background for Floating Brain
func (m model) drawMentalBackground(grid [][]pixel) {
	// Neural network base
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			grid[y][x] = pixel{char: " ", color: lipgloss.Color("17")}
		}
	}
	
	// Synaptic connections
	for _, pulse := range m.pulses {
		x, y := int(pulse.x), int(pulse.y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			if pulse.intensity > 0.5 {
				grid[y][x] = pixel{char: "●", color: lipgloss.Color("21")}
			} else {
				grid[y][x] = pixel{char: "○", color: lipgloss.Color("39")}
			}
		}
	}
	
	// Neural pathways
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			neural := math.Sin(float64(x)*0.2 + m.time*3.0) * math.Cos(float64(y)*0.1 + m.time*2.0)
			if math.Abs(neural) > 0.4 {
				pathChars := []string{"─", "│", "┌", "┐", "└", "┘"}
				char := pathChars[int(math.Abs(neural)*20)%len(pathChars)]
				grid[y][x] = pixel{char: char, color: lipgloss.Color("45")}
			}
		}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
