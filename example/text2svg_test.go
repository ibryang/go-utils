package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

func TestText2svgColors(t *testing.T) {
	options := text2svg.Options{
		// Text:     "Léo",
		Text:         "Benjamin",
		FontPath:     "Cookie",
		FontSize:     230.15,
		Colors:       []string{"#ff0000", "#00ff00", "#0000ff"},
		SavePath:     "text2svg_colors.svg",
		DPI:          300,
		EnableStroke: false,
		StrokeWidth:  0.1,
		StrokeColor:  "#000000",
		RenderMode:   text2svg.RenderModeString,
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

func TestTextToSVG(t *testing.T) {
	options := text2svg.Options{
		Text:       "Hello, Gophers!",
		FontPath:   "Arial",
		FontSize:   24.0,
		Colors:     []string{"#FF0000"},
		SavePath:   "hell222o.svg",
		RenderMode: text2svg.RenderModeChar,
	}

	// 生成SVG
	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("生成SVG失败: %v", err)
	}

}

// TestMirror 测试镜像功能
func TestMirror(t *testing.T) {
	// 1. 水平镜像
	horizontalMirrorOptions := text2svg.Options{
		Text:       "Hello, Gophers!",
		FontPath:   "Arial",
		FontSize:   24.0,
		Colors:     []string{"#FF0000"},
		SavePath:   "mirror_x.svg",
		RenderMode: text2svg.RenderModeChar,
		MirrorX:    true,
	}

	_, err := text2svg.CanvasConvert(horizontalMirrorOptions)
	if err != nil {
		t.Fatalf("生成水平镜像SVG失败: %v", err)
	}

	// 2. 垂直镜像
	verticalMirrorOptions := text2svg.Options{
		Text:       "Hello, Gophers!",
		FontPath:   "Arial",
		FontSize:   24.0,
		Colors:     []string{"#0000FF"},
		SavePath:   "mirror_y.svg",
		RenderMode: text2svg.RenderModeChar,
		MirrorY:    true,
	}

	_, err = text2svg.CanvasConvert(verticalMirrorOptions)
	if err != nil {
		t.Fatalf("生成垂直镜像SVG失败: %v", err)
	}

	// 3. 双向镜像
	bothMirrorOptions := text2svg.Options{
		Text:       "Hello, Gophers!",
		FontPath:   "Arial",
		FontSize:   24.0,
		Colors:     []string{"#00FF00"},
		SavePath:   "mirror_both.svg",
		RenderMode: text2svg.RenderModeChar,
		MirrorX:    true,
		MirrorY:    true,
	}

	_, err = text2svg.CanvasConvert(bothMirrorOptions)
	if err != nil {
		t.Fatalf("生成双向镜像SVG失败: %v", err)
	}
}
