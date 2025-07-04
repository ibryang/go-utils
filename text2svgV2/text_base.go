package text2svgV2

import (
	"errors"
	"fmt"
	"image/color"
	"math"

	"github.com/tdewolff/canvas"
)

// GenerateBaseText 生成基础文本
func GenerateBaseText(option TextOption) (*canvas.Canvas, error) {
	if option.Text == "" {
		return nil, errors.New("text is required")
	}
	// 判断fontColor类型
	var fontColor []color.RGBA
	var strokeColor color.RGBA = canvas.Black
	var strokeWidth float64 = option.StrokeWidth
	if option.RenderMode == 0 {
		option.RenderMode = RenderString
	}
	if option.FontColor != nil {
		if color, ok := option.FontColor.(string); ok {
			if c, ok := ColorMap[color]; ok {
				fontColor = append(fontColor, c)
			} else {
				fontColor = append(fontColor, canvas.Hex(color))
			}
		}
		if color, ok := option.FontColor.([]string); ok {
			for _, c := range color {
				if v, ok := ColorMap[c]; ok {
					fontColor = append(fontColor, v)
				} else {
					fontColor = append(fontColor, canvas.Hex(c))
				}
			}
		}
	}
	if option.StrokeColor != nil {
		if color, ok := option.StrokeColor.(string); ok {
			if c, ok := ColorMap[color]; ok {
				strokeColor = c
			} else {
				strokeColor = canvas.Hex(color)
			}
		}
	}

	font, err := LoadFont(option.FontPath)
	if err != nil {
		return nil, err
	}
	fontface := font.Face(option.FontSize, option.FontColor)

	// 首先计算所有字符的确切边界，以确定整个字符串的实际可视范围
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	// 临时保存每个字符的路径和前进宽度
	var charPaths []canvas.Path
	var advances []float64
	var exactWidth float64
	var exactHeight float64

	// 计算整个字符串的确切边界框
	var xPos float64
	var colorIndices []int
	if option.RenderMode == RenderChar {
		colorCount := 0
		for _, char := range option.Text {
			path, advance, err := fontface.ToPath(string(char))
			if err != nil {
				return nil, err
			}

			bounds := path.Bounds()
			minX = math.Min(minX, bounds.X0+xPos)
			minY = math.Min(minY, bounds.Y0)
			maxX = math.Max(maxX, bounds.X1+xPos)
			maxY = math.Max(maxY, bounds.Y1)

			charPaths = append(charPaths, *path)
			advances = append(advances, advance)
			xPos += advance
			if char == ' ' {
				colorIndices = append(colorIndices, -1)
				continue
			}
			colorIndices = append(colorIndices, colorCount%len(fontColor))
			colorCount++
		}

		// 计算精确的宽度和高度
		exactWidth = maxX - minX
		exactHeight = maxY - minY
	}
	var path *canvas.Path
	if option.RenderMode == RenderString {
		p, _, err := fontface.ToPath(option.Text)
		if err != nil {
			return nil, err
		}
		p = p.Transform(canvas.Matrix{
			{1, 0, -p.Bounds().X0},
			{0, 1, 0},
		})
		exactWidth = p.Bounds().W()
		minY = p.Bounds().Y0
		maxY = p.Bounds().Y1
		exactHeight = maxY - minY
		path = p
	}
	fmt.Println(exactWidth, exactHeight)
	// 创建一个尺寸刚好容纳所有字符的画布
	textCanvas := canvas.New(exactWidth, exactHeight)
	textCtx := canvas.NewContext(textCanvas)

	if option.RectOption != nil {
		DrawRect(textCtx, *option.RectOption)
	}

	// 绘制每个字符
	xPos = -minX // 调整起始位置，确保所有内容都可见
	yPos := -minY
	if option.RenderMode == RenderChar {
		for i, path := range charPaths {
			// 将路径绘制到画布上
			if colorIndices[i] == -1 {
				textCtx.SetFillColor(canvas.Transparent)
			} else {
				textCtx.SetFillColor(fontColor[colorIndices[i]])
			}
			if strokeWidth > 0 {
				textCtx.SetStrokeColor(strokeColor)
				textCtx.SetStrokeWidth(strokeWidth)
				textCtx.Stroke()
			}
			textCtx.DrawPath(xPos, yPos, &path)
			textCtx.Fill()

			// 更新x位置
			xPos += advances[i]
		}
	}
	if option.RenderMode == RenderString {
		textCtx.SetFillColor(fontColor[0])
		if strokeWidth > 0 {
			textCtx.SetStrokeColor(strokeColor)
			textCtx.SetStrokeWidth(strokeWidth)
			textCtx.Stroke()
		}
		textCtx.DrawPath(0, -minY, path)
		textCtx.Fill()
	}
	if len(option.ExtraText) > 0 {
		for _, extOption := range option.ExtraText {
			DrawExtraText(textCtx, extOption)
		}
	}
	// 根据参数设置翻转
	textCanvas = ReverseCanvas(textCanvas, option.ReverseX, option.ReverseY)
	scaleX := 1.0
	scaleY := 1.0
	// 支持 minSize, maxSize 逻辑
	// minSize 优先级高于 maxSize
	if option.MinSize {
		// MinSize 模式，按较小的缩放因子等比缩放
		scaleX = option.Width / exactWidth
		scaleY = option.Height / exactHeight
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		scaleX = scale
		scaleY = scale
		option.Width = exactWidth * scaleX
		option.Height = exactHeight * scaleY
	} else if option.MaxSize {
		// MaxSize 模式，按较大的缩放因子等比缩放
		scaleX = option.Width / exactWidth
		scaleY = option.Height / exactHeight
		scale := scaleX
		if scaleY > scaleX {
			scale = scaleY
		}
		scaleX = scale
		scaleY = scale
		option.Width = exactWidth * scaleX
		option.Height = exactHeight * scaleY
	} else if option.Width == 0 && option.Height == 0 {
		option.Width = exactWidth
		option.Height = exactHeight
	} else if option.Width > 0 {
		// 只设置了宽度，按比例计算高度
		scaleX = option.Width / exactWidth
		scaleY = scaleX
		option.Height = exactHeight * scaleY
	} else if option.Height > 0 {
		// 只设置了高度，按比例计算宽度
		scaleY = option.Height / exactHeight
		scaleX = scaleY
		option.Width = exactWidth * scaleX
	}

	textCanvas.W = option.Width
	textCanvas.H = option.Height

	textCanvas.Transform(canvas.Matrix{
		{scaleX, 0, 0},
		{0, scaleY, 0},
	})
	return textCanvas, nil
}

