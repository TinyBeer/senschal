package component

import "github.com/charmbracelet/lipgloss"

type Container interface {
	GetHeight() int
	GetWidth() int
	GetCurrentContent(frame int) []string
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
func (b *Box) GetCurrentContent(frame int) []string {
	// TODO support multi row and multi cow
	h := b.GetHeight()
	var strList []string
	if b.direction == Direction_H {
		strList = make([]string, h)
	}
	for _, c := range b.subs {
		containerStrList := c.GetCurrentContent(frame)
		if b.direction == Direction_H {
			for i, str := range containerStrList {
				strList[i] += str
			}
		} else {
			strList = append(strList, containerStrList...)
		}
	}
	return strList
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
func (t *InlineText) GetCurrentContent(frame int) []string {
	return []string{
		t.style.Render(slidingWindowDisplayString(t.content, t.width, frame)),
	}
}

var _ Container = new(InlineText)
