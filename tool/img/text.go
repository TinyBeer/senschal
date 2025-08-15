package img

import (
	"fmt"
	"image"
	"image/gif"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type ImageTextData struct {
	Data []string `json:"data,omitempty"`
}

func ConvertImage2Text(inputPath string,
	width, height int, invert, colors bool) (*ImageTextData, error) {
	imgList, err := getImgList(inputPath)
	if err != nil {
		return nil, err
	}
	data := new(ImageTextData)
	for _, img := range imgList {
		data.Data = append(data.Data, imageToAscii(img, width, height, invert, colors))
	}
	return data, nil
}

// 打开并解码图片文件
func getImgList(path string) ([]image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imgList []image.Image
	switch filepath.Ext(path) {
	case ".gif":
		// 解码GIF
		gifImg, err := gif.DecodeAll(file)
		if err != nil {
			return nil, err
		}

		// 处理每一帧
		for _, frame := range gifImg.Image {
			imgList = append(imgList, frame)
		}
	default:
		// 解码图片
		img, _, err := image.Decode(file)
		if err != nil {
			return nil, err
		}
		imgList = append(imgList, img)
	}

	return imgList, nil
}

// 将图片转换为字符画
func imageToAscii(img image.Image, width, height int, invert, useColors bool) string {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// 如果高度为0，则根据原图比例计算
	if height == 0 {
		aspectRatio := float64(imgHeight) / float64(imgWidth)
		height = int(float64(width) * aspectRatio * 0.5) // 0.5是因为字符在终端中通常是高大于宽
	}

	// 字符集，从暗到亮
	charset := []rune("@%#*+=-:. ")
	if invert {
		// 反转字符集，从亮到暗
		for i, j := 0, len(charset)-1; i < j; i, j = i+1, j-1 {
			charset[i], charset[j] = charset[j], charset[i]
		}
	}

	// 计算缩放比例
	xRatio := float64(imgWidth) / float64(width)
	yRatio := float64(imgHeight) / float64(height)

	var result strings.Builder

	// 遍历输出的每个字符位置
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算对应原图中的坐标
			srcX := int(math.Min(float64(imgWidth-1), float64(x)*xRatio))
			srcY := int(math.Min(float64(imgHeight-1), float64(y)*yRatio))

			// 获取像素颜色
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			// 转换为8位值
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// 计算灰度值
			gray := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)

			// 映射到字符集
			index := int(gray * float64(len(charset)-1) / 255.0)
			index = int(math.Max(0, math.Min(float64(len(charset)-1), float64(index))))

			// 如果启用颜色，则添加ANSI颜色代码
			if useColors {
				// 24位颜色ANSI转义序列
				fmt.Fprintf(&result, "\033[38;2;%d;%d;%dm%c\033[0m", r8, g8, b8, charset[index])
			} else {
				result.WriteRune(charset[index])
			}
		}
		result.WriteRune('\n')
	}

	return result.String()
}
