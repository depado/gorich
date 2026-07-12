// Package segment provides the atomic rendering unit for terminal output.
package segment

import (
	"fmt"
	"strings"

	"github.com/depado/gorich/internal/cells"
)

// ControlType represents a terminal control operation.
type ControlType int

const (
	ControlCarriageReturn ControlType = iota
	ControlShowCursor
	ControlHideCursor
	ControlCursorUp
	ControlCursorDown
	ControlCursorForward
	ControlCursorBackward
	ControlCursorMoveTo
	ControlEraseInLine
	ControlEraseInDisplay
	ControlSetTitle
	ControlBeginSyncUpdate
	ControlEndSyncUpdate
)

// ControlCode represents a control operation with optional parameters.
type ControlCode struct {
	Type   ControlType
	Params []int
	Text   string // For SetTitle/SetWindowTitle
}

// Control is a renderable that produces terminal control sequences.
type Control struct {
	Codes []ControlCode
}

// NewControl creates a new Control with the given codes.
func NewControl(codes ...ControlCode) Control {
	return Control{Codes: codes}
}

// CursorUp creates a control to move the cursor up n lines.
func CursorUp(n int) Control {
	return NewControl(ControlCode{Type: ControlCursorUp, Params: []int{n}})
}

// CursorDown creates a control to move the cursor down n lines.
func CursorDown(n int) Control {
	return NewControl(ControlCode{Type: ControlCursorDown, Params: []int{n}})
}

// CursorForward creates a control to move the cursor forward n columns.
func CursorForward(n int) Control {
	return NewControl(ControlCode{Type: ControlCursorForward, Params: []int{n}})
}

// CursorBackward creates a control to move the cursor backward n columns.
func CursorBackward(n int) Control {
	return NewControl(ControlCode{Type: ControlCursorBackward, Params: []int{n}})
}

// CursorMoveTo creates a control to move the cursor to (x, y).
func CursorMoveTo(x, y int) Control {
	return NewControl(ControlCode{Type: ControlCursorMoveTo, Params: []int{x, y}})
}

// CarriageReturn creates a control for carriage return.
func CarriageReturn() Control {
	return NewControl(ControlCode{Type: ControlCarriageReturn})
}

// ShowCursor creates a control to show the cursor.
func ShowCursor() Control {
	return NewControl(ControlCode{Type: ControlShowCursor})
}

// HideCursor creates a control to hide the cursor.
func HideCursor() Control {
	return NewControl(ControlCode{Type: ControlHideCursor})
}

// EraseInLine creates a control to erase in the current line.
// mode: 0 = cursor to end, 1 = start to cursor, 2 = entire line
func EraseInLine(mode int) Control {
	return NewControl(ControlCode{Type: ControlEraseInLine, Params: []int{mode}})
}

// EraseInDisplay creates a control to erase in the display.
// mode: 0 = cursor to end, 1 = start to cursor, 2 = entire display, 3 = entire display + scrollback
func EraseInDisplay(mode int) Control {
	return NewControl(ControlCode{Type: ControlEraseInDisplay, Params: []int{mode}})
}

// SetTitle creates a control to set the terminal title.
func SetTitle(title string) Control {
	return NewControl(ControlCode{Type: ControlSetTitle, Text: title})
}

// BeginSyncUpdate creates a control that opens a synchronized output frame
// (DEC private mode 2026). Terminals that support it buffer everything until
// EndSyncUpdate and paint it atomically, eliminating flicker. Unsupported
// terminals ignore the unknown mode.
func BeginSyncUpdate() Control {
	return NewControl(ControlCode{Type: ControlBeginSyncUpdate})
}

// EndSyncUpdate creates a control that closes a synchronized output frame.
func EndSyncUpdate() Control {
	return NewControl(ControlCode{Type: ControlEndSyncUpdate})
}

// ANSI escape sequence templates
const (
	csi = "\x1b["  // Control Sequence Introducer
	osc = "\x1b]"  // Operating System Command
	st  = "\x1b\\" // String Terminator
)

// Render produces the ANSI escape sequence for the control codes.
func (c Control) Render() string {
	if len(c.Codes) == 0 {
		return ""
	}

	var b strings.Builder
	for _, code := range c.Codes {
		b.WriteString(code.Render())
	}
	return b.String()
}

// Render produces the ANSI escape sequence for this control code.
func (cc ControlCode) Render() string {
	switch cc.Type {
	case ControlCarriageReturn:
		return "\r"
	case ControlShowCursor:
		return csi + "?25h"
	case ControlHideCursor:
		return csi + "?25l"
	case ControlCursorUp:
		n := 1
		if len(cc.Params) > 0 {
			n = cc.Params[0]
		}
		return fmt.Sprintf("%s%dA", csi, n)
	case ControlCursorDown:
		n := 1
		if len(cc.Params) > 0 {
			n = cc.Params[0]
		}
		return fmt.Sprintf("%s%dB", csi, n)
	case ControlCursorForward:
		n := 1
		if len(cc.Params) > 0 {
			n = cc.Params[0]
		}
		return fmt.Sprintf("%s%dC", csi, n)
	case ControlCursorBackward:
		n := 1
		if len(cc.Params) > 0 {
			n = cc.Params[0]
		}
		return fmt.Sprintf("%s%dD", csi, n)
	case ControlCursorMoveTo:
		x, y := 1, 1
		if len(cc.Params) >= 2 {
			x, y = cc.Params[0], cc.Params[1]
		}
		return fmt.Sprintf("%s%d;%dH", csi, y, x) // ANSI is row;col
	case ControlEraseInLine:
		mode := 0
		if len(cc.Params) > 0 {
			mode = cc.Params[0]
		}
		return fmt.Sprintf("%s%dK", csi, mode)
	case ControlEraseInDisplay:
		mode := 0
		if len(cc.Params) > 0 {
			mode = cc.Params[0]
		}
		return fmt.Sprintf("%s%dJ", csi, mode)
	case ControlSetTitle:
		return fmt.Sprintf("%s2;%s%s", osc, cells.SanitizeOSC(cc.Text), st)
	case ControlBeginSyncUpdate:
		return csi + "?2026h"
	case ControlEndSyncUpdate:
		return csi + "?2026l"
	default:
		return ""
	}
}
