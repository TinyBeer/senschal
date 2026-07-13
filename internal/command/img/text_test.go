package img

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// imageToAscii
// ---------------------------------------------------------------------------

func TestImageToAscii_Small(t *testing.T) {
	// 2x2 image with distinct colors
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})       // black
	img.Set(1, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // white
	img.Set(0, 1, color.RGBA{R: 128, G: 128, B: 128, A: 255}) // gray
	img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})     // red

	result := imageToAscii(img, 2, 2, false, false)
	assert.Contains(t, result, "@")
	assert.Contains(t, result, " ")
	assert.Contains(t, result, "\n")
	// 2 rows, each 2 chars + newline
	lines := len(result)
	assert.Greater(t, lines, 2)
}

func TestImageToAscii_Inverted(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})       // black
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // white

	normal := imageToAscii(img, 2, 2, false, false)
	inverted := imageToAscii(img, 2, 2, true, false)

	assert.NotEqual(t, normal, inverted, "inverted output should differ from normal")
}

func TestImageToAscii_WithColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // red

	result := imageToAscii(img, 1, 1, false, true)
	assert.Contains(t, result, "\033[38;2;255;0;0m") // ANSI red
	assert.Contains(t, result, "\033[0m")            // reset
}

func TestImageToAscii_AutoHeight(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 8)) // width=4, height=8
	result := imageToAscii(img, 4, 0, false, false)
	// Auto height = 4 * (8/4) * 0.5 = 4 lines
	lines := 0
	for _, ch := range result {
		if ch == '\n' {
			lines++
		}
	}
	assert.Equal(t, 4, lines, "should have 4 newlines for 4 output rows")
}

// ---------------------------------------------------------------------------
// getImgList and ConvertImage2Text
// ---------------------------------------------------------------------------

func createTestPNG(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(0, 1, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	err = png.Encode(f, img)
	require.NoError(t, err)
	return path
}

func TestGetImgList_PNG(t *testing.T) {
	path := createTestPNG(t)
	imgs, err := getImgList(path)
	require.NoError(t, err)
	require.Len(t, imgs, 1)
	bounds := imgs[0].Bounds()
	assert.Equal(t, 2, bounds.Dx())
	assert.Equal(t, 2, bounds.Dy())
}

func TestGetImgList_FileNotExist(t *testing.T) {
	_, err := getImgList("/no/such/file.png")
	assert.Error(t, err)
}

func TestConvertImage2Text_PNG(t *testing.T) {
	path := createTestPNG(t)
	data, err := ConvertImage2Text(path, 2, 2, false, false)
	require.NoError(t, err)
	require.NotNil(t, data)
	require.Len(t, data.Data, 1)
	assert.Contains(t, data.Data[0], "\n")
}

func TestConvertImage2Text_FileNotExist(t *testing.T) {
	_, err := ConvertImage2Text("/no/such/file.png", 2, 2, false, false)
	assert.Error(t, err)
}
