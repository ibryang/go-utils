package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

func TestText2svg(t *testing.T) {
	options := text2svg.Options{
		Text:     "JosephineQ",
		FontPath: "LDRoadsDEMO",
		FontSize: 230.15,
		ExtraTexts: []text2svg.ExtraTextInfo{
			{
				// 左下角(0,0)位置
				Text:     "001",
				FontSize: 16,
				Color:    "#ff0000",
				X:        20,
				Y:        5, // 从底部算，接近底部
			},
		},
	}

	result := text2svg.Convert(options)
	if result.Error != nil {
		t.Fatalf("转换失败: %v", result.Error)
	}

	t.Logf("SVG生成成功，最终尺寸: %.2f x %.2f\n", result.Width, result.Height)

	if err := text2svg.SaveToFile(result.Svg, "text2svg.svg"); err != nil {
		t.Fatalf("保存文件失败: %v", err)
	}
}

func TestText2svgColors(t *testing.T) {
	options := text2svg.Options{
		// Text:     "Léo",
		Text:         "Benjamin",
		FontPath:     "Damion-Regular",
		FontSize:     230.15,
		Colors:       []string{"none"},
		SavePath:     "text2svg_colors.svg",
		DPI:          300,
		EnableStroke: true,
		StrokeWidth:  0.1,
		StrokeColor:  "#000000",
		RenderMode:   text2svg.RenderModeString,
	}

	err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}
