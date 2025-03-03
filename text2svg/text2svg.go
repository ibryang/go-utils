package text2svg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/ibryang/go-utils/changedpi"
	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

// Options 定义文本转SVG的配置选项
type Options struct {
	Text     string   // 要转换的文本内容
	FontPath string   // 字体文件路径或字体名称
	FontSize float64  // 字体大小
	IsBase64 bool     // 是否输出base64编码的SVG
	Width    float64  // 目标宽度，可选
	Height   float64  // 目标高度，可选
	Colors   []string // 颜色列表
	SavePath string   // 保存路径
	Format   string   // 保存格式
	DPI      float64  // 保存DPI
	DPMM     float64  // 保存DPMM
	Quality  int      // 保存质量
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
)

// SaveConfig 保存配置
type SaveConfig struct {
	Format  SaveFormat
	Path    string
	DPI     float64
	DPMM    float64
	Quality int
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

	svg := buf.String()
	if options.IsBase64 {
		svg = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
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

	return saveToFile(c, config)
}

func GenerateCanvas(options Options) (*canvas.Canvas, error) {
	if len(options.Colors) == 0 {
		options.Colors = []string{"#000000"}
	}

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

	// 第一遍：收集所有路径和边界信息
	for _, char := range options.Text {
		path, advance, err := face.ToPath(string(char))
		if err != nil {
			return nil, fmt.Errorf("转换文本到路径失败: %v", err)
		}
		if path == nil {
			continue
		}
		pathBounds := path.Bounds()
		bounds = append(bounds, pathBounds)
		paths = append(paths, path)
		xOffsets = append(xOffsets, totalWidth)

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

		// 使用实际字符宽度而不是advance值来计算位置
		// 对于最后一个字符，我们使用实际宽度
		if len(paths) == len(options.Text) {
			totalWidth += pathBounds.W()
		} else {
			totalWidth += advance
		}
	}

	// 使用实际的字符高度范围计算总高度
	maxHeight = maxY - minY

	// 计算最终尺寸
	width, height, scaleX, scaleY := calculateDimensions(totalWidth, maxHeight, options.Width, options.Height)

	// 创建画布，确保有足够的空间
	c := canvas.New(width, height)
	ctx := canvas.NewContext(c)

	if scaleX != 1 || scaleY != 1 {
		ctx.Scale(scaleX, scaleY)
	}

	// 绘制每个字符并设置颜色
	for i, path := range paths {
		colorIndex := i % len(options.Colors)
		ctx.SetFillColor(canvas.Hex(options.Colors[colorIndex]))
		// 调整Y偏移以确保字符完全显示
		ctx.DrawPath(xOffsets[i]-bounds[i].X0, -minY, path)
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
