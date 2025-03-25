package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

func TestText2svgColors(t *testing.T) {
	options := text2svg.Options{
		// Text:     "Léo",
		Text:         "Benjamin",
		FontPath:     "Arial",
		FontSize:     230.15,
		Colors:       []string{"#ff0000", "#00ff00", "#0000ff"},
		SavePath:     "text2svg_colors.svg",
		DPI:          300,
		EnableStroke: false,
		StrokeWidth:  0.1,
		StrokeColor:  "#000000",
		RenderMode:   text2svg.RenderModeChar,
	}

	err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}
