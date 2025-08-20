package text2svgV2

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
)

func LoadFont(path string) (*canvas.Font, error) {
	if path == "" {
		// 加载系统默认字体
		if runtime.GOOS == "windows" {
			font, err := LoadFontFamily("Microsoft YaHei")
			if err != nil {
				return nil, fmt.Errorf("加载字体失败: %s", err)
			}
			return font, nil
		}
		font, err := LoadFontFamily("Arial")
		if err != nil {
			return nil, fmt.Errorf("加载字体失败: %s", err)
		}
		return font, nil
	}
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
	font, err := canvas.LoadSystemFont(file.Name(path), canvas.FontStyle(canvas.FontNormal))
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
