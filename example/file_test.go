package example

import (
	"fmt"
	"testing"

	"github.com/ibryang/go-utils/os/file"
)

func TestFile(t *testing.T) {
	ext := file.ExtName("text2svg_colors.png")
	fmt.Println(ext)
}
