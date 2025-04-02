package example_test

import (
	"fmt"
	"testing"

	"github.com/ibryang/go-utils/text2svg"
)

// TestExtraText 测试额外文本功能
func TestExtraText(t *testing.T) {
	// 创建基本配置
	options := text2svg.Options{
		Text:                  "主标题",
		FontPath:              "Arial",
		FontSize:              36,
		Colors:                []string{"#336699"},
		SavePath:              "text2svg_extra_text.svg",
		EnableBackground:      true,
		BackgroundColor:       "#ffffff",
		BackgroundStroke:      "#cccccc",
		BackgroundStrokeWidth: 1.0,
		BorderRadius:          10,
		Padding:               []float64{20},
		Width:                 400,
		Height:                300,
		// 添加额外文本
		ExtraTexts: []text2svg.ExtraTextInfo{
			{
				// 顶部副标题 - 现在Y=300接近顶部
				Text:     "副标题",
				FontSize: 24,
				Color:    "#666666",
				X:        200, // 居中
				Y:        250, // 从底部算，接近顶部
			},
			{
				// 底部说明文字 - 现在Y=0接近底部
				Text:     "版权所有 © 2024",
				FontSize: 12,
				Color:    "#999999",
				X:        200, // 居中
				Y:        20,  // 从底部算，接近底部
			},
			{
				// 右上角带旋转的文字
				Text:     "旋转文本",
				FontSize: 18,
				Color:    "#ff6600",
				X:        350, // 右侧
				Y:        250, // 从底部算，接近顶部
				Rotate:   -30, // 逆时针旋转30度
			},
			{
				// 带描边的文字
				Text:        "描边效果",
				FontSize:    28,
				Color:       "#ff0000",
				X:           50, // 左侧
				Y:           50, // 从底部算，接近底部
				StrokeText:  true,
				StrokeWidth: 0.5,
				StrokeColor: "#000000",
			},
		},
	}

	_, err := text2svg.CanvasConvert(options)
	if err != nil {
		t.Fatalf("生成带额外文本的SVG失败: %v", err)
	}

	// 测试多个文本 - 坐标系统以左下角(0,0)为原点
	optionsMulti := options
	optionsMulti.SavePath = "text2svg_extra_text_multi.svg"
	optionsMulti.ExtraTexts = []text2svg.ExtraTextInfo{
		{
			// 左下角(0,0)位置
			Text:     "001",
			FontSize: 16,
			Color:    "#ff0000",
			X:        20,
			Y:        5, // 从底部算，接近底部
		},
		{
			// X=100, Y=50位置
			Text:     "点(100,50)",
			FontSize: 16,
			Color:    "#00ff00",
			X:        100,
			Y:        50, // 从底部算
		},
		{
			// X=200, Y=100位置
			Text:     "点(200,150)",
			FontSize: 16,
			Color:    "#0000ff",
			X:        200,
			Y:        150, // 从底部算，中部
		},
		{
			// 右上角
			Text:     "右上角",
			FontSize: 16,
			Color:    "#ff00ff",
			X:        optionsMulti.Width - 80,
			Y:        optionsMulti.Height - 20, // 从底部算，接近顶部
		},
	}

	_, err = text2svg.CanvasConvert(optionsMulti)
	if err != nil {
		t.Fatalf("生成多文本SVG失败: %v", err)
	}

	// 测试网格坐标系 - 创建坐标轴和网格来展示坐标系统
	optionsGrid := text2svg.Options{
		Text:             "", // 不需要主文本
		FontPath:         "Arial",
		FontSize:         12,
		SavePath:         "text2svg_coordinate_grid.svg",
		EnableBackground: true,
		BackgroundColor:  "#ffffff",
		Width:            500,
		Height:           300,
		ExtraTexts:       []text2svg.ExtraTextInfo{},
	}

	// 添加X轴标签（底部）
	for i := 0; i <= 500; i += 50 {
		if i == 0 {
			continue
		}

		// X轴标签
		optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
			Text:     fmt.Sprintf("%d", i),
			FontSize: 10,
			Color:    "#666666",
			X:        float64(i),
			Y:        15, // 从底部算，接近底部
		})

		// 垂直网格线
		if i < 500 {
			optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
				Text:     "│",
				FontSize: 8,
				Color:    "#dddddd",
				X:        float64(i) - 2,
				Y:        20,  // 从底部开始
				OffsetY:  220, // 向上延伸
			})
		}
	}

	// 添加Y轴标签（左侧）
	for i := 0; i <= 300; i += 50 {
		if i == 0 {
			continue
		}

		// Y轴标签
		optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
			Text:     fmt.Sprintf("%d", i),
			FontSize: 10,
			Color:    "#666666",
			X:        10,
			Y:        float64(i), // 从底部算
		})

		// 水平网格线
		if i < 300 {
			optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
				Text:     "—",
				FontSize: 8,
				Color:    "#dddddd",
				X:        20,
				Y:        float64(i),
				OffsetX:  420, // 向右延伸
			})
		}
	}

	// 添加坐标原点说明
	optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
		Text:     "坐标原点(0,0)",
		FontSize: 14,
		Color:    "#ff0000",
		X:        20,
		Y:        40, // 从底部算，接近底部
	})

	// 添加示例点
	examplePoints := []struct {
		x, y  float64
		color string
	}{
		{100, 100, "#ff0000"},
		{200, 150, "#00aa00"},
		{300, 200, "#0000ff"},
		{400, 250, "#ff00ff"},
	}

	for _, point := range examplePoints {
		// 点位置文本
		optionsGrid.ExtraTexts = append(optionsGrid.ExtraTexts, text2svg.ExtraTextInfo{
			Text:     fmt.Sprintf("● 点(%.0f,%.0f)", point.x, point.y),
			FontSize: 12,
			Color:    point.color,
			X:        point.x,
			Y:        point.y, // 从底部算
		})
	}

	_, err = text2svg.CanvasConvert(optionsGrid)
	if err != nil {
		t.Fatalf("生成坐标网格的SVG失败: %v", err)
	}
}
