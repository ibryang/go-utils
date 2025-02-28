package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/changedpi"
)

func TestChangeDpi(t *testing.T) {
	output, err := changedpi.ChangeDpiByPath("./text2svg_colors.png", 300)
	if err != nil {
		t.Error(err)
	}
	err = changedpi.SaveImage("example300.png", output)
	if err != nil {
		t.Error(err)
	}
}
