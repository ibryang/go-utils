package text2svgV2

import (
	"bytes"
	"errors"
	"math"
	"os"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

// GenerateMultipleLinesText 生成多行文本
func GenerateMultipleLinesText(option TextLineOption) (*canvas.Canvas, error) {
	if len(option.TextList) == 0 {
		return nil, errors.New("text list is required")
	}

	// 列表反转
	textList := []TextOption{}
	for i := len(option.TextList) - 1; i >= 0; i-- {
		textList = append(textList, option.TextList[i])
	}
	option.TextList = textList

	// 生成每行文本画布并计算尺寸
	textCanvases := make([]*canvas.Canvas, 0, len(option.TextList))
	maxWidth := 0.0
	totalHeight := 0.0

	for _, textOption := range option.TextList {
		textCanvas, err := GenerateBaseText(textOption)
		if err != nil {
			return nil, err
		}

		textCanvases = append(textCanvases, textCanvas)
		maxWidth = math.Max(maxWidth, textCanvas.W)
		totalHeight += textCanvas.H
	}

	// 添加行间距
	if len(option.TextList) > 1 {
		totalHeight += option.LineGap * float64(len(option.TextList)-1)
	}

	// 创建内容画布
	contentWidth := maxWidth
	contentHeight := totalHeight
	c := canvas.New(contentWidth, contentHeight)

	// 布局文本行
	yPos := 0.0
	for _, textCanvas := range textCanvases {
		// 根据对齐方式计算x位置
		xPos := 0.0 // 默认左对齐

		if option.Align == TextAlignCenter {
			xPos = (contentWidth - textCanvas.W) / 2
		} else if option.Align == TextAlignRight {
			xPos = contentWidth - textCanvas.W
		}

		// 渲染到画布
		textCanvas.RenderViewTo(c, canvas.Matrix{
			{1, 0, xPos},
			{0, 1, yPos},
		})

		// 更新位置
		yPos += textCanvas.H + option.LineGap
	}

	// 应用缩放
	scaleX := 1.0
	scaleY := 1.0

	// 计算可用内容区域（减去padding）
	availableWidth := option.Width - option.Padding[1] - option.Padding[3]
	availableHeight := option.Height - option.Padding[0] - option.Padding[2]

	if option.Width == 0 && option.Height == 0 {
		// 保持原始尺寸
	} else if option.Width > 0 && option.Height > 0 {
		// 计算考虑内边距后的缩放比例
		if availableWidth > 0 && availableHeight > 0 {
			scaleX = availableWidth / c.W
			scaleY = availableHeight / c.H
		} else {
			scaleX = option.Width / c.W
			scaleY = option.Height / c.H
		}

		// 锁定比例时，使用较小的缩放因子确保文本内容不超出容器
		if option.LockRatio {
			if scaleX > scaleY {
				scaleX = scaleY
			} else {
				scaleY = scaleX
			}
		}

		// 计算缩放后的内容尺寸
		scaledWidth := c.W * scaleX
		scaledHeight := c.H * scaleY

		// 根据对齐方式计算偏移量
		offsetX := option.Padding[3] // 默认左对齐
		offsetY := option.Padding[0] // 默认上对齐

		// 水平对齐
		if option.Align == TextAlignCenter {
			offsetX = option.Padding[3] + (availableWidth-scaledWidth)/2
		} else if option.Align == TextAlignRight {
			offsetX = option.Width - option.Padding[1] - scaledWidth
		}

		// 垂直对齐
		if option.VAlign == TextVAlignCenter {
			offsetY = option.Padding[0] + (availableHeight-scaledHeight)/2
		} else if option.VAlign == TextVAlignBottom {
			offsetY = option.Height - option.Padding[2] - scaledHeight
		}

		// 创建新画布并应用变换
		newCanvas := canvas.New(option.Width, option.Height)
		c.RenderViewTo(newCanvas, canvas.Matrix{
			{scaleX, 0, offsetX},
			{0, scaleY, offsetY},
		})
		c = newCanvas
	} else if option.Width > 0 {
		// 指定宽度自适应高度
		if availableWidth > 0 {
			scaleX = availableWidth / c.W
		} else {
			scaleX = option.Width / c.W
		}
		scaleY = scaleX
		c.W = option.Width
		c.H = c.H*scaleY + option.Padding[0] + option.Padding[2]

		// 应用偏移
		newCanvas := canvas.New(c.W, c.H)
		c.RenderViewTo(newCanvas, canvas.Matrix{
			{scaleX, 0, option.Padding[3]},
			{0, scaleY, option.Padding[0]},
		})
		c = newCanvas
	} else if option.Height > 0 {
		// 指定高度自适应宽度
		if availableHeight > 0 {
			scaleY = availableHeight / c.H
		} else {
			scaleY = option.Height / c.H
		}
		scaleX = scaleY
		c.H = option.Height
		c.W = c.W*scaleX + option.Padding[1] + option.Padding[3]

		// 应用偏移
		newCanvas := canvas.New(c.W, c.H)
		c.RenderViewTo(newCanvas, canvas.Matrix{
			{scaleX, 0, option.Padding[3]},
			{0, scaleY, option.Padding[0]},
		})
		c = newCanvas
	} else {
		// 仅应用padding
		if option.Padding[0] > 0 || option.Padding[1] > 0 || option.Padding[2] > 0 || option.Padding[3] > 0 {
			newWidth := c.W + option.Padding[1] + option.Padding[3]
			newHeight := c.H + option.Padding[0] + option.Padding[2]
			newCanvas := canvas.New(newWidth, newHeight)
			c.RenderViewTo(newCanvas, canvas.Matrix{
				{1, 0, option.Padding[3]},
				{0, 1, option.Padding[0]},
			})
			c = newCanvas
		}
	}

	// 创建最终画布
	finalCanvas := canvas.New(c.W, c.H)
	ctx := canvas.NewContext(finalCanvas)
	ctx.SetCoordSystem(canvas.CartesianIV)

	// 绘制矩形
	for _, rectOption := range option.RectOption {
		DrawRect(ctx, rectOption)
	}

	// 将内容画布应用到最终画布
	c.RenderViewTo(finalCanvas, canvas.Identity)

	// 绘制额外的文本
	for _, extOption := range option.ExtraText {
		DrawExtraText(ctx, extOption)
	}
	// 根据参数设置翻转
	finalCanvas = ReverseCanvas(finalCanvas, option.ReverseX, option.ReverseY)

	return finalCanvas, nil
}

func GroupSvg(c *canvas.Canvas, output string) *canvas.Canvas {
	var stringWriter bytes.Buffer
	c.Write(&stringWriter, renderers.SVG())
	svg := stringWriter.String()
	svg = strings.Replace(svg, `xlink">`, `xlink"><g>`, -1)
	svg = strings.Replace(svg, `</svg>`, `</g></svg>`, -1)
	os.WriteFile(output, []byte(svg), 0644)
	return c
}
