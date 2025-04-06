package example_test

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

func TestFont(t *testing.T) {
	text := "Benjamin"
	font, _ := canvas.LoadSystemFont("Cookie", canvas.FontBlack)
	fontface := font.Face(100, canvas.Black)

	// 首先计算所有字符的确切边界，以确定整个字符串的实际可视范围
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	// 临时保存每个字符的路径和前进宽度
	charPaths := make([]canvas.Path, 0, len(text))
	advances := make([]float64, 0, len(text))

	// 计算整个字符串的确切边界框
	xPos := 0.0
	for _, char := range text {
		path, advance, err := fontface.ToPath(string(char))
		if err != nil {
			t.Fatalf("转换文本到路径失败: %v", err)
		}

		bounds := path.Bounds()
		minX = math.Min(minX, bounds.X0+xPos)
		minY = math.Min(minY, bounds.Y0)
		maxX = math.Max(maxX, bounds.X1+xPos)
		maxY = math.Max(maxY, bounds.Y1)

		charPaths = append(charPaths, *path)
		advances = append(advances, advance)
		xPos += advance
	}

	// 计算精确的宽度和高度
	exactWidth := maxX - minX
	exactHeight := maxY - minY

	fmt.Printf("精确边界: X:%.2f Y:%.2f W:%.2f H:%.2f\n", minX, minY, exactWidth, exactHeight)

	// 创建一个尺寸刚好容纳所有字符的画布
	textCanvas := canvas.New(exactWidth, exactHeight)
	textCtx := canvas.NewContext(textCanvas)

	// 绘制每个字符
	xPos = -minX // 调整起始位置，确保所有内容都可见
	yPos := -minY

	for i, path := range charPaths {
		// 将路径绘制到画布上
		// ctx.SetFillColor(canvas.Transparent)
		textCtx.SetStrokeColor(canvas.Black)
		textCtx.SetStrokeWidth(0.1)
		textCtx.Stroke()
		textCtx.DrawPath(xPos, yPos, &path)
		textCtx.Fill()

		// 更新x位置
		xPos += advances[i]
	}

	padding := [4]float64{} // 上，右，下，左
	bgColor := "#FFFF00"
	strokeColor := "#0000FF"
	strokeWidth := 0.1

	// 固定宽高: 如果宽度>0,则高度自适应, 如果高度>0,则宽度自适应,如果宽高都大于0,则强制设置宽高
	canvasWidth := .0
	canvasHeight := .0

	// 计算文本的缩放比例
	textScaleX := 1.0
	textScaleY := 1.0
	// 如果宽高都为0,则设置宽高为字符串的宽高+padding
	if canvasWidth == 0 && canvasHeight == 0 {
		canvasWidth = exactWidth + padding[1] + padding[3]
		canvasHeight = exactHeight + padding[0] + padding[2]
	} else if canvasWidth > 0 && canvasHeight > 0 {
		// 如果宽高都大于0,则强制设置宽高
		textScaleX = (canvasWidth - padding[1] - padding[3]) / exactWidth
		textScaleY = (canvasHeight - padding[0] - padding[2]) / exactHeight
	} else if canvasWidth == 0 {
		// 如果宽度为0,则宽度自适应
		textMaxHeight := canvasHeight - padding[0] - padding[2]
		textScaleY = textMaxHeight / exactHeight
		textScaleX = textScaleY
		canvasWidth = textScaleY*exactWidth + padding[1] + padding[3]
	} else if canvasHeight == 0 {
		// 如果高度为0,则高度自适应
		textMaxWidth := canvasWidth - padding[1] - padding[3]
		textScaleX = textMaxWidth / exactWidth
		textScaleY = textScaleX
		canvasHeight = textScaleX*exactHeight + padding[0] + padding[2]
	}

	fmt.Println("canvasWidth", canvasWidth, "canvasHeight", canvasHeight)

	textMatrix := canvas.Matrix{
		{textScaleX, 0, padding[3]},
		{0, textScaleY, padding[2]},
	}

	c2 := canvas.New(canvasWidth, canvasHeight)

	ctx2 := canvas.NewContext(c2)

	// 背景
	// 描边
	ctx2.SetStrokeColor(canvas.Hex(strokeColor))
	ctx2.SetStrokeWidth(strokeWidth)
	ctx2.Stroke()
	// 圆角
	// radius := 10.0
	ctx2.SetFillColor(canvas.Hex(bgColor))
	ctx2.DrawPath(0, 0, canvas.Rectangle(c2.W, c2.H))
	ctx2.Fill()

	// 绘制文本
	// 动态计算文本的宽度
	textCanvas.RenderViewTo(c2, textMatrix)

	// fmt.Println(c.W, c.H)
	// 保存为SVG文件
	renderers.Write("font_mapping.svg", c2)
}

