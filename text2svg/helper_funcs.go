package text2svg

// 辅助函数，用于在常见位置放置文本
// 坐标系统说明：
// - X轴：使用左侧为原点(0)，向右增加
// - Y轴：使用底部为原点(0)，向上增加

// AddTextAtBottomLeft 在左下角添加文本
func AddTextAtBottomLeft(text string, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    padding,
		Y:    padding, // Y=0在底部
	}
}

// AddTextAtBottomCenter 在底部中央添加文本
func AddTextAtBottomCenter(text string, width, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    width / 2,
		Y:    padding, // Y=0在底部
	}
}

// AddTextAtBottomRight 在右下角添加文本
func AddTextAtBottomRight(text string, width, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    width - padding,
		Y:    padding, // Y=0在底部
	}
}

// AddTextAtCenter 在中心添加文本
func AddTextAtCenter(text string, width, height float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    width / 2,
		Y:    height / 2, // 中间高度
	}
}

// AddTextAtTopLeft 在左上角添加文本
func AddTextAtTopLeft(text string, width, height, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    padding,
		Y:    height - padding, // Y=height在顶部
	}
}

// AddTextAtTopCenter 在顶部中央添加文本
func AddTextAtTopCenter(text string, width, height, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    width / 2,
		Y:    height - padding, // Y=height在顶部
	}
}

// AddTextAtTopRight 在右上角添加文本
func AddTextAtTopRight(text string, width, height, padding float64) ExtraTextInfo {
	return ExtraTextInfo{
		Text: text,
		X:    width - padding,
		Y:    height - padding, // Y=height在顶部
	}
}
