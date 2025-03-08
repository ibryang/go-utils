package example_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

// TestCdrCompatibleRoundedRect 测试生成兼容CDR的圆角矩形SVG
func TestCdrCompatibleRoundedRect(t *testing.T) {
	options := text2svg.Options{
		Text:                  "JOSÉPHINEQ",
		FontPath:              "LDRoadsDEMO",
		FontSize:              48,
		Colors:                []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath:              "text2svg_cdr_compatible.svg",
		EnableBackground:      true,
		BackgroundColor:       "#ffffff",
		BackgroundStroke:      "#ff0000",
		BackgroundStrokeWidth: 0.1,
		BorderRadius:          20, // 使用明显的圆角
		Padding:               []float64{15},
		Width:                 300,
		Height:                150,
	}

	err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("生成兼容CDR的SVG失败: %v", err)
	}

	// 验证生成的SVG文件是否存在
	if _, err := os.Stat(options.SavePath); os.IsNotExist(err) {
		t.Fatalf("生成的SVG文件不存在: %s", options.SavePath)
	}

	// 打开文件，验证内容中不包含 %!g 等格式化错误
	svgData, err := os.ReadFile(options.SavePath)
	if err != nil {
		t.Fatalf("读取生成的SVG文件失败: %v", err)
	}

	svgContent := string(svgData)
	if strings.Contains(svgContent, "%!") {
		t.Fatalf("SVG中包含格式化错误: %s", svgContent)
	}

	// 测试更小的圆角值
	optionsSmallRadius := options
	optionsSmallRadius.BorderRadius = 5
	optionsSmallRadius.SavePath = "text2svg_cdr_small_radius.svg"

	err = text2svg.CanvasConvert(optionsSmallRadius)
	if err != nil {
		t.Fatalf("生成小圆角SVG失败: %v", err)
	}

	// 测试非常大的圆角值
	optionsLargeRadius := options
	optionsLargeRadius.BorderRadius = 100 // 大于宽度/高度的一半
	optionsLargeRadius.SavePath = "text2svg_cdr_large_radius.svg"

	err = text2svg.CanvasConvert(optionsLargeRadius)
	if err != nil {
		t.Fatalf("生成大圆角SVG失败: %v", err)
	}

	// 测试锁定尺寸和自动padding
	optionsLocked := options
	optionsLocked.LockWidth = 500
	optionsLocked.LockHeight = 200
	optionsLocked.SavePath = "text2svg_cdr_locked.svg"

	err = text2svg.CanvasConvert(optionsLocked)
	if err != nil {
		t.Fatalf("生成锁定尺寸SVG失败: %v", err)
	}
}
