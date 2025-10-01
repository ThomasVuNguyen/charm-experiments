# charm-experiments

A set of experimental applications using various charm.sh libraries to see how creative a CLI application can get.

## Harmonic Garden

![Harmonic Garden Demo](vhs/harmonic-garden.gif)

`cmd/harmonic-garden` is a synesthetic terminal garden that braids [Harmonica](https://github.com/charmbracelet/harmonica) springs with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

Muses orbit a living focal point while moody palettes wash the background with motion-triggered gradients. Pulse scenes, rose curves, and turbulent wander fields shape the choreography; shimmering "seeds" arc away from the center using Harmonica projectiles. The new "Cosmic Tie-Dye" mood splashes the screen with swirling sunbursts for full hippie glow, and a contextual key legend (powered by Bubbles `help`) keeps tools at hand while you improvise your own kinetic light painting.

### Run it

```bash
go run ./cmd/harmonic-garden
```

Use a terminal that supports the alternate screen buffer and 24-bit color for the best experience.

### Controls

- `space`: toggle auto/manual control of the focal point
- `tab`: cycle motion scenes (elliptic drift, rose bloom, cascade, pulse spiral, wander field)
- `f`: cycle follower formations (halo, ribbon, bloom, helix)
- `m`: cycle colour moods and ambient palettes (Aurora Bloom, Cosmic Tie-Dye, Solar Garden, Deep Current)
- Arrow keys / `h` `j` `k` `l`: nudge the target while in manual mode
- `;` / `'`: decrease / increase spring frequency
- `,` / `.`: decrease / increase damping
- `+` / `-`: grow or trim the follower troupe
- `?` or `/`: toggle the full help sheet (short hints stay in the footer)
- `q`: quit

### How it works

Each Muse owns paired Harmonica springs for the X and Y axes. Formation logic defines the latent offset space the springs try to inhabit, while animated scenes continually retarget the shared focal point. Trails capture recent motion and are re-coloured through Lip Gloss gradients so older motion cools while fresh motion blooms. Harmonica projectiles spawn “seeds” that burst away from the epicentre, adding secondary motion layers. Background wisps are synthesised per-frame with lightweight value-noise, staying in sync with the active mood palette.
