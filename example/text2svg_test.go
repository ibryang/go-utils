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
		Text:     "Tommy leeq-+",
		FontPath: "LDRoadsDEMO",
		FontSize: 230.15,
		Colors:   []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath: "text2svg_colors.png",
		DPI:      300,
	}

	err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}
