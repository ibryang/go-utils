package text2svg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // 注册JPEG解码器
	_ "image/png"  // 注册PNG解码器
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

// ImageElement 定义图片元素
type ImageElement struct {
	Path      string  // 图片或SVG文件路径
	Data      []byte  // 图片或SVG数据（直接提供数据时使用）
	X         float64 // X坐标
	Y         float64 // Y坐标
	Width     float64 // 宽度，0表示使用原始宽度
	Height    float64 // 高度，0表示使用原始高度
	Rotate    float64 // 旋转角度（度数）
	Opacity   float64 // 透明度（0-1）
	ScaleMode string  // 缩放模式："fit", "fill", "stretch"，默认为"fit"
}

// MultiElement 多元素配置结构
type MultiElement struct {
	CanvasWidth     float64        // 画布宽度
	CanvasHeight    float64        // 画布高度
	BackgroundColor string         // 背景颜色
	Images          []ImageElement // 图片元素列表
	SVGs            []ImageElement // SVG元素列表（复用ImageElement结构）
	TextOptions     []Options      // 文本元素列表（复用现有Options结构）
	SavePath        string         // 保存路径
	SaveFormat      string         // 保存格式
	DPI             float64        // 导出DPI
	Quality         int            // 导出质量（JPEG等格式使用）
}

// RenderMultiElement 渲染多元素画布
func RenderMultiElement(config MultiElement) (*canvas.Canvas, error) {
	// 参数验证
	if err := validateMultiElementConfig(&config); err != nil {
		return nil, err
	}

	// 创建画布
	c := canvas.New(config.CanvasWidth, config.CanvasHeight)
	ctx := canvas.NewContext(c)
	ctx.SetCoordSystem(canvas.CartesianIV)

	// 绘制背景
	if config.BackgroundColor != "" && config.BackgroundColor != "none" {
		// 使用hex函数解析颜色
		bgColor := canvas.Hex(config.BackgroundColor)
		ctx.SetFillColor(bgColor)
		ctx.DrawPath(0, 0, canvas.Rectangle(config.CanvasWidth, config.CanvasHeight))
	}

	// 先绘制图片
	for _, img := range config.Images {
		if err := drawImageElement(ctx, img); err != nil {
			return nil, fmt.Errorf("绘制图片元素失败: %v", err)
		}
	}

	// 绘制SVG
	for _, svg := range config.SVGs {
		fmt.Println("svg", svg)
		if err := drawSVGElement(ctx, svg); err != nil {
			return nil, fmt.Errorf("绘制SVG元素失败: %v", err)
		}
	}

	// 绘制文本
	for _, textOpt := range config.TextOptions {
		// 确保ExtraTexts至少有一个元素
		if len(textOpt.ExtraTexts) == 0 {
			continue
		}

		textCanvas, err := GenerateCanvas(textOpt)
		if err != nil {
			return nil, fmt.Errorf("生成文本元素失败: %v", err)
		}

		// 获取文本画布的尺寸
		_, textHeight := textCanvas.Size()

		// 创建一个新的Context来绘制文本画布
		extraText := textOpt.ExtraTexts[0]
		x := extraText.X
		y := config.CanvasHeight - extraText.Y - textHeight

		// 绘制文本
		ctx.Push()
		ctx.Translate(x, y)

		// 处理文本Canvas
		// 注意：这里使用渲染矩阵而不是DrawImage，因为DrawImage需要image.Image
		// 当前实现不完整，可能需要进一步调整

		ctx.Pop()
	}

	// 保存
	if config.SavePath != "" {
		saveConfig := SaveConfig{
			Format:  SaveFormat(config.SaveFormat),
			Path:    config.SavePath,
			DPI:     config.DPI,
			Quality: config.Quality,
		}

		if _, err := saveToFile(c, saveConfig); err != nil {
			return nil, fmt.Errorf("保存文件失败: %v", err)
		}
	}

	return c, nil
}

// validateMultiElementConfig 验证多元素配置
func validateMultiElementConfig(config *MultiElement) error {
	if config.CanvasWidth <= 0 || config.CanvasHeight <= 0 {
		return errors.New("画布尺寸必须大于0")
	}

	// 设置默认值
	if config.DPI == 0 {
		config.DPI = 72
	}
	if config.Quality == 0 {
		config.Quality = 80
	}
	if config.BackgroundColor == "" {
		config.BackgroundColor = "#FFFFFF"
	}

	return nil
}

