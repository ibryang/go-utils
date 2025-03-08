package text2svg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ibryang/go-utils/changedpi"
	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
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
}

// Result 包含转换结果
type Result struct {
	Svg    string  // SVG内容
	Width  float64 // 最终宽度
	Height float64 // 最终高度
	Error  error   // 错误信息
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

// cleanSVG 清理SVG中的空路径
func cleanSVG(svg string) string {
	// 定义多个精确的正则表达式模式，以匹配不同形式的空路径
	patterns := []string{
		// 基本的空路径模式
		`<path d=""></path>`,
		`<path d=''></path>`,
		`<path d=""/>`,
		`<path d=''/>`,
		`<path d></path>`,

		// 带属性的空路径（通用模式）
		`<path [^>]*d=""\s*[^>]*></path>`,
		`<path [^>]*d=''\s*[^>]*></path>`,
		`<path [^>]*d=""\s*[^>]*/>`,
		`<path [^>]*d=''\s*[^>]*/>`,
	}

	// 依次应用每个模式进行替换
	result := svg
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	return result
}

// Convert 将文本转换为SVG
func Convert(options Options) *Result {
	output := &Result{}

	if options.Text == "" {
		output.Error = fmt.Errorf("文本内容不能为空")
		return output
	}

	font, err := loadFontFamily(options.FontPath)
	if err != nil {
		output.Error = fmt.Errorf("加载字体失败: %v", err)
		return output
	}

	face := font.Face(options.FontSize, nil)
	path, _, err := face.ToPath(options.Text)
	if err != nil {
		output.Error = fmt.Errorf("转换文本到路径失败: %v", err)
		return output
	}

	if path == nil {
		output.Error = fmt.Errorf("生成路径失败")
		return output
	}

	bounds := path.Bounds()
	originalWidth, originalHeight := bounds.W(), bounds.H()
	x, y := bounds.X0, bounds.Y0

	width, height, scaleX, scaleY := calculateDimensions(originalWidth, originalHeight, options.Width, options.Height)

	c := canvas.New(width, height)
	ctx := canvas.NewContext(c)

	if scaleX != 1 || scaleY != 1 {
		ctx.Scale(scaleX, scaleY)
	}

	ctx.DrawPath(-x, -y, path)

	var buf bytes.Buffer
	if err := c.Write(&buf, renderers.SVG()); err != nil {
		output.Error = fmt.Errorf("渲染SVG失败: %v", err)
		return output
	}

	svg := cleanSVG(buf.String())
	if options.IsBase64 {
		svg = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	}
	output.Svg = svg
	output.Width = width
	output.Height = height
	return output
}

// saveToFile 保存画布到文件
func saveToFile(c *canvas.Canvas, config SaveConfig) error {
	if config.Path == "" {
		return fmt.Errorf("保存路径不能为空")
	}

	// 设置默认值
	if config.DPI == 0 {
		config.DPI = 72
	}
	if config.DPMM == 0 {
		config.DPMM = 2.8346456692913385
	}
	if config.Quality == 0 {
		config.Quality = 80
	}

	// 根据不同格式保存
	switch config.Format {
	case FormatPNG:
		if err := savePNG(c, config); err != nil {
			return err
		}
	case FormatJPEG, FormatJPG:
		if err := saveJPEG(c, config); err != nil {
			return err
		}
	case FormatSVG:
		if err := c.WriteFile(config.Path, renderers.SVG()); err != nil {
			return fmt.Errorf("保存SVG文件失败: %v", err)
		}
	case FormatPDF:
		if err := c.WriteFile(config.Path, renderers.PDF()); err != nil {
			return fmt.Errorf("保存PDF文件失败: %v", err)
		}
	case FormatTIFF, FormatTIF:
		if err := c.WriteFile(config.Path, renderers.TIFF()); err != nil {
			return fmt.Errorf("保存TIFF文件失败: %v", err)
		}
	default:
		return fmt.Errorf("不支持的文件格式: %s", config.Format)
	}

	return nil
}

// savePNG 保存PNG格式
func savePNG(c *canvas.Canvas, config SaveConfig) error {
	if err := c.WriteFile(config.Path, renderers.PNG(canvas.DPI(config.DPI))); err != nil {
		return fmt.Errorf("保存PNG文件失败: %v", err)
	}

	// 如果DPI不是72，需要更新DPI信息
	if config.DPI != 72 {
		return updateImageDPI(config.Path, int(config.DPI))
	}
	return nil
}

// saveJPEG 保存JPEG格式
func saveJPEG(c *canvas.Canvas, config SaveConfig) error {
	if err := c.WriteFile(config.Path, renderers.JPEG(canvas.DPI(config.DPI), config.Quality)); err != nil {
		return fmt.Errorf("保存JPEG文件失败: %v", err)
	}

	// 如果DPI不是72，需要更新DPI信息
	if config.DPI != 72 {
		return updateImageDPI(config.Path, int(config.DPI))
	}
	return nil
}

// updateImageDPI 更新图片DPI信息
func updateImageDPI(path string, dpi int) error {
	baseData, err := changedpi.ChangeDpiByPath(path, dpi)
	if err != nil {
		return fmt.Errorf("更新DPI失败: %v", err)
	}

	if err := changedpi.SaveImage(path, baseData); err != nil {
		return fmt.Errorf("保存更新DPI后的图片失败: %v", err)
	}
	return nil
}

// CanvasConvert 转换并保存文件
func CanvasConvert(options Options) error {
	c, err := GenerateCanvas(options)
	if err != nil {
		return err
	}

	if options.SavePath == "" {
		return nil
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
		var buf bytes.Buffer
		if err := c.Write(&buf, renderers.SVG()); err != nil {
			return fmt.Errorf("渲染SVG失败: %v", err)
		}

		svg := buf.String()

		// 清理SVG内的空路径
		svg = cleanSVG(svg)

		// 如果启用了背景和圆角，替换圆角矩形路径
		if options.EnableBackground && options.BorderRadius > 0 {
			// 获取画布尺寸
			finalWidth, finalHeight := c.Size()

			// 创建新的圆角矩形路径
			newRectPath := createSVGRoundedRect(finalWidth, finalHeight, options.BorderRadius)

			// 构建新的路径标签
			var newPathTag strings.Builder

			// 开始标签
			newPathTag.WriteString("<path d=\"")
			newPathTag.WriteString(newRectPath)
			newPathTag.WriteString("\" style=\"fill:")
			newPathTag.WriteString(options.BackgroundColor)

			// 如果有描边，添加描边属性
			if options.BackgroundStroke != "" {
				newPathTag.WriteString(";stroke:")
				newPathTag.WriteString(options.BackgroundStroke)
				newPathTag.WriteString(";stroke-width:")
				newPathTag.WriteString(strconv.FormatFloat(options.BackgroundStrokeWidth, 'f', 1, 64))
			}

			// 闭合标签
			newPathTag.WriteString("\"/>")

			// 在SVG中找到第一个路径元素
			replacePattern := regexp.MustCompile(`<path[^>]*d=["'][^"']*["'][^>]*>(?:</path>)?`)
			match := replacePattern.FindStringIndex(svg)

			if match != nil {
				// 替换第一个路径（假设是背景矩形）
				svg = svg[:match[0]] + newPathTag.String() + svg[match[1]:]
			} else {
				// 如果没有找到路径标签，在SVG元素开始后插入
				svgStartTag := regexp.MustCompile(`<svg[^>]*>`)
				startMatch := svgStartTag.FindStringIndex(svg)
				if startMatch != nil {
					svg = svg[:startMatch[1]] + "\n" + newPathTag.String() + svg[startMatch[1]:]
				}
			}
		}

		// 保存修改后的SVG
		if err := SaveToFile(svg, config.Path); err != nil {
			return fmt.Errorf("保存SVG文件失败: %v", err)
		}

		return nil
	}

	return saveToFile(c, config)
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

func GenerateCanvas(options Options) (*canvas.Canvas, error) {
	if len(options.Colors) == 0 {
		options.Colors = []string{"#000000"}
	}

	for i := 0; i < len(options.Colors); i++ {
		if options.Colors[i] == "none" {
			options.Colors[i] = "#00000000"
		}
	}

	// 设置默认值
	if options.EnableStroke && options.StrokeWidth <= 0 {
		options.StrokeWidth = 1.0
	}
	if options.EnableStroke && options.StrokeColor == "" {
		options.StrokeColor = "#000000"
	}
	if options.EnableBackground && options.BackgroundColor == "" {
		options.BackgroundColor = "#FFFFFF"
	}
	if options.BackgroundStroke != "" && options.BackgroundStrokeWidth <= 0 {
		options.BackgroundStrokeWidth = 1.0
	}

	// 处理内边距
	options.Padding = processPadding(options.Padding)

	// 加载字体
	font, err := loadFontFamily(options.FontPath)
	if err != nil {
		return nil, fmt.Errorf("加载字体失败: %v", err)
	}

	face := font.Face(options.FontSize, nil)

	// 计算总宽度和高度
	var totalWidth float64
	var maxHeight float64
	var minY float64
	var maxY float64
	var paths []*canvas.Path
	var xOffsets []float64
	var bounds []canvas.Rect
	var colorIndices []int    // 存储每个字符对应的颜色索引
	var colorCount int        // 非空格字符计数
	var lastCharWidth float64 // 记录最后一个非空格字符的实际宽度

	// 第一遍：收集所有路径和边界信息
	runes := []rune(options.Text)
	for i, char := range runes {
		path, advance, err := face.ToPath(string(char))
		if err != nil {
			return nil, fmt.Errorf("转换文本到路径失败: %v", err)
		}

		// 处理空格字符
		if char == ' ' {
			colorIndices = append(colorIndices, -1) // 用-1标记空格
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
		colorCount++ // 只对非空格字符计数

		// 更新Y坐标范围
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

		// 记录最后一个非空格字符的实际宽度
		lastCharWidth = pathBounds.W()

		// 如果不是最后一个字符，使用advance值
		if i < len(runes)-1 {
			totalWidth += advance
		}
	}

	// 添加最后一个字符的实际宽度
	if lastCharWidth > 0 {
		totalWidth += lastCharWidth
	}

	// 使用实际的字符高度范围计算总高度
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

	// 创建文字上下文
	ctx := canvas.NewContext(c)

	// 计算文字位置以居中显示
	textX := (width - totalWidth*scaleX) / 2
	textY := (height - maxHeight*scaleY) / 2

	// 移动原点到文字位置
	ctx.Translate(textX, textY)

	// 应用缩放
	ctx.Scale(scaleX, scaleY)

	// 计算文本的起始位置（考虑内边距和描边宽度）
	offsetX := options.Padding[3]
	offsetY := options.Padding[0]

	if options.EnableStroke {
		offsetX += options.StrokeWidth
		offsetY += options.StrokeWidth
	}

	// 绘制每个字符并设置颜色
	pathIndex := 0
	for _, colorIndex := range colorIndices {
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
		charX := offsetX + xOffsets[pathIndex] - bounds[pathIndex].X0
		charY := offsetY - minY

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

	// 绘制额外的文本
	if len(options.ExtraTexts) > 0 {
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
				return c, fmt.Errorf("加载额外文本字体失败: %v", err)
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
			if err != nil {
				return c, fmt.Errorf("转换额外文本到路径失败: %v", err)
			}

			if extraPath == nil {
				continue // 跳过无效路径
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

	return c, nil
}

// SaveToFile 将SVG保存到文件
func SaveToFile(svg string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(svg)
	return err
}

// createSVGRoundedRect 创建一个圆角矩形的SVG路径字符串，确保兼容CorelDRAW
func createSVGRoundedRect(width, height, radius float64) string {
	// 确保半径不超过宽度或高度的一半
	if radius > width/2 {
		radius = width / 2
	}
	if radius > height/2 {
		radius = height / 2
	}

	// 明确将所有数值转换为浮点数，避免隐式类型转换问题
	rx := radius
	ry := radius
	x0 := 0.0
	y0 := 0.0
	w := width
	h := height

	// 将浮点数转换为字符串的辅助函数，固定使用1位小数
	f := func(v float64) string {
		return strconv.FormatFloat(v, 'f', 1, 64)
	}

	// 直接拼接字符串，完全避免使用fmt包
	var path strings.Builder

	// M: 起点 (rx, 0)
	path.WriteString("M")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(y0))
	path.WriteString(" ")

	// L: 右上角圆弧起点
	path.WriteString("L")
	path.WriteString(f(w - rx))
	path.WriteString(" ")
	path.WriteString(f(y0))
	path.WriteString(" ")

	// A: 右上角圆弧
	path.WriteString("A")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" 0 0 1 ")
	path.WriteString(f(w))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" ")

	// L: 右边
	path.WriteString("L")
	path.WriteString(f(w))
	path.WriteString(" ")
	path.WriteString(f(h - ry))
	path.WriteString(" ")

	// A: 右下角圆弧
	path.WriteString("A")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" 0 0 1 ")
	path.WriteString(f(w - rx))
	path.WriteString(" ")
	path.WriteString(f(h))
	path.WriteString(" ")

	// L: 下边
	path.WriteString("L")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(h))
	path.WriteString(" ")

	// A: 左下角圆弧
	path.WriteString("A")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" 0 0 1 ")
	path.WriteString(f(x0))
	path.WriteString(" ")
	path.WriteString(f(h - ry))
	path.WriteString(" ")

	// L: 左边
	path.WriteString("L")
	path.WriteString(f(x0))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" ")

	// A: 左上角圆弧
	path.WriteString("A")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(ry))
	path.WriteString(" 0 0 1 ")
	path.WriteString(f(rx))
	path.WriteString(" ")
	path.WriteString(f(y0))

	// Z: 闭合路径
	path.WriteString("Z")

	return path.String()
}
