package text

import (
	"os"

	"github.com/ibryang/go-utils/canvas/common"
	"github.com/ibryang/go-utils/canvas/rect"
	"github.com/tdewolff/canvas"
)

// GenerateCanvasText 生成画布
func GenerateCanvasText(option common.CanvasOption) (*canvas.Canvas, error) {
	c := canvas.New(option.Width, option.Height)
	ctx := canvas.NewContext(c)
	for _, rectOption := range option.RectOption {
		rect.DrawRect(ctx, rectOption)
	}
	ctx.SetCoordSystem(canvas.CartesianIV)

	// 处理文件列表
	loadFileCanvasItem(&option)
	// 计算画布列表
	for _, canvasItem := range option.CanvasList {
		svgCanvas := canvasItem.Canvas
		// 计算缩放
		scaleX := 1.0
		scaleY := 1.0

		if canvasItem.Width == 0 && canvasItem.Height == 0 {
			canvasItem.Width = svgCanvas.W
			canvasItem.Height = svgCanvas.H
			// 最宽不能超过画布宽度
			if c.W/svgCanvas.W < 1 {
				scaleX = c.W / svgCanvas.W
				scaleY = scaleX
				canvasItem.Width = svgCanvas.W * scaleX
				canvasItem.Height = svgCanvas.H * scaleY
			}
			// 最高不能超过画布高度
			if c.H/svgCanvas.H < 1 {
				scaleY = c.H / svgCanvas.H
				scaleX = scaleY
				canvasItem.Width = svgCanvas.W * scaleX
				canvasItem.Height = svgCanvas.H * scaleY
			}
		} else if canvasItem.Width > 0 && canvasItem.Height > 0 {
			scaleX = canvasItem.Width / svgCanvas.W
			scaleY = canvasItem.Height / svgCanvas.H
		} else if canvasItem.Width == 0 && canvasItem.Height > 0 {
			scaleY = canvasItem.Height / svgCanvas.H
			scaleX = scaleY
			canvasItem.Width = svgCanvas.W * scaleX
		} else if canvasItem.Width > 0 && canvasItem.Height == 0 {
			scaleX = canvasItem.Width / svgCanvas.W
			scaleY = scaleX
			canvasItem.Height = svgCanvas.H * scaleY
		}

		var x, y float64
		// 计算对齐方式
		if canvasItem.Align == common.TextAlignLeft {
			x = 0
		} else if canvasItem.Align == common.TextAlignCenter {
			x = (c.W - canvasItem.Width) / 2
		} else if canvasItem.Align == common.TextAlignRight {
			x = c.W - canvasItem.Width
		}

		// 计算垂直对齐方式
		if canvasItem.VAlign == common.TextVAlignTop {
			y = 0
		} else if canvasItem.VAlign == common.TextVAlignCenter {
			y = (c.H - canvasItem.Height) / 2
		} else if canvasItem.VAlign == common.TextVAlignBottom {
			y = c.H - canvasItem.Height
		}

		if canvasItem.X != 0 {
			x += canvasItem.X
		}

		if canvasItem.Y != 0 {
			y += canvasItem.Y
		}

		// 渲染到画布
		svgCanvas.RenderViewTo(c, canvas.Matrix{
			{scaleX, 0, x},
			{0, scaleY, c.H - canvasItem.Height - y},
		})
	}

	// 绘制额外的文本
	for _, extOption := range option.ExtraText {
		DrawExtraText(ctx, extOption)
	}

	c = ReverseCanvas(c, option.ReverseX, option.ReverseY)

	return c, nil
}

func loadFileCanvasItem(option *common.CanvasOption) error {
	for _, file := range option.FileList {
		svgFile, err := os.Open(file.File)
		if err != nil {
			return err
		}

		svgCanvas, err := canvas.ParseSVG(svgFile)
		if err != nil {
			svgFile.Close()
			return err
		}
		svgFile.Close()

		option.CanvasList = append(option.CanvasList, common.CanvasItem{
			Canvas: svgCanvas,
			Align:  file.Align,
			VAlign: file.VAlign,
			Width:  file.Width,
			Height: file.Height,
			X:      file.X,
			Y:      file.Y,
		})
	}
	return nil
}
