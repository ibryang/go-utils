package text2svgV2

import (
	"image/color"

	"github.com/tdewolff/canvas"
)

type RenderMode int

const (
	RenderString RenderMode = 1
	RenderChar   RenderMode = 2
)

// BaseOption 定义了画布的基本选项
type BaseOption struct {
	Width     float64 // 宽度
	Height    float64 // 高度
	ReverseX  bool    // X轴翻转
	ReverseY  bool    // Y轴翻转
	LockRatio bool    // 锁定宽高比例
}

// TextOption 定义了文本绘制选项
type TextOption struct {
	Text        string      // 文本内容
	FontPath    string      // 字体路径
	FontSize    float64     // 字体大小
	FontColor   any         // 字体颜色
	StrokeColor any         // 描边颜色
	StrokeWidth float64     // 描边宽度
	BaseOption              // 嵌入基本选项
	RectOption  *RectOption // 矩形选项（可选）
	// 额外的文本
	ExtraText  []ExtraTextOption // 额外的文本
	RenderMode RenderMode        // 渲染模式
}

// TextLineOption 定义了文本行选项
type TextLineOption struct {
	TextList   []TextOption      // 文本列表
	Padding    [4]float64        // 内边距 [上，右，下，左]
	LineGap    float64           // 行间距
	Align      TextAlign         // 对齐方式
	VAlign     TextAlign         // 垂直对齐方式
	BaseOption                   // 嵌入基本选项
	RectOption []RectOption      // 矩形选项列表（可选）
	ExtraText  []ExtraTextOption // 额外的文本
}

// CanvasOption 定义了画布选项
type CanvasOption struct {
	FileList   []CanvasItem      // 文件列表
	CanvasList []CanvasItem      // 画布列表
	Padding    [4]float64        // 内边距 [上，右，下，左]
	BaseOption                   // 嵌入基本选项
	RectOption []RectOption      // 矩形选项列表（可选）
	ExtraText  []ExtraTextOption // 额外的文本
}

// RectOption 定义了矩形的选项
type RectOption struct {
	Width       float64 // 宽度
	Height      float64 // 高度
	X           float64 // X坐标
	Y           float64 // Y坐标
	Radius      float64 // 圆角半径
	BgColor     string  // 背景颜色
	StrokeColor string  // 描边颜色
	StrokeWidth float64 // 描边宽度
}

// ExtraTextOption 定义了额外的文本选项
type ExtraTextOption struct {
	X          float64   // 位置X
	Y          float64   // 位置Y
	Align      TextAlign // 对齐方式
	VAlign     TextAlign // 垂直对齐方式
	TextOption           // 额外的文本
}

// Point 定义了一个2D点
type Point struct {
	X float64
	Y float64
}

// ColorMap 包含预定义的颜色映射
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

// GetColor 根据颜色名称或十六进制值返回RGBA颜色
func GetColor(colorName string) color.RGBA {
	if c, ok := ColorMap[colorName]; ok {
		return c
	}
	return canvas.Hex(colorName)
}

// TextAlign 定义文本对齐方式类型
type TextAlign string

// 文本对齐方式常量
const (
	TextAlignLeft    TextAlign = "left"
	TextAlignCenter  TextAlign = "center"
	TextAlignRight   TextAlign = "right"
	TextVAlignTop    TextAlign = "top"
	TextVAlignCenter TextAlign = "center"
	TextVAlignBottom TextAlign = "bottom"
)

// CanvasItem 定义了画布项
type CanvasItem struct {
	File   string         // 文件路径
	Canvas *canvas.Canvas // Canvas对象
	Align  TextAlign      // 对齐方式
	VAlign TextAlign      // 垂直对齐方式
	Width  float64        // 宽度
	Height float64        // 高度
	X      float64        // X坐标
	Y      float64        // Y坐标
}
