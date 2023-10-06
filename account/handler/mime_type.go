package handler

var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// isAllowedImageType determine if the mime type is allowed
func isAllowedImageType(mimeType string) bool {
	_, exists := validImageTypes[mimeType]

	return exists
}
