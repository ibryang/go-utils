package text2svg

// calculateDimensions 计算最终尺寸和缩放比例
func calculateDimensions(originalWidth, originalHeight, targetWidth, targetHeight float64) (width, height, scaleX, scaleY float64) {
	if targetWidth <= 0 && targetHeight <= 0 {
		return originalWidth, originalHeight, 1, 1
	}

	originalRatio := originalWidth / originalHeight

	if targetWidth > 0 && targetHeight <= 0 {
		width = targetWidth
		height = targetWidth / originalRatio
		scaleX = targetWidth / originalWidth
		scaleY = scaleX
		return
	}

	targetRatio := targetWidth / targetHeight

	if originalRatio > targetRatio {
		width = targetWidth
		height = targetHeight
		scaleX = targetWidth / originalWidth
		heightScale := targetHeight / (originalHeight * scaleX)
		scaleY = scaleX * heightScale
	} else {
		width = targetWidth
		height = targetHeight
		scaleY = targetHeight / originalHeight
		widthScale := targetWidth / (originalWidth * scaleY)
		scaleX = scaleY * widthScale
	}

	return
}
