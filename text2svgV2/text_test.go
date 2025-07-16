package text2svgV2

import (
	"testing"

	"github.com/tdewolff/canvas/renderers"
)

// 弧形文字

func TestGenerateTextSvg(t *testing.T) {
	option := TextOption{
		Text:      `Universal`,
		FontPath:  "Cookie",
		FontSize:  100,
		FontColor: "red",
		// StrokeColor: "blue",
		// StrokeWidth: .1,
		BaseOption: BaseOption{
			MinSize: true,
			Width:   400,
			Height:  80,
		},
		// ExtraText: []ExtraTextOption{
		// 	{
		// 		X:      10,
		// 		Y:      1,
		// 		Align:  TextAlignCenter,
		// 		VAlign: TextVAlignTop,
		// 		TextOption: TextOption{
		// 			Text:      "001",
		// 			FontSize:  10,
		// 			FontColor: "blue",
		// 		},
		// 	},
		// },
		RenderMode: RenderChar,
	}

	canvas, err := GenerateBaseText(option)
	if err != nil {
		t.Fatalf("生成文本SVG失败: %v", err)
	}

	// 保存为SVG文件
	renderers.Write("text_test.png", canvas)
	// GroupSvg(canvas, "text_test.png")
}

func TestGenerateTextLine(t *testing.T) {
	option := TextLineOption{
		TextList: []TextOption{
			{
				Text:      "Beijing",
				FontPath:  "Cookie",
				FontSize:  100,
				FontColor: []string{"blue", "red", "green"},
			},
			// {
			// 	Text:      "Beijing",
			// 	FontPath:  "Cookie",
			// 	FontSize:  20,
			// 	FontColor: []string{"blue", "red", "green"},
			// },
			// {
			// 	Text:      "Beijing",
			// 	FontPath:  "Cookie",
			// 	FontSize:  16,
			// 	FontColor: []string{"blue", "red", "green"},
			// },
		},
		LineGap: 1,
		Align:   TextAlignCenter,
		VAlign:  TextVAlignCenter,
		BaseOption: BaseOption{
			// ReverseX: true,
			// ReverseY: true,
			Height:    80,
			Width:     100,
			LockRatio: true,
		},
		// RectOption: []RectOption{
		// 	{
		// 		StrokeColor: "red",
		// 		StrokeWidth: 0.1,
		// 		Radius:      5,
		// 	},
		// },
	}

	canvas, err := GenerateMultipleLinesText(option)
	if err != nil {
		t.Fatalf("生成文本行失败: %v", err)
	}

	// 保存为PDF文件
	renderers.Write("text_line_test.png", canvas)
	// GroupSvg(canvas, "text_line_test2.svg")
}

func TestGenerateCanvas(t *testing.T) {
	textOption := TextOption{
		Text:      "Benjamin222",
		FontPath:  "Cookie",
		FontSize:  100,
		FontColor: "red",
		BaseOption: BaseOption{
			Width: 200,
		},
	}

	textCanvas, err := GenerateBaseText(textOption)
	if err != nil {
		t.Fatalf("生成文本SVG失败: %v", err)
	}

	option := CanvasOption{
		FileList: []CanvasItem{
			{
				File:   "text_test.svg",
				Width:  100,
				VAlign: TextVAlignCenter,
			},
		},
		CanvasList: []CanvasItem{
			{
				Canvas: textCanvas,
				Width:  100,
				Align:  TextAlignCenter,
			},
		},
		BaseOption: BaseOption{
			Width:    300,
			Height:   300,
			ReverseX: true,
			ReverseY: true,
		},
		RectOption: []RectOption{
			{
				StrokeColor: "red",
				StrokeWidth: .1,
				Radius:      5,
			},
		},
	}

	canvas, err := GenerateCanvasText(option)
	if err != nil {
		t.Fatalf("生成画布失败: %v", err)
	}

	renderers.Write("canvas_test.pdf", canvas)
}

func TestGenerateOriginalText(t *testing.T) {
	option := TextOption{
		Text: `Dr. Azmel Ainul
Quality Engineer
Dynamic Controls Ltd`,
		FontPath:  "Cookie",
		FontSize:  500,
		FontColor: "#0000FF",
	}
	canvas, err := GenerateOriginalText(option)
	if err != nil {
		t.Fatalf("生成原始文本失败: %v", err)
	}
	renderers.Write("text_original.png", canvas)
}
