package utils

import (
	"strings"
)

func IsValidImage(imageName string, imageList []string) bool {
	for _, _nekoImage := range imageList {
		if _nekoImage == imageName {
			return true
		}
		if strings.Contains(_nekoImage, "*") {
			pattern := strings.Replace(_nekoImage, "*", "", -1)
			if strings.Contains(imageName, pattern) {
				return true
			}
		}
	}
	return false
}