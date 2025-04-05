package text2svg

import (
	"fmt"
	"math"
	"os"

	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
)

// AlignmentMode 定义对齐模式
type AlignmentMode int

const (
	// AlignLeft 左对齐
	AlignLeft AlignmentMode = iota
	// AlignCenter 居中对齐
	AlignCenter
	// AlignRight 右对齐
	AlignRight
)

// MultiLineOptions 多行文本配置选项
type MultiLineOptions struct {
	LineSpacing     float64         // 行间距, 默认为2
	Alignment       AlignmentMode   // 对齐方式: 左对齐、居中对齐、右对齐
	Width           float64         // 固定宽度
	Height          float64         // 固定高度
	Padding         []float64       // 内边距 [上, 右, 下, 左]
	MarginPadding   bool            // 是否启用外边距
	EnableBorder    bool            // 是否启用边框
	BorderColor     string          // 边框颜色
	BorderWidth     float64         // 边框宽度
	BorderRadius    float64         // 边框圆角半径
	BackgroundColor string          // 背景颜色
	SavePath        string          // 保存路径
	DPI             float64         // 保存DPI
	DPMM            float64         // 保存DPMM
	Quality         int             // 保存质量
	MirrorX         bool            // X轴镜像
	MirrorY         bool            // Y轴镜像
	ExtraTexts      []ExtraTextInfo // 额外的文本信息列表
}

