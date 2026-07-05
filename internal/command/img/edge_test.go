package img

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// grayPalette
// ---------------------------------------------------------------------------

func TestGrayPalette(t *testing.T) {
	pal := grayPalette()
	assert.Len(t, pal, 256)
	for i := 0; i < 256; i++ {
		g, ok := pal[i].(color.Gray)
		assert.True(t, ok, "index %d should be color.Gray", i)
		assert.Equal(t, uint8(i), g.Y, "index %d gray value mismatch", i)
	}
}

// ---------------------------------------------------------------------------
// abs
// ---------------------------------------------------------------------------

func TestAbs_Positive(t *testing.T) {
	assert.Equal(t, 5, abs(5))
}

func TestAbs_Negative(t *testing.T) {
	assert.Equal(t, 5, abs(-5))
}

func TestAbs_Zero(t *testing.T) {
	assert.Equal(t, 0, abs(0))
}

// ---------------------------------------------------------------------------
// toGray
// ---------------------------------------------------------------------------

func TestToGray_RGBA(t *testing.T) {
	// 2x2 RGBA image
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // red
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255}) // green
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255}) // blue
	img.Set(1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 0})     // transparent

	gray := toGray(img)
	assert.Equal(t, image.Rect(0, 0, 2, 2), gray.Bounds())

	// Red -> grayscale (0.299*255)
	assert.Equal(t, uint8(76), gray.GrayAt(0, 0).Y)
	// Green -> grayscale (0.587*255)
	assert.Equal(t, uint8(149), gray.GrayAt(1, 0).Y)
	// Blue -> grayscale (0.114*255)
	assert.Equal(t, uint8(29), gray.GrayAt(0, 1).Y)
	// Transparent -> 0
	assert.Equal(t, uint8(0), gray.GrayAt(1, 1).Y)
}

// ---------------------------------------------------------------------------
// sobelEdgeDetection
// ---------------------------------------------------------------------------

func TestSobelEdgeDetection_Uniform(t *testing.T) {
	// 3x3 uniform gray image — no edges
	gray := image.NewGray(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			gray.SetGray(x, y, color.Gray{Y: 128})
		}
	}

	edges := sobelEdgeDetection(gray)
	// Border pixels are unchanged (0), center pixel is 0 because uniform
	assert.Equal(t, uint8(0), edges.GrayAt(1, 1).Y)
}

func TestSobelEdgeDetection_Edge(t *testing.T) {
	// 3x3 with a vertical edge
	gray := image.NewGray(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		gray.SetGray(0, y, color.Gray{Y: 0})   // black
		gray.SetGray(1, y, color.Gray{Y: 255}) // white
		gray.SetGray(2, y, color.Gray{Y: 255}) // white
	}

	edges := sobelEdgeDetection(gray)
	// Center pixel should detect edge (non-zero magnitude)
	assert.NotZero(t, edges.GrayAt(1, 1).Y)
}

// ---------------------------------------------------------------------------
// grayToPaletted
// ---------------------------------------------------------------------------

func TestGrayToPaletted(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 2, 2))
	gray.SetGray(0, 0, color.Gray{Y: 0})
	gray.SetGray(1, 0, color.Gray{Y: 128})
	gray.SetGray(0, 1, color.Gray{Y: 255})
	gray.SetGray(1, 1, color.Gray{Y: 64})

	paletted := grayToPaletted(gray)
	assert.Equal(t, image.Rect(0, 0, 2, 2), paletted.Bounds())
	assert.Equal(t, uint8(0), paletted.ColorIndexAt(0, 0))
	assert.Equal(t, uint8(128), paletted.ColorIndexAt(1, 0))
	assert.Equal(t, uint8(255), paletted.ColorIndexAt(0, 1))
	assert.Equal(t, uint8(64), paletted.ColorIndexAt(1, 1))
}

// ---------------------------------------------------------------------------
// ProcessGIF (integration test with a real in-memory GIF)
// ---------------------------------------------------------------------------

func createMinimalGIF(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.gif")

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	pal := color.Palette{color.Black, color.White}
	bounds := image.Rect(0, 0, 2, 2)
	frame := image.NewPaletted(bounds, pal)
	frame.SetColorIndex(0, 0, 0) // black
	frame.SetColorIndex(1, 0, 1) // white
	frame.SetColorIndex(0, 1, 1) // white
	frame.SetColorIndex(1, 1, 0) // black

	err = gif.EncodeAll(f, &gif.GIF{
		Image: []*image.Paletted{frame},
		Delay: []int{0},
	})
	require.NoError(t, err)
	return path
}

func TestProcessGIF(t *testing.T) {
	inputPath := createMinimalGIF(t)
	outputPath := filepath.Join(t.TempDir(), "out.gif")

	err := ProcessGIF(inputPath, outputPath)
	require.NoError(t, err)

	// Verify output file exists and is a valid GIF
	outFile, err := os.Open(outputPath)
	require.NoError(t, err)
	defer outFile.Close()

	_, err = gif.DecodeAll(outFile)
	require.NoError(t, err, "output should be a valid GIF")
}

func TestProcessGIF_InputNotExist(t *testing.T) {
	err := ProcessGIF("/no/such.gif", "/tmp/out.gif")
	assert.Error(t, err)
}
