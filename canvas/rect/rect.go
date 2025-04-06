package rect

import (
	"fmt"

	"github.com/ibryang/go-utils/canvas/common"
	"github.com/tdewolff/canvas"
)

// GenerateRoundedRectPath 生成圆角矩形路径
func GenerateRoundedRectPath(width, height, radius float64) *canvas.Path {
	// 路径点定义
	type Point struct {
		X float64
		Y float64
	}

	points := []Point{
		{radius, 0}, // 开始点
		{width - radius, 0},
		{width, radius},
		{width, height - radius},
		{width - radius, height},
		{radius, height},
		{0, height - radius},
		{0, radius},
		{radius, 0}, // 闭合路径
	}

	// 构建路径数据
	var pathData string
	pathData += fmt.Sprintf("M %.6f %.6f ", points[0].X, points[0].Y)                          // 移动起点
	pathData += fmt.Sprintf("L %.6f %.6f ", points[1].X, points[1].Y)                          // 左上到右上直线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", width, 0.0, points[2].X, points[2].Y)    // 右上圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[3].X, points[3].Y)                          // 右线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", width, height, points[4].X, points[4].Y) // 右下圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[5].X, points[5].Y)                          // 下线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f ", 0.0, height, points[6].X, points[6].Y)   // 左下圆角
	pathData += fmt.Sprintf("L %.6f %.6f ", points[7].X, points[7].Y)                          // 左线
	pathData += fmt.Sprintf("Q %.6f %.6f %.6f %.6f Z", 0.0, 0.0, points[8].X, points[8].Y)     // 左上圆角并闭合
	path := canvas.MustParseSVGPath(pathData)
	return path
}

// DrawRect 在上下文中绘制矩形
func DrawRect(ctx *canvas.Context, rectOption common.RectOption) {
	// 处理矩形宽高
	if rectOption.Width <= 0 {
		rectOption.Width = ctx.Width()
	}

	if rectOption.Height <= 0 {
		rectOption.Height = ctx.Height()
	}

	// 设置背景色
	bgColor := canvas.Transparent
	if rectOption.BgColor != "" {
		if color, ok := common.ColorMap[rectOption.BgColor]; ok {
			bgColor = color
		} else {
			bgColor = canvas.Hex(rectOption.BgColor)
		}
	}

	ctx.SetFillColor(bgColor)

	// 绘制矩形路径
	var path *canvas.Path
	if rectOption.Radius > 0 {
		path = GenerateRoundedRectPath(rectOption.Width, rectOption.Height, rectOption.Radius)
	} else {
		path = canvas.Rectangle(rectOption.Width, rectOption.Height)
	}

	// 处理描边
	if rectOption.StrokeWidth > 0 {
		var strokeColor = canvas.Black
		if rectOption.StrokeColor != "" {
			if color, ok := common.ColorMap[rectOption.StrokeColor]; ok {
				strokeColor = color
			} else {
				strokeColor = canvas.Hex(rectOption.StrokeColor)
			}
		}
		ctx.SetStrokeColor(strokeColor)
		ctx.SetStrokeWidth(rectOption.StrokeWidth)
		ctx.Stroke()
	}
	ctx.DrawPath(rectOption.X, rectOption.Y, path)

	ctx.Fill()
}
