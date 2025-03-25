package text2svg

import (
	"fmt"
	"math"

	"github.com/tdewolff/canvas"
)

// validateOptions 验证选项并设置默认值
func validateOptions(options *Options) error {
	if options.Text == "" {
		return fmt.Errorf("文本内容不能为空")
	}

	// 设置默认颜色
	if len(options.Colors) == 0 {
		options.Colors = []string{"#000000"}
	}

	// 处理none颜色（透明）
	for i := 0; i < len(options.Colors); i++ {
		if options.Colors[i] == "none" {
			options.Colors[i] = "#00000000"
		}
	}

	// 设置默认描边属性
	if options.EnableStroke && options.StrokeWidth <= 0 {
		options.StrokeWidth = 1.0
	}
	if options.EnableStroke && options.StrokeColor == "" {
		options.StrokeColor = "#000000"
	}

	// 设置默认背景属性
	if options.EnableBackground && options.BackgroundColor == "" {
		options.BackgroundColor = "#FFFFFF"
	}
	if options.BackgroundStroke != "" && options.BackgroundStrokeWidth <= 0 {
		options.BackgroundStrokeWidth = 1.0
	}

	// 处理内边距
	options.Padding = processPadding(options.Padding)

	return nil
}

// generateCanvasInternal 生成画布的内部实现
func generateCanvasInternal(options Options) (*canvas.Canvas, error) {
	// 加载字体
	font, err := loadFontFamily(options.FontPath)
	if err != nil {
		return nil, fmt.Errorf("加载字体失败: %v", err)
	}

	face := font.Face(options.FontSize, nil)

	var totalWidth float64
	var maxHeight float64
	var minY float64
	var maxY float64
	var paths []*canvas.Path
	var xOffsets []float64
	var bounds []canvas.Rect
	var colorIndices []int

	if options.RenderMode == RenderModeString {
		// 整体字符串路径模式
		path, _, err := face.ToPath(options.Text)
		if err != nil {
			return nil, fmt.Errorf("转换文本到路径失败: %v", err)
		}

		if path == nil {
			return nil, fmt.Errorf("生成路径失败")
		}

		path = path.Transform(canvas.Matrix{
			{1, 0, -path.Bounds().X0},
			{0, 1, 0},
		})
		pathBounds := path.Bounds()
		totalWidth = pathBounds.W()
		minY = pathBounds.Y0
		maxY = pathBounds.Y1

		paths = []*canvas.Path{path}
		bounds = []canvas.Rect{pathBounds}
		xOffsets = []float64{0}
		colorIndices = []int{0}
	} else {
		// 单字符路径模式
		runes := []rune(options.Text)
		colorCount := 0
		for i, char := range runes {
			path, advance, err := face.ToPath(string(char))
			if err != nil {
				return nil, fmt.Errorf("转换文本到路径失败: %v", err)
			}

			if char == ' ' {
				colorIndices = append(colorIndices, -1)
				totalWidth += advance
				continue
			}

			if path == nil {
				continue
			}

			pathBounds := path.Bounds()
			bounds = append(bounds, pathBounds)
			paths = append(paths, path)
			xOffsets = append(xOffsets, totalWidth)
			colorIndices = append(colorIndices, colorCount%len(options.Colors))
			colorCount++

			if len(paths) == 1 {
				minY = pathBounds.Y0
				maxY = pathBounds.Y1
			} else {
				if pathBounds.Y0 < minY {
					minY = pathBounds.Y0
				}
				if pathBounds.Y1 > maxY {
					maxY = pathBounds.Y1
				}
			}

			if i < len(runes)-1 {
				totalWidth += advance
			} else {
				totalWidth += pathBounds.W()
			}
		}
	}

	maxHeight = maxY - minY

	// 计算内边距的影响
	contentWidth := totalWidth
	contentHeight := maxHeight

	// 考虑描边宽度
	if options.EnableStroke {
		contentWidth += options.StrokeWidth * 2
		contentHeight += options.StrokeWidth * 2
	}

	// 处理LockWidth和LockHeight（动态调整padding）
	if options.LockWidth > 0 || options.LockHeight > 0 {
		// 计算当前内容加上当前padding后的尺寸
		currentTotalWidth := contentWidth + options.Padding[1] + options.Padding[3]
		currentTotalHeight := contentHeight + options.Padding[0] + options.Padding[2]

		// 如果设置了锁定宽度，并且当前宽度与锁定宽度不同，调整水平内边距
		if options.LockWidth > 0 && currentTotalWidth != options.LockWidth {
			// 计算所需的额外内边距
			extraPaddingTotal := options.LockWidth - contentWidth

			// 确保结果不小于0
			if extraPaddingTotal < 0 {
				extraPaddingTotal = 0
			}

			// 平均分配到左右内边距
			extraPadding := extraPaddingTotal / 2
			options.Padding[1] = extraPadding // 右内边距
			options.Padding[3] = extraPadding // 左内边距
		}

		// 如果设置了锁定高度，并且当前高度与锁定高度不同，调整垂直内边距
		if options.LockHeight > 0 && currentTotalHeight != options.LockHeight {
			// 计算所需的额外内边距
			extraPaddingTotal := options.LockHeight - contentHeight

			// 确保结果不小于0
			if extraPaddingTotal < 0 {
				extraPaddingTotal = 0
			}

			// 平均分配到上下内边距
			extraPadding := extraPaddingTotal / 2
			options.Padding[0] = extraPadding // 上内边距
			options.Padding[2] = extraPadding // 下内边距
		}
	}

	// 应用更新后的内边距计算最终尺寸
	totalWidth = contentWidth + options.Padding[1] + options.Padding[3]
	maxHeight = contentHeight + options.Padding[0] + options.Padding[2]

	// 确定最终尺寸（考虑Width/Height属性）
	var width, height, scaleX, scaleY float64
	if options.LockWidth > 0 || options.LockHeight > 0 {
		// 如果设置了锁定尺寸，优先使用锁定尺寸
		if options.LockWidth > 0 {
			width = options.LockWidth
		} else {
			width = totalWidth
		}

		if options.LockHeight > 0 {
			height = options.LockHeight
		} else {
			height = maxHeight
		}

		// 计算缩放比例（为了保持内容的原始比例）
		scaleX = width / totalWidth
		scaleY = height / maxHeight
	} else {
		// 否则使用传统的计算方式
		width, height, scaleX, scaleY = calculateDimensions(totalWidth, maxHeight, options.Width, options.Height)
	}

	// 创建最终画布
	c := canvas.New(width, height)

	// 如果需要背景，先绘制背景
	if options.EnableBackground {
		drawBackground(c, width, height, options)
	}

	// 绘制文本内容
	drawTextContent(c, width, height, paths, colorIndices, bounds, xOffsets, minY, scaleX, scaleY, options)

	// 绘制额外的文本
	if len(options.ExtraTexts) > 0 {
		drawExtraTexts(c, width, height, options)
	}

	return c, nil
}

