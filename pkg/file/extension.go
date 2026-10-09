package file

var extensionDir = map[string]string{
	".css":  CSSFolder,
	".js":   JSFolder,
	".jpg":  ImgFolder,
	".jpeg": ImgFolder,
	".gif":  ImgFolder,
	".png":  ImgFolder,
	".webp": ImgFolder,
	".svg":  ImgFolder,
}

func FolderFromExtension(ext string) string {
	return extensionDir[ext]
}