// drawImageElement 绘制图片元素
func drawImageElement(ctx *canvas.Context, element ImageElement) error {
	// 使用文件路径直接加载图片
	if element.Path == "" && element.Data == nil {
		return errors.New("图片元素必须提供路径或数据")
	}

	// 判断文件扩展名
	filePath := element.Path
	ext := strings.ToLower(filepath.Ext(filePath))

	// 打开图片文件
	var file *os.File
	var err error

	if filePath != "" {
		file, err = os.Open(filePath)
		if err != nil {
			return fmt.Errorf("打开图片文件失败: %v", err)
		}
		defer file.Close()
	} else {
		// 如果是数据而不是文件路径，则创建临时文件
		file, err = ioutil.TempFile("", "img_*"+ext)
		if err != nil {
			return fmt.Errorf("创建临时文件失败: %v", err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		// 写入图片数据
		if _, err = file.Write(element.Data); err != nil {
			return fmt.Errorf("写入图片数据失败: %v", err)
		}

		// 重置文件指针
		if _, err = file.Seek(0, 0); err != nil {
			return fmt.Errorf("重置文件指针失败: %v", err)
		}
	}

	// 根据图片类型加载
	var img image.Image
	img, _, err = image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	// 获取图片尺寸
	bounds := img.Bounds()
	srcWidth := float64(bounds.Dx())
	srcHeight := float64(bounds.Dy())

	// 确定目标尺寸
	destWidth := element.Width
	destHeight := element.Height

	// 如果未指定尺寸，使用原始尺寸或按比例智能适应
	if destWidth == 0 && destHeight == 0 {
		// 两个尺寸都未指定，使用原始尺寸
		destWidth = srcWidth
		destHeight = srcHeight
	} else if destWidth > 0 && destHeight == 0 {
		// 只指定宽度，高度按比例计算
		aspectRatio := srcHeight / srcWidth
		destHeight = destWidth * aspectRatio
	} else if destWidth == 0 && destHeight > 0 {
		// 只指定高度，宽度按比例计算
		aspectRatio := srcWidth / srcHeight
		destWidth = destHeight * aspectRatio
	}

	fmt.Println("srcWidth", srcWidth, "srcHeight", srcHeight)
	fmt.Println("destWidth", destWidth, "destHeight", destHeight)

	// 处理不同的缩放模式
	scaleX := destWidth / srcWidth
	scaleY := destHeight / srcHeight

	if element.ScaleMode == "fit" || element.ScaleMode == "" {
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		destWidth = srcWidth * scale
		destHeight = srcHeight * scale
	} else if element.ScaleMode == "fill" {
		scale := scaleX
		if scaleY > scaleX {
			scale = scaleY
		}
		destWidth = srcWidth * scale
		destHeight = srcHeight * scale
	}

	fmt.Println("destWidth", destWidth, "destHeight", destHeight)
	dpi := srcWidth * 25.4 / destWidth
	fmt.Println("dpi", dpi)

	// 重新打开文件，因为image.Decode会消耗文件指针
	if _, err = file.Seek(0, 0); err != nil {
		return fmt.Errorf("重置文件指针失败: %v", err)
	}

	// 直接绘制图片
	var imgObj canvas.Image
	if strings.HasSuffix(strings.ToLower(ext), ".jpg") ||
		strings.HasSuffix(strings.ToLower(ext), ".jpeg") {
		imgObj, err = canvas.NewJPEGImage(file)
	} else {
		imgObj, err = canvas.NewPNGImage(file)
	}

	if err != nil {
		return fmt.Errorf("创建图片Canvas对象失败: %v", err)
	}
	// 不需要变换时直接绘制
	ctx.DrawImage(element.X, element.Y, imgObj, canvas.DPI(dpi))

	return nil
}

// drawSVGElement 绘制SVG元素
func drawSVGElement(ctx *canvas.Context, element ImageElement) error {
	var svgData []byte
	var err error

	// 加载SVG数据
	if element.Path != "" {
		svgData, err = ioutil.ReadFile(element.Path)
	} else if element.Data != nil {
		svgData = element.Data
	} else {
		return errors.New("SVG元素必须提供路径或数据")
	}

	if err != nil {
		return err
	}

	// 解析SVG
	svgReader := bytes.NewReader(svgData)
	svgCanvas, err := canvas.ParseSVG(svgReader)
	if err != nil {
		return fmt.Errorf("解析SVG失败: %v", err)
	}
	fmt.Println("svgCanvas", svgCanvas.H, svgCanvas.W)

	scaleX := element.Width / svgCanvas.W
	scaleY := element.Height / svgCanvas.H
	if element.Width == 0 && element.Height == 0 {
		scaleX = 1
		scaleY = 1
	} else if element.Width == 0 && element.Height > 0 {
		scaleY = element.Height / svgCanvas.H
		scaleX = scaleY
	} else if element.Width > 0 && element.Height == 0 {
		scaleX = element.Width / svgCanvas.W
		scaleY = scaleX
	}

	// 要打破循环引用，我们将svgCanvas直接渲染到主Canvas
	// 注意：这不是完整解决方案，可能需要进一步开发
	// 此处未使用不存在的GetOps和DrawOp方法
	// 使用RenderViewTo渲染
	svgCanvas.RenderViewTo(ctx, canvas.Matrix{
		{scaleX, 0, 0},
		{0, scaleY, 0},
	})

	return nil
}

// CreateMultiElementSVG 创建多元素SVG并返回SVG字符串
func CreateMultiElementSVG(config MultiElement) (string, error) {
	c, err := RenderMultiElement(config)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := c.Write(&buf, renderers.SVG()); err != nil {
		return "", fmt.Errorf("渲染SVG失败: %v", err)
	}

	return buf.String(), nil
}

// CreateMultiElementDataURL 创建多元素SVG并返回Data URL
func CreateMultiElementDataURL(config MultiElement) (string, error) {
	svgString, err := CreateMultiElementSVG(config)
	if err != nil {
		return "", err
	}

	// 转换为base64
	encoded := base64.StdEncoding.EncodeToString([]byte(svgString))
	return "data:image/svg+xml;base64," + encoded, nil
}

// ExportMultiElementToFormat 导出多元素内容到指定格式
func ExportMultiElementToFormat(config MultiElement, format string, outputPath string) error {
	// 保存为原始格式
	origSavePath := config.SavePath
	origSaveFormat := config.SaveFormat

	// 设置目标格式
	config.SavePath = outputPath
	config.SaveFormat = format

	// 渲染并保存
	_, err := RenderMultiElement(config)

	// 恢复原始配置
	config.SavePath = origSavePath
	config.SaveFormat = origSaveFormat

	return err
}

// GetSupportedExportFormats 获取支持的导出格式列表
func GetSupportedExportFormats() []string {
	return []string{
		"svg",
		"png",
		"jpg",
		"jpeg",
		"pdf",
		"tiff",
		"tif",
	}
}
