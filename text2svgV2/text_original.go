package text2svgV2

import (
	"errors"
	"fmt"

	"github.com/tdewolff/canvas"
)

func GenerateOriginalText(option TextOption) (*canvas.Canvas, error) {
	if option.Text == "" {
		return nil, errors.New("text is required")
	}

	font, err := LoadFont(option.FontPath)
	if err != nil {
		return nil, err
	}
	fontface := font.Face(option.FontSize, canvas.Hex(option.FontColor.(string)))

	// txt := canvas.NewTextLine(fontface, option.Text, canvas.Center)
	txt := canvas.NewTextBox(fontface, option.Text, 0, 0, canvas.Justify, canvas.Center, 0, 0)

	bounds := txt.OutlineBounds()
	c := canvas.New(bounds.W(), bounds.H())
	ctx := canvas.NewContext(c)
	ctx.SetCoordSystem(canvas.CartesianIII)
	// ctx.DrawText(-bounds.X0, -bounds.Y0, txt)
	fmt.Println("bounds", bounds)
	ctx.DrawText(bounds.X1, bounds.Y1, txt)

	return c, nil
}