type TextOption struct {
	Text      string  // 文本
	FontPath  string  // 字体路径
	FontSize  float64 // 字体大小
	FontColor any     // 字体颜色
	BaseOption
	RectOption *RectOption
}

var ColorMap = map[string]color.RGBA{
	"transparent": {0, 0, 0, 0},
	"none":        {0, 0, 0, 0},
	"black":       {0, 0, 0, 255},
	"white":       {255, 255, 255, 255},
	"red":         {255, 0, 0, 255},
	"green":       {0, 255, 0, 255},
	"blue":        {0, 0, 255, 255},
	"yellow":      {255, 255, 0, 255},
	"orange":      {255, 165, 0, 255},
	"purple":      {128, 0, 128, 255},
	"pink":        {255, 192, 203, 255},
	"brown":       {139, 69, 19, 255},
	"gray":        {128, 128, 128, 255},
	"cyan":        {0, 255, 255, 255},
}

// 生成文本Svg
func GenerateTextSvg(option TextOption) (*canvas.Canvas, error) {
	if option.Text == "" {
		return nil, errors.New("text is required")
	}
	// 判断fontColor类型
	var fontColor []color.RGBA
	var strokeColor color.RGBA = canvas.Black
	var strokeWidth float64 = 0
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

	font, _ := canvas.LoadSystemFont(option.FontPath, canvas.FontBlack)
	fontface := font.Face(option.FontSize, option.FontColor)

	// 首先计算所有字符的确切边界，以确定整个字符串的实际可视范围
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	// 临时保存每个字符的路径和前进宽度
	charPaths := make([]canvas.Path, 0, len(option.Text))
	advances := make([]float64, 0, len(option.Text))

	// 计算整个字符串的确切边界框
	xPos := 0.0
	colorCount := 0
	var colorIndices []int
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
	exactWidth := maxX - minX
	exactHeight := maxY - minY

	// 创建一个尺寸刚好容纳所有字符的画布
	textCanvas := canvas.New(exactWidth, exactHeight)
	textCtx := canvas.NewContext(textCanvas)

	if option.RectOption != nil {
		bgColor := canvas.Transparent
		if color, ok := ColorMap[option.RectOption.BgColor]; ok {
			bgColor = color
		} else {
			bgColor = canvas.Hex(option.RectOption.BgColor)
		}
		textCtx.SetFillColor(bgColor)
		textCtx.DrawPath(0, 0, canvas.Rectangle(textCanvas.W, textCanvas.H))
		textCtx.Fill()
	}

	// 绘制每个字符
	xPos = -minX // 调整起始位置，确保所有内容都可见
	yPos := -minY

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
	// 根据参数设置翻转
	scaleX := 1.0
	scaleY := 1.0
	offsetX := 0.0
	offsetY := 0.0
	if option.ReversX {
		scaleX = -1.0
		offsetX = exactWidth
	}
	if option.ReversY {
		scaleY = -1.0
		offsetY = exactHeight
	}

	textCanvas.Transform(canvas.Matrix{
		{scaleX, 0, offsetX},
		{0, scaleY, offsetY},
	})

	if option.Width == 0 && option.Height == 0 {
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

func TestGenerateTextSvg(t *testing.T) {
	textOption := TextOption{
		Text:      "Hello World",
		FontPath:  "Cookie",
		FontSize:  48,
		FontColor: "blue",
		BaseOption: BaseOption{
			Width: 300,
		},
		RectOption: &RectOption{
			BgColor:     "yellow",
			StrokeColor: "red",
			StrokeWidth: .1,
			Radius:      2,
		},
	}
	canvas, err := GenerateTextSvg(textOption)
	if err != nil {
		t.Fatalf("生成文本Svg失败: %v", err)
	}

	// 保存为SVG文件
	renderers.Write("1.pdf", canvas)
}

type RectOption struct {
	Width       float64
	Height      float64
	X           float64
	Y           float64
	Radius      float64
	BgColor     string
	StrokeColor string
	StrokeWidth float64
}

type BaseOption struct {
	Width   float64
	Height  float64
	ReversX bool
	ReversY bool
}

type CanvasItem struct {
	File   string
	Canvas *canvas.Canvas
	Width  float64
	Height float64
	X      float64
	Y      float64
}

type CanvasOption struct {
	FileList   []CanvasItem
	CanvasList []CanvasItem
	Padding    [4]float64
	BaseOption
}

type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

type TextLineOption struct {
	TextList []TextOption // 文本列表
	Padding  [4]float64   // 上，右，下，左
	LineGap  float64      // 行间距
	Align    TextAlign    // 对齐方式: left, center, right
	BaseOption
	RectOption []RectOption
}

func GenerateCanvas(option CanvasOption) (*canvas.Canvas, error) {
	c := canvas.New(option.Width, option.Height)
	ctx := canvas.NewContext(c)
	ctx.SetCoordSystem(canvas.CartesianIV)

	// 绘制文件
	for _, file := range option.FileList {
		svgFile, err := os.Open(file.File)
		if err != nil {
			return nil, err
		}
		svgCanvas, err := canvas.ParseSVG(svgFile)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		scaleX := 1.0
		scaleY := 1.0
		if file.Width == 0 && file.Height == 0 {
			scaleX = c.W / svgCanvas.W
			scaleY = c.H / svgCanvas.H
		} else if file.Width > 0 {
			scaleX = file.Width / svgCanvas.W
			scaleY = scaleX
		} else if file.Height > 0 {
			scaleY = file.Height / svgCanvas.H
			scaleX = scaleY
		}
		fmt.Println("scaleX", scaleX, "scaleY", scaleY)
		fmt.Println("file.X", file.X, "file.Y", file.Y)
		fmt.Println("c.W", c.W, "c.H", c.H)
		fmt.Println("file.Width", file.Width, "file.Height", file.Height)
		svgCanvas.RenderViewTo(c, canvas.Matrix{
			{scaleX, 0, file.X},
			{0, scaleY, c.H - file.Height - file.Y},
		})
	}

	return c, nil
}

func TestGenerateCanvas(t *testing.T) {
	option := CanvasOption{
		FileList: []CanvasItem{
			{
				File:  "font_gen_text.svg",
				Width: 100,
				X:     0,
				Y:     30,
			},
		},
		BaseOption: BaseOption{
			Width:  300,
			Height: 300,
		},
	}
	canvas, err := GenerateCanvas(option)
	if err != nil {
		t.Fatalf("生成画布失败: %v", err)
	}
	renderers.Write("font_gen_canvas.pdf", canvas)
}

func TestGenerateTextLine(t *testing.T) {
	option := TextLineOption{
		TextList: []TextOption{
			{
				Text:      "Beijing",
				FontPath:  "Cookie",
				FontSize:  100,
				FontColor: []string{"blue", "red", "green"},
			},
			{
				Text:      "Beijing",
				FontPath:  "Cookie",
				FontSize:  20,
				FontColor: []string{"blue", "red", "green"},
			},
			{
				Text:      "Beijing",
				FontPath:  "Cookie",
				FontSize:  16,
				FontColor: []string{"blue", "red", "green"},
			},
		},
		// Padding:    [4]float64{10, 10, 10, 10}, // 上、右、下、左
		LineGap:    1,               // 行间距
		Align:      TextAlignCenter, // 居中对齐
		BaseOption: BaseOption{
			// Width:       300,
			// Height:      100,
		},
		RectOption: []RectOption{
			{
				StrokeColor: "red",
				StrokeWidth: 0.1,
				Radius:      5,
			},
		},
	}

	canvas, err := GenerateTextLine(option)
	if err != nil {
		t.Fatalf("生成文本行失败: %v", err)
	}

	// 保存为PDF文件
	renderers.Write("font_gen_text_line.pdf", canvas)
}

func GenerateTextLine(option TextLineOption) (*canvas.Canvas, error) {
	if len(option.TextList) == 0 {
		return nil, errors.New("text list is required")
	}

	// 列表反转
	textList := []TextOption{}
	for i := len(option.TextList) - 1; i >= 0; i-- {
		textList = append(textList, option.TextList[i])
	}
	option.TextList = textList

	// 生成每行文本的canvas并计算最大宽度
	textCanvases := make([]*canvas.Canvas, 0, len(option.TextList))
	maxWidth := 0.0
	totalHeight := 0.0

	for _, textOption := range option.TextList {
		textCanvas, err := GenerateTextSvg(textOption)
		if err != nil {
			return nil, err
		}

		textCanvases = append(textCanvases, textCanvas)
		maxWidth = math.Max(maxWidth, textCanvas.W)
		totalHeight += textCanvas.H
	}

	// 添加行间距到总高度
	if len(option.TextList) > 1 {
		totalHeight += option.LineGap * float64(len(option.TextList)-1)
	}

	// 创建不包含padding的内容画布
	contentWidth := maxWidth
	contentHeight := totalHeight

	// 创建最终的画布（不含padding）
	c := canvas.New(contentWidth, contentHeight)

	// 放置每一行文本
	yPos := 0.0

	for _, textCanvas := range textCanvases {
		// 根据对齐方式计算x坐标
		xPos := 0.0 // 默认左对齐

		if option.Align == TextAlignCenter {
			xPos = (contentWidth - textCanvas.W) / 2
		} else if option.Align == TextAlignRight {
			xPos = contentWidth - textCanvas.W
		}

		// 渲染文本到画布
		textCanvas.RenderViewTo(c, canvas.Matrix{
			{1, 0, xPos},
			{0, 1, yPos},
		})

		// 更新y坐标为下一行
		yPos += textCanvas.H + option.LineGap
	}

	// 应用缩放
	scaleX := 1.0
	scaleY := 1.0
	if option.Width == 0 && option.Height == 0 {
		// 保持原始尺寸
	} else if option.Width > 0 && option.Height > 0 {
		// 强制宽高
		scaleX = option.Width / c.W
		scaleY = option.Height / c.H
		c.W = option.Width
		c.H = option.Height
	} else if option.Width > 0 {
		// 指定宽度，高度自适应
		scaleX = option.Width / c.W
		scaleY = scaleX
		c.W = option.Width
		c.H = c.H * scaleY
	} else if option.Height > 0 {
		// 指定高度，宽度自适应
		scaleY = option.Height / c.H
		scaleX = scaleY
		c.H = option.Height
		c.W = c.W * scaleX
	}

	c.Transform(canvas.Matrix{
		{scaleX, 0, 0},
		{0, scaleY, 0},
	})

	finalCanvas := canvas.New(c.W, c.H)
	ctx := canvas.NewContext(finalCanvas)
	ctx.SetCoordSystem(canvas.CartesianIV)
	for _, rectOption := range option.RectOption {
		if rectOption.Width == 0 {
			rectOption.Width = finalCanvas.W
		} else {
			if rectOption.Width < 0 {
				rectOption.Width = finalCanvas.W + rectOption.Width
			}
		}
		if rectOption.Height == 0 {
			rectOption.Height = finalCanvas.H
		} else {
			if rectOption.Height < 0 {
				rectOption.Height = finalCanvas.H + rectOption.Height
			}
		}
		bgColor := canvas.Transparent
		if color, ok := ColorMap[rectOption.BgColor]; ok {
			bgColor = color
		}
		ctx.SetFillColor(bgColor)
		if rectOption.Radius > 0 {
			// 绘制矩形
			ctx.DrawPath(rectOption.X, rectOption.Y, GenerateRoundedRectPath(rectOption.Width, rectOption.Height, rectOption.Radius))
		} else {
			ctx.DrawPath(rectOption.X, rectOption.Y, canvas.Rectangle(rectOption.Width, rectOption.Height))
		}
		if rectOption.StrokeWidth > 0 {
			var strokeColor color.RGBA
			if rectOption.StrokeColor != "" {
				if color, ok := ColorMap[rectOption.StrokeColor]; ok {
					strokeColor = color
				} else {
					strokeColor = canvas.Hex(rectOption.StrokeColor)
				}
			}
			ctx.SetStrokeColor(strokeColor)
			ctx.SetStrokeWidth(rectOption.StrokeWidth)
			ctx.DrawPath(rectOption.X, rectOption.Y, GenerateRoundedRectPath(rectOption.Width, rectOption.Height, rectOption.Radius))
			ctx.Stroke()
		}
		ctx.Fill()
	}

	scaleX = (c.W - option.Padding[1] - option.Padding[3]) / finalCanvas.W
	scaleY = (c.H - option.Padding[0] - option.Padding[2]) / finalCanvas.H

	// 将缩放后的内容画布放置到最终画布，应用padding
	c.RenderViewTo(finalCanvas, canvas.Matrix{
		{scaleX, 0, option.Padding[3]},
		{0, scaleY, option.Padding[0]},
	})

	return finalCanvas, nil
}

type Point struct {
	X float64
	Y float64
}

func GenerateRoundedRectPath(width, height, radius float64) *canvas.Path {
	// 路径点定义
	points := []Point{
		{radius, 0}, // 开始点
		{width - radius, 0},
		{width, radius},
		{width, height - radius},
		{width - radius, height},
		{radius, height},
		{0, height - radius},
		{0, radius},
		{radius, 0}, // 闭合路径
	}

	// 构建路径数据
	var pathData string
	pathData += fmt.Sprintf("M %.6f %.6f ", points[0].X, points[0].Y)                          // 移动起点
	pathData += fmt.Sprintf("L %.6f %.6f ", points[1].X, points[1].Y)                          // 左上到右上直线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", width, 0.0, points[2].X, points[2].Y)    // 右上圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[3].X, points[3].Y)                          // 右线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", width, height, points[4].X, points[4].Y) // 右下圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[5].X, points[5].Y)                          // 下线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", 0.0, height, points[6].X, points[6].Y)   // 左下圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[7].X, points[7].Y)                          // 左线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f Z", 0.0, 0.0, points[8].X, points[8].Y)     // 左上圆角并闭合
	path := canvas.MustParseSVGPath(pathData)
	return path
}