// CanvasConvertMultipeLine 处理多行文本
// files: SVG文件列表
// options: 多行文本配置
// baseOptions: 基础配置选项(用于每个单独文本的样式设置)
func CanvasConvertMultipeLine(files []string, options *MultiLineOptions) (c *canvas.Canvas, err error) {
	// 加载SVG文件
	var canvases []*canvas.Canvas
	var widths []float64
	var heights []float64
	var maxWidth float64

	// 读取所有文件并解析为Canvas对象
	for _, file := range files {
		reader, err := os.Open(file)
		if err != nil {
			continue
		}
		defer reader.Close()

		c, err := canvas.ParseSVG(reader)
		if err != nil {
			reader.Close()
			continue
		}

		// 获取Canvas宽高
		w, h := c.Size()

		// 更新最大宽度
		if w > maxWidth {
			maxWidth = w
		}

		canvases = append(canvases, c)
		widths = append(widths, w)
		heights = append(heights, h)

		reader.Close()
	}

	// 如果没有有效文件，返回错误
	if len(canvases) == 0 {
		return nil, fmt.Errorf("没有可用的SVG文件")
	}

	// 设置行间距，如果未指定则使用默认值2
	lineSpacing := 2.0
	if options != nil && options.LineSpacing > 0 {
		lineSpacing = options.LineSpacing
	}

	// 计算总高度（包括行间距）
	totalHeight := 0.0
	for i, h := range heights {
		totalHeight += h
		// 除最后一行外，每行添加行间距
		if i < len(heights)-1 {
			totalHeight += lineSpacing
		}
	}

	contentWidth := maxWidth
	contentHeight := totalHeight

	// 计算最终画布宽度和高度
	finalWidth := contentWidth
	finalHeight := contentHeight

	// 创建最终画布
	finalCanvas := canvas.New(finalWidth, finalHeight)

	// 绘制每一行文本
	yPos := 0.0 // 初始Y位置（从上边距开始）

	// 将canvases数组翻转顺序，以便文本从上到下显示
	reversedCanvases := make([]*canvas.Canvas, len(canvases))
	reversedWidths := make([]float64, len(widths))
	reversedHeights := make([]float64, len(heights))

	for i := 0; i < len(canvases); i++ {
		reversedCanvases[i] = canvases[len(canvases)-1-i]
		reversedWidths[i] = widths[len(widths)-1-i]
		reversedHeights[i] = heights[len(heights)-1-i]
	}

	for i, c := range reversedCanvases {
		// 根据对齐方式计算X位置
		var xPos float64

		if options != nil {
			switch options.Alignment {
			case AlignCenter:
				// 居中对齐
				xPos = (contentWidth - reversedWidths[i]) / 2
			case AlignRight:
				// 右对齐
				xPos = contentWidth - reversedWidths[i]
			default:
				// 默认左对齐
				xPos = 0.0
			}
		} else {
			// 默认左对齐
			xPos = 0.0
		}

		// 在最终画布上绘制当前行
		// 使用变换矩阵定位当前Canvas的内容到最终Canvas上的正确位置
		transformMatrix := canvas.Identity.Translate(xPos, yPos)
		c.RenderViewTo(finalCanvas, transformMatrix)

		// 更新Y位置，为下一行做准备
		yPos += reversedHeights[i] + lineSpacing
	}

	var newWidth float64
	var newHeight float64

	// 处理固定宽度和高度（优先级高于最大宽度/高度）
	if options != nil {

		// 如果同时指定了宽度和高度，且内容需要缩放
		if options.Width > 0 && options.Height > 0 {
			// 计算内容的原始宽高比和目标区域的宽高比
			contentRatio := maxWidth / totalHeight
			targetRatio := contentWidth / contentHeight
			if options.Width == options.Height {
				newWidth = options.Width
				newHeight = options.Height
			} else {
				// 如果内容比例不匹配目标区域，需要缩放
				if math.Abs(contentRatio-targetRatio) > 0.001 {
					// 确定缩放因子（保持宽高比）
					scaleX := contentWidth / maxWidth
					scaleY := contentHeight / totalHeight
					scaleFactor := math.Min(scaleX, scaleY)

					// 重新计算所有宽度和高度
					for i := range widths {
						widths[i] *= scaleFactor
					}

					for i := range heights {
						heights[i] *= scaleFactor
					}

					// 更新最大宽度和总高度
					maxWidth *= scaleFactor
					totalHeight = 0.0
					for i, h := range heights {
						totalHeight += h
						if i < len(heights)-1 {
							totalHeight += lineSpacing
						}
					}
				}
			}
		} else {
			if options.Width == 0 && options.Height == 0 {
				newWidth = maxWidth
				newHeight = totalHeight
			}
			// 如果指定了固定宽度
			if options.Width > 0 {
				newWidth = options.Width
				contentWidth = newWidth
			}

			// 如果指定了固定高度
			if options.Height > 0 {
				newHeight = options.Height
				contentHeight = newHeight
			}
		}
	}

	// 应用内边距
	var padding []float64
	if options != nil && len(options.Padding) > 0 {
		padding = processPadding(options.Padding)
	} else {
		padding = []float64{0, 0, 0, 0} // 默认内边距为0
	}
	scale := 1.0
	cw := 0.0
	ch := 0.0
	offsetX := 0.0
	offsetY := 0.0
	if newWidth > 0 && newHeight == 0 {
		scale = (newWidth - padding[1] - padding[3]) / finalWidth
		cw = newWidth
		ch = finalHeight*scale + padding[0] + padding[2]
		offsetX = padding[3]
		offsetY = padding[0]
	}
	if newWidth == 0 && newHeight > 0 {
		scale = (newHeight - padding[0] - padding[2]) / finalHeight
		cw = finalWidth*scale + padding[1] + padding[3]
		ch = newHeight
		offsetY = padding[0]
		offsetX = padding[3]
	}
	if newWidth > 0 && newHeight > 0 {
		s1 := (newWidth - padding[1] - padding[3]) / finalWidth
		s2 := (newHeight - padding[0] - padding[2]) / finalHeight
		scale = math.Min(s1, s2)
		cw = newWidth
		ch = newHeight
		// 计算水平和垂直方向的偏移量，考虑内边距
		if s1 > s2 {
			// 当宽度约束更宽松时，水平居中并考虑左右内边距
			offsetX = padding[3] + (newWidth-padding[1]-padding[3]-finalWidth*scale)/2
			offsetY = padding[0] // 顶部内边距
		} else {
			// 当高度约束更宽松时，垂直居中并考虑上下内边距
			offsetX = padding[3] // 左侧内边距
			offsetY = padding[0] + (newHeight-padding[0]-padding[2]-finalHeight*scale)/2
		}
		if options.MarginPadding {
			s1 = (newWidth) / finalWidth
			s2 = (newHeight) / finalHeight
			scale = math.Min(s1, s2)
			cw = newWidth + padding[1] + padding[3]
			ch = newHeight + padding[0] + padding[2]
			// 计算水平和垂直方向的偏移量，考虑内边距
			if s1 > s2 {
				// 当宽度约束更宽松时，水平居中并考虑左右内边距
				offsetX = padding[3] + (newWidth-finalWidth*scale)/2
				offsetY = padding[0] // 顶部内边距
			} else {
				// 当高度约束更宽松时，垂直居中并考虑上下内边距
				offsetX = padding[3] // 左侧内边距
				offsetY = padding[0] + (newHeight-finalHeight*scale)/2
			}
		}
	}
	newCanvas := canvas.New(cw, ch)

	// 如果启用背景，绘制背景
	if options != nil && options.EnableBorder {
		// 创建背景上下文
		bgCtx := canvas.NewContext(newCanvas)
		if options.BackgroundColor != "" {
			// 设置填充颜色
			bgCtx.SetFillColor(canvas.Hex(options.BackgroundColor))
		} else {
			bgCtx.SetFillColor(canvas.RGBA(255, 255, 255, 0))
		}

		// 绘制矩形路径
		var bgPath *canvas.Path
		if options.BorderRadius > 0 {
			// 创建圆角矩形路径
			bgPath = &canvas.Path{}
			r := options.BorderRadius

			// 路径起点（左上角圆弧起点）
			bgPath.MoveTo(r, 0)

			// 上边线
			bgPath.LineTo(newWidth-r, 0)

			// 右上角圆弧
			bgPath.ArcTo(r, r, 0, false, true, newWidth, r)

			// 右边线
			bgPath.LineTo(newWidth, newHeight-r)

			// 右下角圆弧
			bgPath.ArcTo(r, r, 0, false, true, newWidth-r, newHeight)

			// 下边线
			bgPath.LineTo(r, newHeight)

			// 左下角圆弧
			bgPath.ArcTo(r, r, 0, false, true, 0, newHeight-r)

			// 左边线
			bgPath.LineTo(0, r)

			// 左上角圆弧
			bgPath.ArcTo(r, r, 0, false, true, r, 0)

			// 闭合路径
			bgPath.Close()
		} else {
			bgPath = canvas.Rectangle(cw, ch)
		}

		// 如果有边框，设置描边属性
		if options.EnableBorder && options.BorderColor != "" && options.BorderWidth > 0 {
			bgCtx.SetStrokeColor(canvas.Hex(options.BorderColor))
			bgCtx.SetStrokeWidth(options.BorderWidth)
			bgCtx.DrawPath(0, 0, bgPath)
			bgCtx.FillStroke()
		} else {
			bgCtx.DrawPath(0, 0, bgPath)
			bgCtx.Fill()
		}
	}
	finalCanvas.RenderViewTo(newCanvas, canvas.Matrix{
		{scale, 0, offsetX},
		{0, scale, offsetY},
	})

	// 应用镜像变换（如果启用）
	if options != nil && (options.MirrorX || options.MirrorY) {
		// 创建镜像画布
		mirrorCanvas := canvas.New(cw, ch)

		// 计算镜像变换矩阵
		var mirrorMatrix canvas.Matrix
		if options.MirrorX && options.MirrorY {
			// X和Y轴都镜像
			mirrorMatrix = canvas.Matrix{
				{-1, 0, cw},
				{0, -1, ch},
			}
		} else if options.MirrorX {
			// 只镜像X轴
			mirrorMatrix = canvas.Matrix{
				{-1, 0, cw},
				{0, 1, 0},
			}
		} else {
			// 只镜像Y轴
			mirrorMatrix = canvas.Matrix{
				{1, 0, 0},
				{0, -1, ch},
			}
		}

		// 将newCanvas渲染到镜像画布上
		newCanvas.RenderViewTo(mirrorCanvas, mirrorMatrix)

		// 用镜像画布替换原画布
		newCanvas = mirrorCanvas
	}

	// 绘制额外的文本（如果有）
	if options != nil && len(options.ExtraTexts) > 0 {
		drawExtraTexts(newCanvas, cw, ch, Options{
			ExtraTexts: options.ExtraTexts,
		})
	}

	// 如果设置了保存路径，保存画布
	if options != nil && options.SavePath != "" {
		format := file.ExtName(options.SavePath)
		// 创建保存配置
		config := SaveConfig{
			Format:  SaveFormat(format),
			Path:    options.SavePath,
			DPI:     options.DPI,
			DPMM:    options.DPMM,
			Quality: options.Quality,
		}

		// 保存到文件
		if config.Format == FormatSVG {
			// 为SVG保存创建一个临时Options对象（因为handleSVGSave需要Options参数）
			tempOptions := &Options{
				EnableBackground:      options.EnableBorder,
				BackgroundColor:       options.BackgroundColor,
				BackgroundStroke:      options.BorderColor,
				BackgroundStrokeWidth: options.BorderWidth,
				BorderRadius:          options.BorderRadius,
			}
			return handleSVGSave(newCanvas, tempOptions, config)
		}

		return saveToFile(newCanvas, config)
	}

	return newCanvas, nil
}
