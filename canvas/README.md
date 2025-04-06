# Canvas文本处理库

这个库提供了一套用于处理文本渲染、画布合成和图形生成的工具。它基于 [tdewolff/canvas](https://github.com/tdewolff/canvas) 库进行了封装和扩展，提供了更高级的API。

## 功能特点

- 文本渲染：支持多种字体、颜色和样式
- 矩形处理：支持圆角、填充色和描边
- 文本行布局：支持多行文本、对齐方式和行间距
- 画布合成：支持将多个元素组合到一个画布
- 灵活的缩放和转换
- 多种输出格式：SVG、PDF等

## 目录结构

```
canvas/
├── util/          # 通用工具和基础结构
├── rect/          # 矩形相关功能
├── text/          # 文本处理功能
└── examples/      # 使用示例
```

## 使用示例

### 生成简单文本

```go
textOption := text.TextOption{
    Text:      "Hello 你好",
    FontPath:  "Arial",
    FontSize:  48,
    FontColor: "blue",
    BaseOption: util.BaseOption{
        Width: 300,
    },
    RectOption: &rect.RectOption{
        BgColor:     "yellow",
        StrokeColor: "red",
        StrokeWidth: 2,
        Radius:      10,
    },
}

canvas, err := text.GenerateTextSvg(textOption)
if err != nil {
    log.Fatalf("生成文本失败: %v", err)
}
renderers.Write("text_example.svg", canvas)
```

### 生成文本行

```go
textLineOption := text.TextLineOption{
    TextList: []text.TextOption{
        {
            Text:      "第一行",
            FontPath:  "Arial",
            FontSize:  36,
            FontColor: "red",
        },
        {
            Text:      "第二行文本",
            FontPath:  "Arial",
            FontSize:  24,
            FontColor: "blue",
        },
    },
    Padding:    [4]float64{10, 10, 10, 10}, // 上、右、下、左
    LineGap:    5,                           // 行间距
    Align:      util.TextAlignCenter,        // 居中对齐
    BaseOption: util.BaseOption{
        Width: 400,
    },
    RectOption: []rect.RectOption{
        {
            BgColor:     "white",
            StrokeColor: "black",
            StrokeWidth: 1,
            Radius:      5,
        },
    },
}

canvas, err := text.GenerateTextLine(textLineOption)
if err != nil {
    log.Fatalf("生成文本行失败: %v", err)
}
renderers.Write("text_line_example.svg", canvas)
```

### 组合多个画布

```go
canvasOption := text.CanvasOption{
    CanvasList: []util.CanvasItem{
        {
            Canvas: canvas1,
            Width:  300,
            X:      50,
            Y:      50,
        },
        {
            Canvas: canvas2,
            Width:  400,
            X:      50,
            Y:      200,
        },
    },
    BaseOption: util.BaseOption{
        Width:  500,
        Height: 500,
    },
    Padding: [4]float64{20, 20, 20, 20}, // 上、右、下、左
}

canvas, err := text.GenerateCanvas(canvasOption)
if err != nil {
    log.Fatalf("生成组合画布失败: %v", err)
}
renderers.Write("combined_example.svg", canvas)
renderers.Write("combined_example.pdf", canvas)
```

## 运行示例

```bash
# 进入项目目录
cd go-utils

# 运行示例
go run canvas/examples/main.go
```

生成的文件将保存在 `canvas/examples/` 目录下。 