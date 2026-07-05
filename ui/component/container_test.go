package component

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// JoinStyleStringMatrix
// ---------------------------------------------------------------------------

func TestJoinStyleStringMatrix_Empty(t *testing.T) {
	result := JoinStyleStringMatrix([][]StyleString{})
	assert.Empty(t, result)
}

func TestJoinStyleStringMatrix_SingleCell(t *testing.T) {
	matrix := [][]StyleString{
		{{Content: "hello"}},
	}
	result := JoinStyleStringMatrix(matrix)
	assert.Equal(t, []string{"hello"}, result)
}

func TestJoinStyleStringMatrix_MultiRowMultiCell(t *testing.T) {
	matrix := [][]StyleString{
		{
			{Content: "a"},
			{Content: "b"},
		},
		{
			{Content: "c"},
			{Content: "d"},
		},
	}
	result := JoinStyleStringMatrix(matrix)
	assert.Len(t, result, 2)
	assert.Equal(t, "ab", result[0])
	assert.Equal(t, "cd", result[1])
}

func TestJoinStyleStringMatrix_WithStyle(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	matrix := [][]StyleString{
		{{Content: "x", Style: style}},
	}
	result := JoinStyleStringMatrix(matrix)
	// With style rendered via lipgloss, output should contain the content
	assert.Contains(t, result[0], "x")
}

// ---------------------------------------------------------------------------
// Box
// ---------------------------------------------------------------------------

func TestNewBox_Vertical(t *testing.T) {
	b := NewBox(Direction_V)
	assert.NotNil(t, b)
	assert.Equal(t, 0, b.GetHeight())
	assert.Equal(t, 0, b.GetWidth())
}

func TestNewBox_Horizontal(t *testing.T) {
	b := NewBox(Direction_H)
	assert.NotNil(t, b)
	assert.Equal(t, 0, b.GetHeight())
	assert.Equal(t, 0, b.GetWidth())
}

func TestBox_AddSub(t *testing.T) {
	b := NewBox(Direction_V)
	a := NewInlineText(5, "hello")
	b.AddSub(a)
	assert.Equal(t, 1, b.GetHeight())
	assert.Equal(t, 5, b.GetWidth())
}

func TestBox_VerticalHeightWidth(t *testing.T) {
	b := NewBox(Direction_V)
	b.AddSub(NewInlineText(5, "hello"))
	b.AddSub(NewInlineText(10, "world!"))
	// Vertical: height = sum, width = max
	assert.Equal(t, 2, b.GetHeight())
	assert.Equal(t, 10, b.GetWidth())
}

func TestBox_HorizontalHeightWidth(t *testing.T) {
	b := NewBox(Direction_H)
	b.AddSub(NewInlineText(5, "hello"))
	b.AddSub(NewInlineText(10, "world!"))
	// Horizontal: height = max, width = sum
	assert.Equal(t, 1, b.GetHeight())
	assert.Equal(t, 15, b.GetWidth())
}

func TestBox_GetContent_Horizontal(t *testing.T) {
	b := NewBox(Direction_H)
	b.AddSub(NewInlineText(3, "ab"))
	b.AddSub(NewInlineText(3, "cd"))
	content := b.GetCurrentContent(0)
	assert.Len(t, content, 1)
	// InlineText pads content to width 3, each sub contributes a StyleString
	assert.Len(t, content[0], 2)
	assert.Equal(t, "ab ", content[0][0].Content)
	assert.Equal(t, "cd ", content[0][1].Content)
	rendered := JoinStyleStringMatrix(content)
	assert.Equal(t, []string{"ab cd "}, rendered)
}

func TestBox_GetContent_Vertical(t *testing.T) {
	b := NewBox(Direction_V)
	b.AddSub(NewInlineText(5, "hello"))
	b.AddSub(NewInlineText(5, "world"))
	content := b.GetCurrentContent(0)
	assert.Len(t, content, 2)
}

func TestBox_Empty(t *testing.T) {
	b := NewBox(Direction_V)
	content := b.GetCurrentContent(0)
	assert.Nil(t, content)
}

// ---------------------------------------------------------------------------
// InlineText
// ---------------------------------------------------------------------------

func TestNewInlineText(t *testing.T) {
	it := NewInlineText(10, "test content")
	assert.NotNil(t, it)
	assert.Equal(t, 1, it.GetHeight())
	assert.Equal(t, 10, it.GetWidth())
}

func TestNewInlineTextWithStyle(t *testing.T) {
	style := lipgloss.NewStyle().Bold(true)
	it := NewInlineTextWithStyle(5, "bold", style)
	assert.NotNil(t, it)
	assert.Equal(t, 5, it.GetWidth())
}

func TestInlineText_GetContent(t *testing.T) {
	it := NewInlineText(10, "hello world")
	content := it.GetCurrentContent(0)
	require.Len(t, content, 1)
	require.Len(t, content[0], 1)
	assert.Equal(t, "hello worl", content[0][0].Content) // truncated to width
}

func TestInlineText_ContentShorterThanWidth(t *testing.T) {
	it := NewInlineText(20, "short")
	content := it.GetCurrentContent(0)
	// Content is padded with spaces to match width
	assert.Equal(t, "short               ", content[0][0].Content)
}

func TestInlineText_Scrolling(t *testing.T) {
	it := NewInlineText(5, "hello world")
	content := it.GetCurrentContent(3) // scroll frame 3
	assert.Equal(t, "lo wo", content[0][0].Content)
}

// ---------------------------------------------------------------------------
// Rectangle
// ---------------------------------------------------------------------------

func TestNewRectangle(t *testing.T) {
	inner := NewInlineText(5, "hi")
	r := NewRectangle(inner, true, lipgloss.NewStyle())
	assert.NotNil(t, r)
	// height = inner(1) + 2 = 3, width = inner(5) + 2 = 7
	assert.Equal(t, 3, r.GetHeight())
	assert.Equal(t, 7, r.GetWidth())
}

func TestRectangle_GetContent(t *testing.T) {
	inner := NewInlineText(3, "ab")
	r := NewRectangle(inner, true, lipgloss.NewStyle())
	content := r.GetCurrentContent(0)
	assert.Len(t, content, 3)
	// First row: "╭───╮" (width: 3 + 2 = 5)
	assert.Len(t, content[0], 1)
	assert.Equal(t, "╭───╮", content[0][0].Content)
	// Middle row: "│ab │" (padded by InlineText to width 3)
	assert.Len(t, content[1], 3)
	assert.Equal(t, "│", content[1][0].Content)
	assert.Equal(t, "ab ", content[1][1].Content)
	assert.Equal(t, "│", content[1][2].Content)
	// Last row: "╰───╯"
	assert.Equal(t, "╰───╯", content[2][0].Content)
}

func TestRectangle_NonRoundPanics(t *testing.T) {
	inner := NewInlineText(1, "x")
	r := NewRectangle(inner, false, lipgloss.NewStyle())
	assert.Panics(t, func() {
		r.GetCurrentContent(0)
	})
}

// ---------------------------------------------------------------------------
// Direction String
// ---------------------------------------------------------------------------

func TestDirection_String(t *testing.T) {
	assert.Equal(t, "Direction_V", Direction_V.String())
	assert.Equal(t, "Direction_H", Direction_H.String())
}

func TestDirection_String_Invalid(t *testing.T) {
	var d Direction = 99
	assert.Contains(t, d.String(), "Direction(99)")
}
