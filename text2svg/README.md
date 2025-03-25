# text2svg

text2svg是一个用于将文本转换为SVG图像的Go语言库，支持丰富的格式化选项和输出格式。

## 功能特点

- 将文本转换为SVG、PNG、JPEG和PDF等多种格式
- 支持全局配置字体、颜色、尺寸和描边效果
- 支持自定义背景和圆角边框
- 灵活的内边距设置，类似CSS Padding
- 支持精确锁定最终尺寸（LockWidth/LockHeight）或保持比例缩放（Width/Height）
- 支持添加额外文本，可独立设置位置、旋转、字体和颜色

## 模块化结构

重构后的text2svg包包含以下模块：

- `text2svg.go`: 主模块，包含核心API和数据结构
- `canvas_generator.go`: 画布生成，负责文本到画布的转换逻辑
- `svg_handler.go`: SVG处理，包含SVG特有的处理逻辑
- `file_saver.go`: 文件保存，处理不同格式的输出保存
- `font.go`: 字体加载，管理字体的加载和处理
- `dimensions.go`: 尺寸计算，处理缩放和尺寸相关的计算
- `helper_funcs.go`: 辅助函数，提供位置相关的便捷函数

## 重构与修复说明

在重构过程中，解决了以下问题：

1. 对原有代码进行了模块化拆分，将超过1000行的单一文件分解为多个职责单一的模块
2. 修复了文本渲染位置计算问题，确保文本正确居中并考虑内边距
3. 统一了Convert和CanvasConvert方法的实现，使它们生成一致的SVG输出
4. 改进了错误处理和参数验证
5. 添加了单元测试以验证功能一致性

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/ibryang/go-utils/text2svg"
)

func main() {
    // 创建转换选项
    options := text2svg.Options{
        Text:     "Hello World",
        FontPath: "Arial",
        FontSize: 48,
        Colors:   []string{"#FF0000"},
        Width:    400,
        Height:   200,
        Padding:  []float64{20},
        EnableBackground: true,
        BackgroundColor:  "#FFFFFF",
        BorderRadius:     10,
        SavePath:         "output.svg",
    }

    // 转换并保存
    if err := text2svg.CanvasConvert(options); err != nil {
        fmt.Printf("转换失败: %v", err)
    }
}
```

## 许可证

MIT 