// drawBackground 绘制背景
func drawBackground(c *canvas.Canvas, width, height float64, options Options) {
	// 使用单独的Context绘制背景
	bgCtx := canvas.NewContext(c)

	// 设置填充颜色
	bgCtx.SetFillColor(canvas.Hex(options.BackgroundColor))

	// 绘制矩形路径
	var bgPath *canvas.Path
	if options.BorderRadius > 0 {
		// 不使用RoundedRectangle，改为手动创建路径以兼容CDR
		// 使用标准SVG路径命令：M(移动), L(线条), A(圆弧)
		bgPath = &canvas.Path{}
		r := options.BorderRadius

		// 路径起点（左上角圆弧起点）
		bgPath.MoveTo(r, 0)

		// 上边线
		bgPath.LineTo(width-r, 0)

		// 右上角圆弧
		bgPath.ArcTo(r, r, 0, false, true, width, r)

		// 右边线
		bgPath.LineTo(width, height-r)

		// 右下角圆弧
		bgPath.ArcTo(r, r, 0, false, true, width-r, height)

		// 下边线
		bgPath.LineTo(r, height)

		// 左下角圆弧
		bgPath.ArcTo(r, r, 0, false, true, 0, height-r)

		// 左边线
		bgPath.LineTo(0, r)

		// 左上角圆弧
		bgPath.ArcTo(r, r, 0, false, true, r, 0)

		// 闭合路径
		bgPath.Close()
	} else {
		bgPath = canvas.Rectangle(width, height)
	}

	// 如果有背景描边，设置描边属性
	if options.BackgroundStroke != "" {
		bgCtx.SetStrokeColor(canvas.Hex(options.BackgroundStroke))
		bgCtx.SetStrokeWidth(options.BackgroundStrokeWidth)
		bgCtx.DrawPath(0, 0, bgPath)
		bgCtx.FillStroke()
	} else {
		bgCtx.DrawPath(0, 0, bgPath)
		bgCtx.Fill()
	}
}

