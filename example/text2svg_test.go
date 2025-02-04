package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

func TestText2svg(t *testing.T) {
	options := text2svg.Options{
		Text:     "Hello World",
		FontPath: "Cookie.ttf",
		FontSize: 10,
		Width:    100,
		Height:   30,
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
