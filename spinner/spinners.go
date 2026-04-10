// Package spinner provides animated spinner widgets.
package spinner

import "time"

// Definition describes a spinner animation.
type Definition struct {
	Interval time.Duration
	Frames   []string
}

// Spinners contains all available spinner definitions.
// Sourced from cli-spinners (MIT License, Sindre Sorhus).
var Spinners = map[string]Definition{
	"dots": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	},
	"dots2": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
	},
	"dots3": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"},
	},
	"dots9": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⢹", "⢺", "⢼", "⣸", "⣇", "⡧", "⡗", "⡏"},
	},
	"dots11": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"},
	},
	"line": {
		Interval: 130 * time.Millisecond,
		Frames:   []string{"-", "\\", "|", "/"},
	},
	"line2": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"⠂", "-", "–", "—", "–", "-"},
	},
	"pipe": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"},
	},
	"simpleDots": {
		Interval: 400 * time.Millisecond,
		Frames:   []string{".  ", ".. ", "...", "   "},
	},
	"simpleDotsScrolling": {
		Interval: 200 * time.Millisecond,
		Frames:   []string{".  ", ".. ", "...", " ..", "  .", "   "},
	},
	"star": {
		Interval: 70 * time.Millisecond,
		Frames:   []string{"✶", "✸", "✹", "✺", "✹", "✷"},
	},
	"star2": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"+", "x", "*"},
	},
	"flip": {
		Interval: 70 * time.Millisecond,
		Frames:   []string{"_", "_", "_", "-", "`", "`", "'", "´", "-", "_", "_", "_"},
	},
	"hamburger": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"☱", "☲", "☴"},
	},
	"growVertical": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"▁", "▃", "▄", "▅", "▆", "▇", "▆", "▅", "▄", "▃"},
	},
	"growHorizontal": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "▊", "▋", "▌", "▍", "▎"},
	},
	"balloon": {
		Interval: 140 * time.Millisecond,
		Frames:   []string{" ", ".", "o", "O", "@", "*", " "},
	},
	"balloon2": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{".", "o", "O", "°", "O", "o", "."},
	},
	"noise": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"▓", "▒", "░"},
	},
	"bounce": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"⠁", "⠂", "⠄", "⠂"},
	},
	"boxBounce": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"▖", "▘", "▝", "▗"},
	},
	"boxBounce2": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"▌", "▀", "▐", "▄"},
	},
	"triangle": {
		Interval: 50 * time.Millisecond,
		Frames:   []string{"◢", "◣", "◤", "◥"},
	},
	"arc": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"◜", "◠", "◝", "◞", "◡", "◟"},
	},
	"circle": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"◡", "⊙", "◠"},
	},
	"squareCorners": {
		Interval: 180 * time.Millisecond,
		Frames:   []string{"◰", "◳", "◲", "◱"},
	},
	"circleQuarters": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"◴", "◷", "◶", "◵"},
	},
	"circleHalves": {
		Interval: 50 * time.Millisecond,
		Frames:   []string{"◐", "◓", "◑", "◒"},
	},
	"squish": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"╫", "╪"},
	},
	"toggle": {
		Interval: 250 * time.Millisecond,
		Frames:   []string{"⊶", "⊷"},
	},
	"toggle2": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"▫", "▪"},
	},
	"toggle3": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"□", "■"},
	},
	"toggle4": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"■", "□", "▪", "▫"},
	},
	"toggle5": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"▮", "▯"},
	},
	"toggle6": {
		Interval: 300 * time.Millisecond,
		Frames:   []string{"ဝ", "၀"},
	},
	"toggle7": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⦾", "⦿"},
	},
	"toggle8": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"◍", "◌"},
	},
	"toggle9": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"◉", "◎"},
	},
	"toggle10": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"㊂", "㊀", "㊁"},
	},
	"toggle11": {
		Interval: 50 * time.Millisecond,
		Frames:   []string{"⧇", "⧆"},
	},
	"toggle12": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"☗", "☖"},
	},
	"toggle13": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"=", "*", "-"},
	},
	"arrow": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	},
	"arrow2": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"⬆️ ", "↗️ ", "➡️ ", "↘️ ", "⬇️ ", "↙️ ", "⬅️ ", "↖️ "},
	},
	"arrow3": {
		Interval: 120 * time.Millisecond,
		Frames:   []string{"▹▹▹▹▹", "▸▹▹▹▹", "▹▸▹▹▹", "▹▹▸▹▹", "▹▹▹▸▹", "▹▹▹▹▸"},
	},
	"bouncingBar": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"[    ]", "[=   ]", "[==  ]", "[=== ]", "[ ===]", "[  ==]",
			"[   =]", "[    ]", "[   =]", "[  ==]", "[ ===]", "[====]",
			"[=== ]", "[==  ]", "[=   ]",
		},
	},
	"bouncingBall": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"( ●    )", "(  ●   )", "(   ●  )", "(    ● )", "(     ●)",
			"(    ● )", "(   ●  )", "(  ●   )", "( ●    )", "(●     )",
		},
	},
	"smiley": {
		Interval: 200 * time.Millisecond,
		Frames:   []string{"😄 ", "😝 "},
	},
	"monkey": {
		Interval: 300 * time.Millisecond,
		Frames:   []string{"🙈 ", "🙈 ", "🙉 ", "🙊 "},
	},
	"hearts": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"💛 ", "💙 ", "💜 ", "💚 ", "❤️ "},
	},
	"clock": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"🕛 ", "🕐 ", "🕑 ", "🕒 ", "🕓 ", "🕔 ", "🕕 ", "🕖 ", "🕗 ", "🕘 ", "🕙 ", "🕚 "},
	},
	"earth": {
		Interval: 180 * time.Millisecond,
		Frames:   []string{"🌍 ", "🌎 ", "🌏 "},
	},
	"moon": {
		Interval: 80 * time.Millisecond,
		Frames:   []string{"🌑 ", "🌒 ", "🌓 ", "🌔 ", "🌕 ", "🌖 ", "🌗 ", "🌘 "},
	},
	"runner": {
		Interval: 140 * time.Millisecond,
		Frames:   []string{"🚶 ", "🏃 "},
	},
	"pong": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"▐⠂       ▌", "▐⠈       ▌", "▐ ⠂      ▌", "▐ ⠠      ▌",
			"▐  ⡀     ▌", "▐  ⠠     ▌", "▐   ⠂    ▌", "▐   ⠈    ▌",
			"▐    ⠂   ▌", "▐    ⠠   ▌", "▐     ⡀  ▌", "▐     ⠠  ▌",
			"▐      ⠂ ▌", "▐      ⠈ ▌", "▐       ⠂▌", "▐       ⠠▌",
			"▐       ⡀▌", "▐      ⠠ ▌", "▐      ⠂ ▌", "▐     ⠈  ▌",
			"▐     ⠂  ▌", "▐    ⠠   ▌", "▐    ⡀   ▌", "▐   ⠠    ▌",
			"▐   ⠂    ▌", "▐  ⠈     ▌", "▐  ⠂     ▌", "▐ ⠠      ▌",
			"▐ ⡀      ▌", "▐⠠       ▌",
		},
	},
	"shark": {
		Interval: 120 * time.Millisecond,
		Frames: []string{
			"▐|\\____________▌", "▐_|\\___________▌", "▐__|\\__________▌",
			"▐___|\\_________▌", "▐____|\\________▌", "▐_____|\\_______▌",
			"▐______|\\______▌", "▐_______|\\_____▌", "▐________|\\____▌",
			"▐_________|\\___▌", "▐__________|\\__▌", "▐___________|\\_▌",
			"▐____________|\\▌", "▐____________/|▌", "▐___________/|_▌",
			"▐__________/|__▌", "▐_________/|___▌", "▐________/|____▌",
			"▐_______/|_____▌", "▐______/|______▌", "▐_____/|_______▌",
			"▐____/|________▌", "▐___/|_________▌", "▐__/|__________▌",
			"▐_/|___________▌", "▐/|____________▌",
		},
	},
	"dqpb": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"d", "q", "p", "b"},
	},
	"weather": {
		Interval: 100 * time.Millisecond,
		Frames:   []string{"☀️ ", "☀️ ", "☀️ ", "🌤 ", "⛅️ ", "🌥 ", "☁️ ", "🌧 ", "🌨 ", "🌧 ", "🌨 ", "🌧 ", "🌨 ", "⛈ ", "🌨 ", "🌧 ", "🌨 ", "☁️ ", "🌥 ", "⛅️ ", "🌤 ", "☀️ ", "☀️ "},
	},
	"christmas": {
		Interval: 400 * time.Millisecond,
		Frames:   []string{"🌲", "🎄"},
	},
	"grenade": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"،  ", "′  ", " ´ ", " ‾ ", "  ⸌", "  ⸊", "  |", "  ⁎",
			"  ⁕", " ෴ ", "  ⁓", "   ", "   ", "   ",
		},
	},
	"point": {
		Interval: 125 * time.Millisecond,
		Frames:   []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"},
	},
	"layer": {
		Interval: 150 * time.Millisecond,
		Frames:   []string{"-", "=", "≡"},
	},
	"betaWave": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"ρββββββ", "βρβββββ", "ββρββββ", "βββρβββ",
			"ββββρββ", "βββββρβ", "ββββββρ",
		},
	},
	"aesthetic": {
		Interval: 80 * time.Millisecond,
		Frames: []string{
			"▰▱▱▱▱▱▱", "▰▰▱▱▱▱▱", "▰▰▰▱▱▱▱", "▰▰▰▰▱▱▱",
			"▰▰▰▰▰▱▱", "▰▰▰▰▰▰▱", "▰▰▰▰▰▰▰", "▰▱▱▱▱▱▱",
		},
	},
}

// Default is the default spinner name.
const Default = "dots"

// Get returns a spinner definition by name, or the default if not found.
func Get(name string) Definition {
	if def, ok := Spinners[name]; ok {
		return def
	}
	return Spinners[Default]
}

// Names returns all available spinner names.
func Names() []string {
	names := make([]string, 0, len(Spinners))
	for name := range Spinners {
		names = append(names, name)
	}
	return names
}
