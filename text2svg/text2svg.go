package text2svg

import (
	"fmt"
	"strings"

	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
)

// RenderMode 定义文本渲染模式
type RenderMode int

const (
	// RenderModeChar 逐字符单独渲染模式
	RenderModeChar RenderMode = iota
	// RenderModeString 整体字符串渲染模式
	RenderModeString
)

// Options 定义文本转SVG的配置选项
//
// Width和Height与LockWidth和LockHeight的区别：
//
//   - Width/Height: 用于指定目标尺寸，但会保持内容的原始宽高比。
//     如果只设置Width或Height中的一个，另一个会自动按比例计算。
//     如果同时设置Width和Height，图像会缩放以适应这个区域，但可能会有一些"空白"。
//
//   - LockWidth/LockHeight: 用于精确锁定最终输出的尺寸，并通过动态调整内边距实现。
//     设置LockWidth会自动计算并调整左右内边距，使最终宽度精确等于指定值。
//     设置LockHeight会自动计算并调整上下内边距，使最终高度精确等于指定值。
//     可以同时设置两者，此时会精确控制最终的宽度和高度。
//
// 使用场景：
//   - 需要指定内容大小但保持比例：使用Width/Height
//   - 需要输出精确尺寸的图像：使用LockWidth/LockHeight
//   - 需要在固定尺寸下自动居中内容：使用LockWidth/LockHeight
type Options struct {
	Text                  string          // 要转换的文本内容
	FontPath              string          // 字体文件路径或字体名称
	FontSize              float64         // 字体大小
	IsBase64              bool            // 是否输出base64编码的SVG
	Width                 float64         // 目标宽度，可选
	Height                float64         // 目标高度，可选
	Colors                []string        // 颜色列表
	SavePath              string          // 保存路径
	Format                string          // 保存格式
	DPI                   float64         // 保存DPI
	DPMM                  float64         // 保存DPMM
	Quality               int             // 保存质量
	EnableStroke          bool            // 是否启用描边
	StrokeWidth           float64         // 描边宽度
	StrokeColor           string          // 描边颜色
	EnableBackground      bool            // 是否启用背景矩形
	BackgroundColor       string          // 背景颜色
	BackgroundStroke      string          // 背景描边颜色
	BackgroundStrokeWidth float64         // 背景描边宽度
	BorderRadius          float64         // 背景矩形圆角半径
	Padding               []float64       // 内边距：[上, 右, 下, 左]，支持1-4个值，类似CSS padding
	LockWidth             float64         // 锁定最终宽度（如果设置，将动态调整水平内边距）
	LockHeight            float64         // 锁定最终高度（如果设置，将动态调整垂直内边距）
	ExtraTexts            []ExtraTextInfo // 额外的文本信息列表
	RenderMode            RenderMode      // 渲染模式
	MirrorX               bool            // X轴镜像
	MirrorY               bool            // Y轴镜像
}

// SaveFormat 定义保存格式
type SaveFormat string

const (
	FormatSVG  SaveFormat = "svg"
	FormatPNG  SaveFormat = "png"
	FormatJPEG SaveFormat = "jpeg"
	FormatJPG  SaveFormat = "jpg"
	FormatPDF  SaveFormat = "pdf"
	FormatTIFF SaveFormat = "tiff"
	FormatTIF  SaveFormat = "tif"
)

// SaveConfig 保存配置
type SaveConfig struct {
	Format  SaveFormat
	Path    string
	DPI     float64
	DPMM    float64
	Quality int
}

// ExtraTextInfo 定义额外的文本信息
// 坐标系统说明：
// - X轴：使用左侧为原点(0)，向右增加
// - Y轴：使用底部为原点(0)，向上增加
// - 文本会精确放置在(X,Y)坐标位置
// - OffsetX和OffsetY可用于微调位置
type ExtraTextInfo struct {
	Text        string  // 文本内容
	FontPath    string  // 字体路径，如果为空则使用主文本的字体
	FontSize    float64 // 字体大小，如果为0则使用主文本的字体大小
	Color       string  // 文本颜色，如果为空则使用黑色
	X           float64 // X坐标（左侧为原点）
	Y           float64 // Y坐标（底部为原点）
	Rotate      float64 // 旋转角度（度数）
	Opacity     float64 // 透明度（0-1）
	StrokeText  bool    // 是否启用文本描边
	StrokeWidth float64 // 描边宽度
	StrokeColor string  // 描边颜色
	OffsetX     float64 // X方向额外偏移
	OffsetY     float64 // Y方向额外偏移
}

// CanvasConvert 转换并保存文件
func CanvasConvert(options Options) (canvas *canvas.Canvas, err error) {
	// 参数验证
	if err := validateOptions(&options); err != nil {
		return nil, err
	}

	// 生成画布
	c, err := GenerateCanvas(options)
	if err != nil {
		return nil, err
	}

	if options.SavePath == "" {
		return nil, fmt.Errorf("保存路径不能为空")
	}

	// 如果未指定格式，从文件扩展名获取
	if options.Format == "" {
		options.Format = strings.ToLower(file.ExtName(options.SavePath))
	}

	// 创建保存配置
	config := SaveConfig{
		Format:  SaveFormat(options.Format),
		Path:    options.SavePath,
		DPI:     options.DPI,
		DPMM:    options.DPMM,
		Quality: options.Quality,
	}

	// 如果是SVG格式，进行特殊处理
	if config.Format == FormatSVG {
		return handleSVGSave(c, &options, config)
	}

	return saveToFile(c, config)
}

// GenerateCanvas 将文本转换为画布 - 为兼容旧版API而保留
func GenerateCanvas(options Options) (*canvas.Canvas, error) {
	// 参数验证和默认值设置
	if err := validateOptions(&options); err != nil {
		return nil, err
	}

	// 委托给内部实现函数
	return generateCanvasInternal(options)
}

// processPadding 处理内边距，根据提供的值的数量返回[上,右,下,左]格式的完整内边距
// 类似CSS padding，支持1-4个值:
// - 1个值: 所有方向使用相同的内边距
// - 2个值: 第一个值用于上下，第二个值用于左右
// - 3个值: 第一个值用于上，第二个值用于左右，第三个值用于下
// - 4个值: 分别用于上、右、下、左
func processPadding(padding []float64) []float64 {
	result := []float64{0, 0, 0, 0} // 默认值：[上,右,下,左]

	switch len(padding) {
	case 0:
		// 使用默认值
	case 1:
		// 一个值：所有方向相同
		result[0] = padding[0] // 上
		result[1] = padding[0] // 右
		result[2] = padding[0] // 下
		result[3] = padding[0] // 左
	case 2:
		// 两个值：上下、左右
		result[0] = padding[0] // 上
		result[1] = padding[1] // 右
		result[2] = padding[0] // 下
		result[3] = padding[1] // 左
	case 3:
		// 三个值：上、左右、下
		result[0] = padding[0] // 上
		result[1] = padding[1] // 右
		result[2] = padding[2] // 下
		result[3] = padding[1] // 左
	case 4:
		// 四个值：上、右、下、左
		result = padding
	default:
		// 超过4个值：只使用前4个
		result = padding[:4]
	}

	// 确保所有值都不小于0
	for i := range result {
		if result[i] < 0 {
			result[i] = 0
		}
	}

	return result
}
