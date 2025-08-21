package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/changedpi"
)

func TestChangeDpi(t *testing.T) {
	changedpi.ChangeDpi("./text2svg_colors.png", "./text2svg_colors.png", 300)
}
