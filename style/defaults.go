package style

var (
	Bold      = New().WithBold(true)
	Dim       = New().WithDim(true)
	Italic    = New().WithItalic(true)
	Underline = New().WithUnderline(true)
	Blink     = New().WithBlink(true)
	Reverse   = New().WithReverse(true)
	Strike    = New().WithStrike(true)
	Conceal   = New().WithConceal(true)
	Overline  = New().WithOverline(true)
)

var (
	Red     = New().WithForeground(StandardColor(1))
	Green   = New().WithForeground(StandardColor(2))
	Yellow  = New().WithForeground(StandardColor(3))
	Blue    = New().WithForeground(StandardColor(4))
	Magenta = New().WithForeground(StandardColor(5))
	Cyan    = New().WithForeground(StandardColor(6))
	White   = New().WithForeground(StandardColor(7))
	Black   = New().WithForeground(StandardColor(0))
)
