package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func SearchCacheFiles(dir string, content string) ([]string, error) {
	var matchedFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(path)
			if err == nil {
				var cacheData map[string]interface{}
				if json.Unmarshal(data, &cacheData) == nil {
					if req, ok := cacheData["request"].(map[string]interface{}); ok {
						if messages, ok := req["messages"].([]interface{}); ok && len(messages) > 0 {
							if msg, ok := messages[0].(map[string]interface{}); ok {
								if msgContent, ok := msg["content"].(string); ok && strings.Contains(msgContent, content) {
									matchedFiles = append(matchedFiles, path)
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
	return matchedFiles, err
}

func DeleteCacheFilesByContent(dir string, content string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(path)
			if err == nil {
				var cacheData map[string]interface{}
				if json.Unmarshal(data, &cacheData) == nil {
					if req, ok := cacheData["request"].(map[string]interface{}); ok {
						if messages, ok := req["messages"].([]interface{}); ok && len(messages) > 0 {
							if msg, ok := messages[0].(map[string]interface{}); ok {
								if msgContent, ok := msg["content"].(string); ok && strings.Contains(msgContent, content) {
									return os.Remove(path)
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
}

func SearchCacheFilesWithPath(dir string, path string, value string) ([]string, error) {
	var matchedFiles []string
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var cacheData map[string]interface{}
				if json.Unmarshal(data, &cacheData) == nil {
					if matchJSONPath(cacheData, path, value) {
						matchedFiles = append(matchedFiles, filePath)
					}
				}
			}
		}
		return nil
	})
	return matchedFiles, err
}

func DeleteCacheFilesByContentWithPath(dir string, path string, value string) error {
	return filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var cacheData map[string]interface{}
				if json.Unmarshal(data, &cacheData) == nil {
					if matchJSONPath(cacheData, path, value) {
						return os.Remove(filePath)
					}
				}
			}
		}
		return nil
	})
}

func SearchCacheFilesContentWithPath(dir string, path string, value string, cursor uint64, pageSize int64) ([]map[string]interface{}, uint64, error) {
	matchedItems := []map[string]interface{}{} // 初始化为空数组
	var files []string

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filePath)
		}
		return nil
	})

	if err != nil {
		return matchedItems, 0, err
	}

	startIndex := int(cursor)
	endIndex := startIndex + int(pageSize)
	if endIndex > len(files) {
		endIndex = len(files)
	}

	for _, filePath := range files[startIndex:endIndex] {
		data, err := os.ReadFile(filePath)
		if err == nil {
			var cacheData map[string]interface{}
			if json.Unmarshal(data, &cacheData) == nil {
				if matchJSONPath(cacheData, path, value) {
					matchedItems = append(matchedItems, cacheData)
				}
			}
		}
	}

	nextCursor := uint64(endIndex)
	if endIndex >= len(files) {
		nextCursor = 0
	}

	return matchedItems, nextCursor, nil
}

func SearchCacheFilesContentWithMultiPath(dir string, contents []ContentQuery, cursor uint64, pageSize int64) ([]map[string]interface{}, uint64, error) {
	matchedItems := []map[string]interface{}{}
	var files []string

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filePath)
		}
		return nil
	})

	if err != nil {
		return matchedItems, 0, err
	}

	startIndex := int(cursor)
	endIndex := startIndex + int(pageSize)
	if endIndex > len(files) {
		endIndex = len(files)
	}

	for _, filePath := range files[startIndex:endIndex] {
		data, err := os.ReadFile(filePath)
		if err == nil {
			var cacheData map[string]interface{}
			if json.Unmarshal(data, &cacheData) == nil {
				if matchAllJSONPaths(cacheData, contents) {
					matchedItems = append(matchedItems, cacheData)
				}
			}
		}
	}

	nextCursor := uint64(endIndex)
	if endIndex >= len(files) {
		nextCursor = 0
	}

	return matchedItems, nextCursor, nil
}

func DeleteCacheFilesByContentWithMultiPath(dir string, contents []ContentQuery) (int, error) {
	deletedCount := 0
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var cacheData map[string]interface{}
				if json.Unmarshal(data, &cacheData) == nil {
					if matchAllJSONPaths(cacheData, contents) {
						err = os.Remove(filePath)
						if err != nil {
							return err
						}
						deletedCount++
					}
				}
			}
		}
		return nil
	})
	return deletedCount, err
}