// drawTextContent 绘制文本内容
func drawTextContent(c *canvas.Canvas, width, height float64, paths []*canvas.Path,
	colorIndices []int, bounds []canvas.Rect, xOffsets []float64, minY float64,
	scaleX, scaleY float64, options Options) {

	// 创建文字上下文
	ctx := canvas.NewContext(c)

	// 确定文本需要的总宽度和总高度（用于居中计算）
	var contentWidth, contentHeight float64

	if options.RenderMode == RenderModeString {
		// 整体字符串模式使用整体尺寸
		contentWidth = paths[0].Bounds().W() * scaleX
		contentHeight = (bounds[0].Y1 - bounds[0].Y0) * scaleY
	} else if len(paths) > 0 {
		// 单字符模式计算总宽度
		if len(xOffsets) > 0 && len(paths) > 0 {
			lastIdx := len(paths) - 1
			contentWidth = (xOffsets[lastIdx] + paths[lastIdx].Bounds().W()) * scaleX

			// 计算内容高度 - 使用所有路径的最大高度
			var maxBoundsY float64
			for _, rect := range bounds {
				if rect.Y1 > maxBoundsY {
					maxBoundsY = rect.Y1
				}
			}
			contentHeight = (maxBoundsY - minY) * scaleY
		}
	}

	// 计算文本在画布上的位置（考虑居中和内边距）
	var baseX, baseY float64

	// 水平居中：(画布宽度 - 内容宽度) / 2
	if options.Padding[1] == options.Padding[3] && options.Padding[1] > 0 {
		// 如果左右内边距相等，说明已经通过内边距实现了居中
		baseX = options.Padding[3] * scaleX
	} else {
		// 否则需要手动计算居中位置
		baseX = (width - contentWidth) / 2
	}

	// 垂直居中：(画布高度 - 内容高度) / 2
	if options.Padding[0] == options.Padding[2] && options.Padding[0] > 0 {
		// 如果上下内边距相等，说明已经通过内边距实现了居中
		baseY = options.Padding[0] * scaleY
	} else {
		// 否则需要手动计算居中位置
		baseY = (height - contentHeight) / 2
	}

	// 如果启用了描边，需要考虑描边宽度
	if options.EnableStroke {
		baseX += options.StrokeWidth * scaleX
		baseY += options.StrokeWidth * scaleY
	}

	// 移动原点到基础位置
	ctx.Translate(baseX, baseY)

	// 应用缩放
	ctx.Scale(scaleX, scaleY)

	// 绘制每个字符并设置颜色
	pathIndex := 0
	for _, colorIndex := range colorIndices {
		if options.RenderMode == RenderModeString {
			// 整体字符串路径模式
			path := paths[0]

			// 保存当前上下文状态
			ctx.Push()

			// 如果启用描边，设置描边属性
			if options.EnableStroke && options.StrokeWidth > 0 {
				ctx.SetStrokeWidth(options.StrokeWidth)
				ctx.SetStrokeColor(canvas.Hex(options.StrokeColor))
				// 设置填充颜色
				ctx.SetFillColor(canvas.Hex(options.Colors[0]))
				// 绘制路径 - 调整Y坐标，确保基线位置正确
				ctx.DrawPath(0, -minY, path)
				// 同时填充和描边
				ctx.FillStroke()
			} else {
				// 只填充文字
				ctx.SetFillColor(canvas.Hex(options.Colors[0]))
				ctx.DrawPath(0, -minY, path)
				ctx.Fill()
			}

			// 恢复上下文状态
			ctx.Pop()
			break // 只需要绘制一次
		} else {
			if colorIndex == -1 { // 跳过空格
				continue
			}

			path := paths[pathIndex]

			// 检查路径是否有效（非空）
			if path == nil {
				pathIndex++
				continue
			}

			// 检查路径是否为空（通过检查其边界来判断）
			pathBounds := path.Bounds()
			if pathBounds.W() <= 0 || pathBounds.H() <= 0 {
				pathIndex++
				continue
			}

			// 计算当前字符的位置
			charX := xOffsets[pathIndex] - bounds[pathIndex].X0
			charY := -minY // 调整Y坐标，使基线位置一致

			// 保存当前上下文状态
			ctx.Push()

			// 如果启用描边，设置描边属性
			if options.EnableStroke && options.StrokeWidth > 0 {
				ctx.SetStrokeWidth(options.StrokeWidth)
				ctx.SetStrokeColor(canvas.Hex(options.StrokeColor))
				// 设置填充颜色
				ctx.SetFillColor(canvas.Hex(options.Colors[colorIndex]))
				// 绘制路径
				ctx.DrawPath(charX, charY, path)
				// 同时填充和描边
				ctx.FillStroke()
			} else {
				// 只填充文字
				ctx.SetFillColor(canvas.Hex(options.Colors[colorIndex]))
				ctx.DrawPath(charX, charY, path)
				ctx.Fill()
			}

			// 恢复上下文状态
			ctx.Pop()

			pathIndex++
		}
	}
}

