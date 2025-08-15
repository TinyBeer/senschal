package img

import (
	"image"
	"image/color"
	"image/gif"
	"os"
)

// 创建灰度调色板（0-255级灰度）
func grayPalette() color.Palette {
	pal := make(color.Palette, 256)
	for i := 0; i < 256; i++ {
		pal[i] = color.Gray{Y: uint8(i)} // 每个索引对应一个灰度值
	}
	return pal
}

// 将*image.Gray转换为*image.Paletted
func grayToPaletted(gray *image.Gray) *image.Paletted {
	bounds := gray.Bounds()
	pal := grayPalette() // 使用灰度调色板
	paletted := image.NewPaletted(bounds, pal)

	// 遍历每个像素，映射灰度值到调色板索引
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 灰度值Y直接作为调色板索引（因为调色板按0-255顺序排列）
			grayVal := gray.GrayAt(x, y).Y
			paletted.SetColorIndex(x, y, uint8(grayVal))
		}
	}
	return paletted
}

// 灰度转换
func toGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	grayImg := image.NewGray(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 如果完全透明，直接设为黑色
			if a == 0 {
				grayImg.SetGray(x, y, color.Gray{0})
				continue
			}

			// 转换为8位颜色值
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// 灰度公式：0.299*R + 0.587*G + 0.114*B
			gray := uint8(0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8))
			grayImg.SetGray(x, y, color.Gray{gray})
		}
	}
	return grayImg
}

// Sobel边缘检测
func sobelEdgeDetection(grayImg *image.Gray) *image.Gray {
	bounds := grayImg.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	result := image.NewGray(bounds)

	// Sobel算子
	Gx := [3][3]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	Gy := [3][3]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	// 边缘处理，忽略最外一层像素
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var sumX, sumY int

			// 应用卷积核
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := int(grayImg.GrayAt(x+kx, y+ky).Y)
					sumX += pixel * Gx[ky+1][kx+1]
					sumY += pixel * Gy[ky+1][kx+1]
				}
			}

			// 计算梯度幅值，使用近似值避免开方运算
			magnitude := abs(sumX) + abs(sumY)

			// 限制在0-255范围内
			if magnitude > 255 {
				magnitude = 255
			}

			result.SetGray(x, y, color.Gray{uint8(magnitude)})
		}
	}

	return result
}

// 绝对值函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// 处理GIF的每一帧
func ProcessGIF(inputPath, outputPath string) error {
	// 打开输入文件
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 解码GIF
	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		return err
	}

	// 处理每一帧
	for i, frame := range gifImg.Image {
		// 转换为灰度图
		gray := toGray(frame)
		// 应用边缘检测
		edges := sobelEdgeDetection(gray)
		// 替换原帧
		gifImg.Image[i] = grayToPaletted(edges)
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 编码并保存处理后的GIF
	return gif.EncodeAll(outFile, gifImg)
}
