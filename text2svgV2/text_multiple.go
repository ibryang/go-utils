package text2svgV2

import (
	"errors"
	"math"

	"github.com/tdewolff/canvas"
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

	if option.Width == 0 && option.Height == 0 {
		// 保持原始尺寸
	} else if option.Width > 0 && option.Height > 0 {
		// 指定宽高
		scaleX = option.Width / c.W
		scaleY = option.Height / c.H
		c.W = option.Width
		c.H = option.Height
	} else if option.Width > 0 {
		// 指定宽度自适应高度
		scaleX = option.Width / c.W
		scaleY = scaleX
		c.W = option.Width
		c.H = c.H * scaleY
	} else if option.Height > 0 {
		// 指定高度自适应宽度
		scaleY = option.Height / c.H
		scaleX = scaleY
		c.H = option.Height
		c.W = c.W * scaleX
	}

	c.Transform(canvas.Matrix{
		{scaleX, 0, 0},
		{0, scaleY, 0},
	})

	// 创建最终画布
	finalCanvas := canvas.New(c.W, c.H)
	ctx := canvas.NewContext(finalCanvas)
	ctx.SetCoordSystem(canvas.CartesianIV)

	// 绘制矩形
	for _, rectOption := range option.RectOption {
		DrawRect(ctx, rectOption)
	}

	// 应用内边距
	scaleX = (c.W - option.Padding[1] - option.Padding[3]) / finalCanvas.W
	scaleY = (c.H - option.Padding[0] - option.Padding[2]) / finalCanvas.H

	// 将内容画布应用到最终画布
	c.RenderViewTo(finalCanvas, canvas.Matrix{
		{scaleX, 0, option.Padding[3]},
		{0, scaleY, option.Padding[0]},
	})

	// 绘制额外的文本
	for _, extOption := range option.ExtraText {
		DrawExtraText(ctx, extOption)
	}
	// 根据参数设置翻转
	finalCanvas = ReverseCanvas(finalCanvas, option.ReverseX, option.ReverseY)

	return finalCanvas, nil
}