// drawExtraTexts 绘制额外的文本
func drawExtraTexts(c *canvas.Canvas, width, height float64, options Options) {
	for _, extraText := range options.ExtraTexts {
		if extraText.Text == "" {
			continue // 跳过空文本
		}

		// 确定字体
		extraFontPath := extraText.FontPath
		if extraFontPath == "" {
			extraFontPath = options.FontPath // 使用主文本字体
		}

		// 加载字体
		extraFont, err := loadFontFamily(extraFontPath)
		if err != nil {
			continue // 跳过加载失败的字体
		}

		// 确定字体大小
		extraFontSize := extraText.FontSize
		if extraFontSize <= 0 {
			extraFontSize = options.FontSize // 使用主文本字体大小
		}

		// 创建字体Face
		extraFace := extraFont.Face(extraFontSize, nil)

		// 创建文本路径
		extraPath, _, err := extraFace.ToPath(extraText.Text)
		if err != nil || extraPath == nil {
			continue // 跳过转换失败或无效路径
		}

		// 设置透明度
		opacity := extraText.Opacity
		if opacity <= 0 || opacity > 1 {
			opacity = 1.0 // 默认不透明
		}

		// 获取路径边界
		extraBounds := extraPath.Bounds()

		// 计算实际渲染位置 - X坐标从左侧开始，Y坐标从底部开始
		textX := extraText.X - extraBounds.X0

		// 转换Y坐标：Y=0在底部，Y=height在顶部
		// 首先，从底部将坐标转换为从顶部的坐标
		convertedY := height - extraText.Y
		// 然后，像之前一样调整为左上角基准点
		textY := convertedY - extraBounds.Y0

		// 应用偏移
		textX += extraText.OffsetX
		textY += extraText.OffsetY

		// 创建上下文
		extraCtx := canvas.NewContext(c)

		// 应用变换
		extraCtx.Translate(textX, textY)

		// 如果有旋转，应用旋转变换
		if extraText.Rotate != 0 {
			// 计算旋转中心（文本中心点）
			centerX := extraBounds.W() / 2
			centerY := extraBounds.H() / 2

			// 移动到旋转中心点
			extraCtx.Translate(centerX, centerY)
			// 旋转（角度转弧度）
			extraCtx.Rotate(extraText.Rotate * math.Pi / 180)
			// 移回原位置
			extraCtx.Translate(-centerX, -centerY)
		}

		// 设置颜色
		textColor := extraText.Color
		if textColor == "" {
			textColor = "#000000" // 默认黑色
		}

		// 如果需要描边，设置描边属性
		if extraText.StrokeText && extraText.StrokeWidth > 0 {
			extraCtx.SetStrokeWidth(extraText.StrokeWidth)

			strokeColor := extraText.StrokeColor
			if strokeColor == "" {
				strokeColor = "#000000" // 默认黑色描边
			}

			extraCtx.SetStrokeColor(canvas.Hex(strokeColor))
			extraCtx.SetFillColor(canvas.Hex(textColor))

			// 绘制路径 - 使用原点(0,0)，已经通过Translate调整了位置
			extraCtx.DrawPath(0, 0, extraPath)

			// 同时填充和描边
			extraCtx.FillStroke()
		} else {
			// 只填充文字
			extraCtx.SetFillColor(canvas.Hex(textColor))

			// 绘制路径 - 使用原点(0,0)，已经通过Translate调整了位置
			extraCtx.DrawPath(0, 0, extraPath)

			extraCtx.Fill()
		}
	}
}
