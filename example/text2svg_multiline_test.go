package example_test

import (
	"os"
	"testing"

	"github.com/ibryang/go-utils/text2svg"
	"github.com/tdewolff/canvas/renderers"
)

func TestMultiLineText(t *testing.T) {
	// 首先创建几个SVG文件用于测试
	// temp1 := createTempSVGFile(t, "Abcdefg", 36, "#ff0000")
	// temp2 := createTempSVGFile(t, "Abedefsffsgga", 36, "#00ff00")
	// temp3 := createTempSVGFile(t, "dddds", 36, "#0000ff")
	// defer os.Remove(temp1)
	// defer os.Remove(temp2)
	// defer os.Remove(temp3)

	files := []string{
		"/Users/ibryang/Desktop/demo/worker/manager/test_kf/车贴/pdf原图/YT006-1.svg",
		"/Users/ibryang/Desktop/demo/worker/manager/test_kf/车贴/pdf原图/YT006-2.svg",
	}

	// 测试不同对齐方式
	// testAlignments(t, files)

	// // 测试固定尺寸
	testFixedSize(t, files)

	// // 测试最大尺寸约束
	// testMaxSize(t, files)

	// // 测试边框和背景
	// testBorderAndBackground(t, files)
}

// 创建临时SVG文件
func createTempSVGFile(t *testing.T, text string, fontSize float64, color string) string {
	options := text2svg.Options{
		Text:       text,
		FontPath:   "Cookie",
		FontSize:   fontSize,
		Colors:     []string{color, "#000000"},
		SavePath:   "", // 临时生成，不保存
		RenderMode: text2svg.RenderModeChar,
	}

	// 生成Canvas
	c, err := text2svg.GenerateCanvas(options)
	if err != nil {
		t.Fatalf("生成Canvas失败: %v", err)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "temp_svg_*.svg")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer tmpFile.Close()

	// 将Canvas写入临时文件
	err = c.WriteFile(tmpFile.Name(), renderers.SVG())
	if err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	return tmpFile.Name()
}

// 测试不同对齐方式
func testAlignments(t *testing.T, files []string) {
	// 左对齐
	leftOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignLeft,
		SavePath:    "multiline_left_align.pdf",
		// Padding:     []float64{10, 20},
	}
	_, err := text2svg.CanvasConvertMultipeLine(files, leftOptions)
	if err != nil {
		t.Fatalf("左对齐测试失败: %v", err)
	}

	// 居中对齐
	centerOptions := &text2svg.MultiLineOptions{
		LineSpacing:   5,
		Alignment:     text2svg.AlignCenter,
		SavePath:      "multiline_center_align.svg",
		Padding:       []float64{10, 20},
		MarginPadding: true,
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, centerOptions)
	if err != nil {
		t.Fatalf("居中对齐测试失败: %v", err)
	}

	// 右对齐
	rightOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignRight,
		SavePath:    "multiline_right_align.pdf",
		Padding:     []float64{10, 20},
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, rightOptions)
	if err != nil {
		t.Fatalf("右对齐测试失败: %v", err)
	}
}

// 测试固定尺寸
func testFixedSize(t *testing.T, files []string) {
	// 固定宽度
	widthOptions := &text2svg.MultiLineOptions{
		LineSpacing:  5,
		Alignment:    text2svg.AlignCenter,
		Width:        400,
		Padding:      []float64{10, 0, 5},
		SavePath:     "multiline_fixed_width.pdf",
		EnableBorder: true,
		BorderColor:  "#ff0000",
		BorderWidth:  2,
	}
	_, err := text2svg.CanvasConvertMultipeLine(files, widthOptions)
	if err != nil {
		t.Fatalf("固定宽度测试失败: %v", err)
	}

	// 固定高度
	heightOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignCenter,
		Height:      300,
		Padding:     []float64{10, 10, 10, 10},
		SavePath:    "multiline_fixed_height.pdf",
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, heightOptions)
	if err != nil {
		t.Fatalf("固定高度测试失败: %v", err)
	}

	// 固定宽度和高度
	bothOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignCenter,
		Width:       200,
		Height:      200,
		Padding:     []float64{0, 30},
		SavePath:    "multiline_fixed_both.pdf",
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, bothOptions)
	if err != nil {
		t.Fatalf("固定宽度和高度测试失败: %v", err)
	}
}

// 测试最大尺寸约束
func testMaxSize(t *testing.T, files []string) {
	// 最大宽度约束
	maxWidthOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignCenter,
		SavePath:    "multiline_max_width.svg",
	}
	_, err := text2svg.CanvasConvertMultipeLine(files, maxWidthOptions)
	if err != nil {
		t.Fatalf("最大宽度约束测试失败: %v", err)
	}

	// 最大高度约束
	maxHeightOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignCenter,
		SavePath:    "multiline_max_height.svg",
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, maxHeightOptions)
	if err != nil {
		t.Fatalf("最大高度约束测试失败: %v", err)
	}

	// 最大宽度和高度约束
	maxBothOptions := &text2svg.MultiLineOptions{
		LineSpacing: 5,
		Alignment:   text2svg.AlignCenter,
		SavePath:    "multiline_max_both.svg",
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, maxBothOptions)
	if err != nil {
		t.Fatalf("最大宽度和高度约束测试失败: %v", err)
	}
}

// 测试边框和背景
func testBorderAndBackground(t *testing.T, files []string) {
	// 带边框
	borderOptions := &text2svg.MultiLineOptions{
		LineSpacing:  5,
		Alignment:    text2svg.AlignCenter,
		EnableBorder: true,
		BorderColor:  "#ff0000",
		BorderWidth:  0.1,
		SavePath:     "multiline_border.svg",
	}
	_, err := text2svg.CanvasConvertMultipeLine(files, borderOptions)
	if err != nil {
		t.Fatalf("边框测试失败: %v", err)
	}

	// 带背景和圆角边框
	bgOptions := &text2svg.MultiLineOptions{
		LineSpacing:     5,
		Alignment:       text2svg.AlignCenter,
		EnableBorder:    true,
		BorderColor:     "#ff0000",
		BorderWidth:     0.1,
		BorderRadius:    10,
		BackgroundColor: "#f0f0f0",
		SavePath:        "multiline_background.svg",
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, bgOptions)
	if err != nil {
		t.Fatalf("背景和圆角边框测试失败: %v", err)
	}

	// 导出为PNG格式
	pngOptions := &text2svg.MultiLineOptions{
		LineSpacing:     5,
		Alignment:       text2svg.AlignCenter,
		EnableBorder:    true,
		BorderColor:     "#ff0000",
		BorderWidth:     0.1,
		BorderRadius:    10,
		BackgroundColor: "#f0f0f0",
		SavePath:        "multiline_export.png",
		DPI:             300,
	}
	_, err = text2svg.CanvasConvertMultipeLine(files, pngOptions)
	if err != nil {
		t.Fatalf("PNG导出测试失败: %v", err)
	}
}