func ReverseCanvas(c *canvas.Canvas, reversX, reversY bool) *canvas.Canvas {
	if !reversX && !reversY {
		return c
	}
	scaleX := 1.0
	scaleY := 1.0
	offsetX := 0.0
	offsetY := 0.0
	if reversX {
		scaleX = -1.0
		offsetX = c.W
	}
	if reversY {
		scaleY = -1.0
		offsetY = c.H
	}
	c.Transform(canvas.Matrix{
		{scaleX, 0, offsetX},
		{0, scaleY, offsetY},
	})
	return c
}

func DrawExtraText(c *canvas.Context, extOption ExtraTextOption) {
	textCanvas, err := GenerateBaseText(extOption.TextOption)
	if err != nil {
		return
	}
	offsetX := 0.0
	offsetY := 0.0
	if extOption.Align == TextAlignCenter {
		offsetX = (c.Width() - textCanvas.W) / 2
	} else if extOption.Align == TextAlignRight {
		offsetX = c.Width() - textCanvas.W
	} else if extOption.Align == TextAlignLeft {
		offsetX = 0
	}

	if extOption.VAlign == TextVAlignCenter {
		offsetY = (c.Height() - textCanvas.H) / 2
	} else if extOption.VAlign == TextVAlignBottom {
		offsetY = c.Height() - textCanvas.H
	} else if extOption.VAlign == TextVAlignTop {
		offsetY = 0
	}

	if extOption.X > 0 {
		offsetX += extOption.X
	}
	if extOption.Y > 0 {
		offsetY += extOption.Y
	}

	textCanvas.RenderViewTo(c, canvas.Matrix{
		{1, 0, offsetX},
		{0, 1, c.Height() - offsetY - textCanvas.H},
	})
}
