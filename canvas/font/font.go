package font

import (
	"fmt"
	"os"

	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
)

func LoadFont(path string) (*canvas.Font, error) {
	if !isExist(path) {
		font, err := LoadFontFamily(path)
		if err != nil {
			return nil, fmt.Errorf("加载字体失败: %s", err)
		}
		return font, nil
	}
	return canvas.LoadFontFile(path, canvas.FontBlack)
}

func LoadFontFamily(path string) (*canvas.Font, error) {
	font, err := canvas.LoadSystemFont(file.Name(path), canvas.FontBlack)
	if err != nil {
		return nil, err
	}
	return font, nil
}

func LoadFontLocal(path string) (*canvas.Font, error) {
	return canvas.LoadLocalFont(path, canvas.FontBlack)
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
