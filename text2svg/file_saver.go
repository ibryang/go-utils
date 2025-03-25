package text2svg

import (
	"fmt"

	"github.com/ibryang/go-utils/changedpi"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

// saveToFile 保存画布到文件
func saveToFile(c *canvas.Canvas, config SaveConfig) error {
	if config.Path == "" {
		return fmt.Errorf("保存路径不能为空")
	}

	// 设置默认值
	if config.DPI == 0 {
		config.DPI = 72
	}
	if config.DPMM == 0 {
		config.DPMM = 2.8346456692913385
	}
	if config.Quality == 0 {
		config.Quality = 80
	}

	// 根据不同格式保存
	switch config.Format {
	case FormatPNG:
		if err := savePNG(c, config); err != nil {
			return err
		}
	case FormatJPEG, FormatJPG:
		if err := saveJPEG(c, config); err != nil {
			return err
		}
	case FormatSVG:
		if err := c.WriteFile(config.Path, renderers.SVG()); err != nil {
			return fmt.Errorf("保存SVG文件失败: %v", err)
		}
	case FormatPDF:
		if err := c.WriteFile(config.Path, renderers.PDF()); err != nil {
			return fmt.Errorf("保存PDF文件失败: %v", err)
		}
	case FormatTIFF, FormatTIF:
		if err := c.WriteFile(config.Path, renderers.TIFF()); err != nil {
			return fmt.Errorf("保存TIFF文件失败: %v", err)
		}
	default:
		return fmt.Errorf("不支持的文件格式: %s", config.Format)
	}

	return nil
}

// savePNG 保存PNG格式
func savePNG(c *canvas.Canvas, config SaveConfig) error {
	if err := c.WriteFile(config.Path, renderers.PNG(canvas.DPI(config.DPI))); err != nil {
		return fmt.Errorf("保存PNG文件失败: %v", err)
	}

	// 如果DPI不是72，需要更新DPI信息
	if config.DPI != 72 {
		return updateImageDPI(config.Path, int(config.DPI))
	}
	return nil
}

// saveJPEG 保存JPEG格式
func saveJPEG(c *canvas.Canvas, config SaveConfig) error {
	if err := c.WriteFile(config.Path, renderers.JPEG(canvas.DPI(config.DPI), config.Quality)); err != nil {
		return fmt.Errorf("保存JPEG文件失败: %v", err)
	}

	// 如果DPI不是72，需要更新DPI信息
	if config.DPI != 72 {
		return updateImageDPI(config.Path, int(config.DPI))
	}
	return nil
}

// updateImageDPI 更新图片DPI信息
func updateImageDPI(path string, dpi int) error {
	baseData, err := changedpi.ChangeDpiByPath(path, dpi)
	if err != nil {
		return fmt.Errorf("更新DPI失败: %v", err)
	}

	if err := changedpi.SaveImage(path, baseData); err != nil {
		return fmt.Errorf("保存更新DPI后的图片失败: %v", err)
	}
	return nil
}
