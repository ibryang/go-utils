package text2svg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

// Options 定义文本转SVG的配置选项
type Options struct {
	Text     string  // 要转换的文本内容
	FontPath string  // 字体文件路径或字体名称
	FontSize float64 // 字体大小
	IsBase64 bool    // 是否输出base64编码的SVG
	Width    float64 // 目标宽度，可选
	Height   float64 // 目标高度，可选
}

// Result 包含转换结果
type Result struct {
	Svg    string  // SVG内容
	Width  float64 // 最终宽度
	Height float64 // 最终高度
	Error  error   // 错误信息
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
