package component

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	RoundedRectangle_TopLeft        = "╭"
	RoundedRectangle_TopRight       = "╮"
	RoundedRectangle_BottomLeft     = "╰"
	RoundedRectangle_BottomRight    = "╯"
	RoundedRectangle_HorizontalLine = "─"
	RoundedRectangle_VerticalLine   = "│"
)

type Container interface {
	GetHeight() int
	GetWidth() int
	GetCurrentContent(frame int) [][]StyleString
}

type StyleString struct {
	Content string
	Style   lipgloss.Style
}

func JoinStyleStringMatrix(matrix [][]StyleString) []string {
	res := make([]string, 0, len(matrix))
	for _, arr := range matrix {
		var str string
		for _, sStr := range arr {
			str += sStr.Style.Render(sStr.Content)
		}
		res = append(res, str)
	}
	return res
}

// box
type Box struct {
	direction Direction
	subs      []Container
}

func NewBox(dir Direction) *Box {
	return &Box{
		direction: dir,
	}
}

func (b *Box) AddSub(c Container) {
	b.subs = append(b.subs, c)
}

// GetCurrentContent implements Container.
func (b *Box) GetCurrentContent(frame int) [][]StyleString {
	// TODO support multi row and multi cow
	h := b.GetHeight()
	var content [][]StyleString
	if b.direction == Direction_H {
		content = make([][]StyleString, h)
	}
	for _, c := range b.subs {
		containerStrList := c.GetCurrentContent(frame)
		if b.direction == Direction_H {
			maxLen := 0
			for _, sStrList := range containerStrList {
				length := 0
				for _, sStr := range sStrList {
					length += getDisplayLength(sStr.Content)
				}
				maxLen = max(maxLen, length)
			}
			for i := range content {
				length := 0
				if i < len(containerStrList) {
					for _, sStr := range containerStrList[i] {
						length += getDisplayLength(sStr.Content)
					}
					content[i] = append(content[i], containerStrList[i]...)
				}
				if maxLen-length > 0 {
					content[i] = append(content[i], StyleString{Content: strings.Repeat(" ", maxLen-length)})
				}
			}
		} else {
			content = append(content, containerStrList...)
		}
	}

	return content
}

// GetHeight implements Container.
func (b *Box) GetHeight() int {
	height := 0
	for _, c := range b.subs {
		if b.direction == Direction_H {
			height = max(height, c.GetHeight())
		} else {
			height += c.GetHeight()
		}
	}
	return height
}

// GetWidth implements Container.
func (b *Box) GetWidth() int {
	width := 0
	for _, c := range b.subs {
		if b.direction == Direction_V {
			width = max(width, c.GetWidth())
		} else {
			width += c.GetWidth()
		}
	}
	return width
}

var _ Container = new(Box)

// text container

type InlineText struct {
	content string
	width   int
	style   lipgloss.Style
}

func NewInlineText(w int, content string) *InlineText {
	return &InlineText{
		content: content,
		width:   w,
	}

}

func NewInlineTextWithStyle(w int, content string, style lipgloss.Style) *InlineText {
	return &InlineText{
		content: content,
		width:   w,
		style:   style,
	}

}

// GetHeight implements Container.
func (t *InlineText) GetHeight() int {
	return 1
}

// GetWidth implements Container.
func (t *InlineText) GetWidth() int {
	return t.width
}

// GetCurrentContent implements Container.
func (t *InlineText) GetCurrentContent(frame int) [][]StyleString {
	return [][]StyleString{
		{
			{
				Content: slidingWindowDisplayString(t.content, t.width, frame),
				Style:   t.style,
			},
		},
	}
}

var _ Container = new(InlineText)

type Rectangle struct {
	content Container
	round   bool
	style   lipgloss.Style
}

func NewRectangle(content Container, round bool, style lipgloss.Style) *Rectangle {
	return &Rectangle{
		content: content,
		round:   round,
		style:   style,
	}
}

// GetCurrentContent implements Container.
func (r *Rectangle) GetCurrentContent(frame int) [][]StyleString {
	h, w := r.GetHeight(), r.GetWidth()
	cc := r.content.GetCurrentContent(frame)

	content := make([][]StyleString, h)
	if r.round {
		content[0] = append(content[0], StyleString{
			Content: RoundedRectangle_TopLeft + strings.Repeat(RoundedRectangle_HorizontalLine, w-2) + RoundedRectangle_TopRight,
			Style:   r.style,
		})
		for i := range cc {
			content[i+1] = append([]StyleString{{
				Content: RoundedRectangle_VerticalLine,
				Style:   r.style,
			}}, append(cc[i], StyleString{
				Content: RoundedRectangle_VerticalLine,
				Style:   r.style,
			})...)
		}
		content[h-1] = append(content[h-1], StyleString{
			Content: RoundedRectangle_BottomLeft + strings.Repeat(RoundedRectangle_HorizontalLine, w-2) + RoundedRectangle_BottomRight,
			Style:   r.style,
		})
	} else {
		panic("not support yet")
	}
	return content
}

// GetHeight implements Container.
func (r *Rectangle) GetHeight() int {
	return r.content.GetHeight() + 2
}

// GetWidth implements Container.
func (r *Rectangle) GetWidth() int {
	return r.content.GetWidth() + 2
}

var _ Container = new(Rectangle)
