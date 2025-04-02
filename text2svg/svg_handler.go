package text2svg

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

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

// handleSVGSave 处理SVG格式保存的特殊逻辑
func handleSVGSave(c *canvas.Canvas, options *Options, config SaveConfig) (canvas *canvas.Canvas, err error) {
	var buf bytes.Buffer
	if err := c.Write(&buf, renderers.SVG()); err != nil {
		return nil, fmt.Errorf("渲染SVG失败: %v", err)
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
		return nil, fmt.Errorf("保存SVG文件失败: %v", err)
	}

	return c, nil
}
