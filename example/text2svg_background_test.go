package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

func TestText2svgWithBackground(t *testing.T) {
	options := text2svg.Options{
		Text:     "JoseéphineQ",
		FontPath: "LDRoadsDEMO",
		FontSize: 230.15,
		Colors:   []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath: "text2svg_background.svg",
		// EnableBackground: true,
		BackgroundColor: "#f0f0f0",
		BorderRadius:    10,
		Padding:         []float64{20},
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

func TestText2svgWithStroke(t *testing.T) {
	options := text2svg.Options{
		Text:         "JoseéphineQ",
		FontPath:     "LDRoadsDEMO",
		FontSize:     230.15,
		Colors:       []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath:     "text2svg_stroke.svg",
		EnableStroke: true,
		StrokeWidth:  0.1,
		Width:        500,
		StrokeColor:  "#FF0000",
		Padding:      []float64{10},
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

func TestText2svgWithBackgroundAndStroke(t *testing.T) {
	options := text2svg.Options{
		Text:                  "JoseéphineQ",
		FontPath:              "LDRoadsDEMO",
		FontSize:              230.15,
		Colors:                []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath:              "text2svg_full.svg",
		EnableBackground:      true,
		BackgroundColor:       "#2179b900",
		BackgroundStroke:      "#2179b9",
		BackgroundStrokeWidth: 0.1,
		BorderRadius:          15,
		EnableStroke:          true,
		StrokeWidth:           0.1,
		StrokeColor:           "#f3b747",
		Padding:               []float64{25},
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

// 测试背景带描边
func TestBackgroundWithStroke(t *testing.T) {
	options := text2svg.Options{
		Text:                  "JosephineQ",
		FontPath:              "LDRoadsDEMO",
		FontSize:              229.8,
		Colors:                []string{"none"},
		SavePath:              "text2svg_bg_stroke.svg",
		EnableBackground:      true,
		EnableStroke:          true,      // 开启文字描边
		StrokeWidth:           0.05,      // 文字描边宽度
		StrokeColor:           "#FF0000", // 文字描边颜色
		BackgroundColor:       "none",    // 背景填充色
		BackgroundStroke:      "#FF0000", // 背景描边颜色
		BackgroundStrokeWidth: 0.1,       // 背景描边宽度
		BorderRadius:          5,
		Padding:               []float64{9.5, 20},
		LockWidth:             450,
		LockHeight:            95,
		ExtraTexts: []text2svg.ExtraTextInfo{
			{
				Text:     "001",
				FontPath: "Cookie",
				FontSize: 16,
				Color:    "#ff0000",
				X:        0,
				Y:        5,
			},
		},
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

// 测试透明背景，只有边框
func TestTransparentBackground(t *testing.T) {
	options := text2svg.Options{
		Text:                  "JOSÉPHINEQ",
		FontPath:              "LDRoadsDEMO",
		FontSize:              48,
		Colors:                []string{"#ca2128", "#dc602c", "#f3b747", "#07954b", "#2179b9", "#21378c"},
		SavePath:              "text2svg_transparent_bg.svg",
		EnableBackground:      true,
		EnableStroke:          false,       // 开启文字描边
		StrokeWidth:           1.0,         // 文字描边宽度
		StrokeColor:           "#333333",   // 文字描边颜色
		BackgroundColor:       "#ffffff00", // 透明背景
		BackgroundStroke:      "#0000ff",   // 蓝色背景边框
		BackgroundStrokeWidth: 0.1,         // 背景边框宽度
		BorderRadius:          5,
		Padding:               []float64{0},
		Height:                50, // 指定宽度
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

// 测试矩形边框是否正常
func TestRectangleShape(t *testing.T) {
	options := text2svg.Options{
		Text:                  "矩形测试",
		FontPath:              "Arial",
		FontSize:              36,
		Colors:                []string{"#000000"},
		SavePath:              "text2svg_rectangle.svg",
		EnableBackground:      true,
		EnableStroke:          false,
		BackgroundColor:       "#ffffff",
		BackgroundStroke:      "#ff0000",
		BackgroundStrokeWidth: 4.0,
		BorderRadius:          3, // 不使用圆角
		Padding:               []float64{15},
		Width:                 300,
		Height:                150, // 指定高度和宽度
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
}

// 测试不同Padding值的效果
func TestPaddingFormats(t *testing.T) {
	// 基本配置
	baseOptions := text2svg.Options{
		Text:                  "内边距测试",
		FontPath:              "Arial",
		FontSize:              36,
		Colors:                []string{"#000000"},
		EnableBackground:      true,
		BackgroundColor:       "#ffffff",
		BackgroundStroke:      "#ff0000",
		BackgroundStrokeWidth: 2.0,
		BorderRadius:          5,
	}

	// 测试单值内边距 [所有方向相同]
	singleValueOptions := baseOptions
	singleValueOptions.Padding = []float64{20}
	singleValueOptions.SavePath = "text2svg_padding_single.svg"
	_, err := text2svg.CanvasConvert(singleValueOptions)
	if err != nil {
		t.Fatalf("单值内边距测试失败: %v", err)
	}

	// 测试双值内边距 [上下, 左右]
	twoValueOptions := baseOptions
	twoValueOptions.Padding = []float64{10, 30}
	twoValueOptions.SavePath = "text2svg_padding_two.svg"
	_, err = text2svg.CanvasConvert(twoValueOptions)
	if err != nil {
		t.Fatalf("双值内边距测试失败: %v", err)
	}

	// 测试三值内边距 [上, 左右, 下]
	threeValueOptions := baseOptions
	threeValueOptions.Padding = []float64{5, 20, 35}
	threeValueOptions.SavePath = "text2svg_padding_three.svg"
	_, err = text2svg.CanvasConvert(threeValueOptions)
	if err != nil {
		t.Fatalf("三值内边距测试失败: %v", err)
	}

	// 测试四值内边距 [上, 右, 下, 左]
	fourValueOptions := baseOptions
	fourValueOptions.Padding = []float64{10, 20, 30, 40}
	fourValueOptions.SavePath = "text2svg_padding_four.svg"
	_, err = text2svg.CanvasConvert(fourValueOptions)
	if err != nil {
		t.Fatalf("四值内边距测试失败: %v", err)
	}
}

// 测试固定大小的情况
func TestLockDimensions(t *testing.T) {
	// 基本配置
	baseOptions := text2svg.Options{
		Text:                  "固定尺寸测试",
		FontPath:              "Arial",
		FontSize:              36,
		Colors:                []string{"#000000"},
		EnableBackground:      true,
		BackgroundColor:       "#ffffff",
		BackgroundStroke:      "#ff0000",
		BackgroundStrokeWidth: 2.0,
		BorderRadius:          5,
	}

	// 测试固定宽度
	lockWidthOptions := baseOptions
	lockWidthOptions.Padding = []float64{10}
	lockWidthOptions.LockWidth = 500
	lockWidthOptions.SavePath = "text2svg_lock_width.svg"
	_, err := text2svg.CanvasConvert(lockWidthOptions)
	if err != nil {
		t.Fatalf("固定宽度测试失败: %v", err)
	}

	// 测试固定高度
	lockHeightOptions := baseOptions
	lockHeightOptions.Padding = []float64{10}
	lockHeightOptions.LockHeight = 200
	lockHeightOptions.SavePath = "text2svg_lock_height.svg"
	_, err = text2svg.CanvasConvert(lockHeightOptions)
	if err != nil {
		t.Fatalf("固定高度测试失败: %v", err)
	}

	// 测试同时固定宽度和高度
	lockBothOptions := baseOptions
	lockBothOptions.Padding = []float64{10}
	lockBothOptions.LockWidth = 500
	lockBothOptions.LockHeight = 200
	lockBothOptions.SavePath = "text2svg_lock_both.svg"
	_, err = text2svg.CanvasConvert(lockBothOptions)
	if err != nil {
		t.Fatalf("固定宽高测试失败: %v", err)
	}

	// 测试与预设内边距结合使用
	customPaddingOptions := baseOptions
	customPaddingOptions.Padding = []float64{5, 30, 50, 10}
	customPaddingOptions.LockWidth = 600
	customPaddingOptions.LockHeight = 300
	customPaddingOptions.SavePath = "text2svg_lock_custom_padding.svg"
	_, err = text2svg.CanvasConvert(customPaddingOptions)
	if err != nil {
		t.Fatalf("固定尺寸与自定义内边距测试失败: %v", err)
	}
}
