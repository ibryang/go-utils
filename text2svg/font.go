package text2svg

import (
	"fmt"
	"os"

	"github.com/ibryang/go-utils/os/file"
	"github.com/tdewolff/canvas"
)

// fontLoader 接口定义字体加载行为
type fontLoader interface {
	load(path string) (*canvas.Font, error)
}

// fileFontLoader 实现从文件加载字体
type fileFontLoader struct{}

func (l *fileFontLoader) load(path string) (*canvas.Font, error) {
	if !isExist(path) {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}
	return canvas.LoadFontFile(path, canvas.FontBlack)
}

// systemFontLoader 实现从系统加载字体
type systemFontLoader struct{}

func (l *systemFontLoader) load(path string) (font *canvas.Font, err error) {
	font, err = canvas.LoadSystemFont(path, canvas.FontBlack)
	if err != nil {
		return canvas.LoadSystemFont(file.Name(path), canvas.FontBlack)
	}
	return
}

// fontManager 管理字体加载
type fontManager struct {
	loaders []fontLoader
}

func newFontManager() *fontManager {
	return &fontManager{
		loaders: []fontLoader{
			&fileFontLoader{},
			&systemFontLoader{},
		},
	}
}

func (fm *fontManager) loadFont(path string) (*canvas.Font, error) {
	var lastErr error
	for _, loader := range fm.loaders {
		font, err := loader.load(path)
		if err == nil {
			return font, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("所有加载方式均失败，最后错误: %v", lastErr)
}

func loadFontFamily(fontPath string) (*canvas.Font, error) {
	fontManager := newFontManager()
	return fontManager.loadFont(fontPath)
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
