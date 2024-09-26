package common

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

type ContentQuery struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func RedisKeys(pattern string) ([]string, error) {
	ctx := context.Background()
	return RDB.Keys(ctx, pattern).Result()
}

func RedisDelPrefix(prefix string) error {
	ctx := context.Background()
	iter := RDB.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		err := RDB.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}

func RedisKeysPaginated(pattern string, cursor uint64, pageSize int64) ([]string, uint64, error) {
	ctx := context.Background()
	keys, nextCursor, err := RDB.Scan(ctx, cursor, pattern, pageSize).Result()
	return keys, nextCursor, err
}

func RedisSearchKeys(pattern string, content string, cursor uint64, pageSize int64) ([]string, uint64, error) {
	ctx := context.Background()
	var matchedKeys []string
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return nil, 0, err
		}

		for _, key := range keys {
			value, err := RDB.Get(ctx, key).Result()
			if err == nil && strings.Contains(value, content) {
				matchedKeys = append(matchedKeys, key)
			}
		}

		if nextCursor == 0 || int64(len(matchedKeys)) >= pageSize {
			return matchedKeys, nextCursor, nil
		}
		cursor = nextCursor
	}
}

func RedisDelByContent(prefix string, content string) error {
	ctx := context.Background()
	var cursor uint64 = 0
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			value, err := RDB.Get(ctx, key).Result()
			if err == nil && strings.Contains(value, content) {
				err = RDB.Del(ctx, key).Err()
				if err != nil {
					return err
				}
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

func RedisDelByPrefix(prefix string) error {
	ctx := context.Background()
	iter := RDB.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		err := RDB.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}

func RedisSearchKeysWithPath(pattern string, path string, value string, cursor uint64, pageSize int64) ([]string, uint64, error) {
	ctx := context.Background()
	var matchedKeys []string
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return nil, 0, err
		}

		for _, key := range keys {
			cacheData, err := RDB.Get(ctx, key).Result()
			if err == nil {
				var data map[string]interface{}
				if json.Unmarshal([]byte(cacheData), &data) == nil {
					if matchJSONPath(data, path, value) {
						matchedKeys = append(matchedKeys, key)
					}
				}
			}
		}

		if nextCursor == 0 || int64(len(matchedKeys)) >= pageSize {
			return matchedKeys, nextCursor, nil
		}
		cursor = nextCursor
	}
}

func RedisDelByContentWithPath(prefix string, path string, value string) error {
	ctx := context.Background()
	var cursor uint64 = 0
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			cacheData, err := RDB.Get(ctx, key).Result()
			if err == nil {
				var data map[string]interface{}
				if json.Unmarshal([]byte(cacheData), &data) == nil {
					if matchJSONPath(data, path, value) {
						err = RDB.Del(ctx, key).Err()
						if err != nil {
							return err
						}
					}
				}
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

func matchJSONPath(data map[string]interface{}, path string, value interface{}) bool {
	parts := strings.Split(path, ".")
	current := data
	for i, part := range parts {
		if strings.HasSuffix(part, "]") {
			arrayPart := strings.Split(part, "[")
			indexStr := strings.TrimRight(arrayPart[1], "]")
			index, _ := strconv.Atoi(indexStr)
			if arr, ok := current[arrayPart[0]].([]interface{}); ok && index < len(arr) {
				if i == len(parts)-1 {
					return reflect.DeepEqual(arr[index], value)
				}
				if mapValue, ok := arr[index].(map[string]interface{}); ok {
					current = mapValue
				} else {
					return false
				}
			} else {
				return false
			}
		} else {
			if i == len(parts)-1 {
				return reflect.DeepEqual(current[part], value)
			}
			if mapValue, ok := current[part].(map[string]interface{}); ok {
				current = mapValue
			} else {
				return false
			}
		}
	}
	return false
}

func RedisSearchCacheWithPath(pattern string, path string, value string, cursor uint64, pageSize int64) ([]map[string]interface{}, uint64, error) {
	ctx := context.Background()
	matchedItems := []map[string]interface{}{} // 初始化为空数组
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return matchedItems, 0, err
		}

		for _, key := range keys {
			cacheData, err := RDB.Get(ctx, key).Result()
			if err == nil {
				var data map[string]interface{}
				if json.Unmarshal([]byte(cacheData), &data) == nil {
					if matchJSONPath(data, path, value) {
						matchedItems = append(matchedItems, data)
					}
				}
			}
		}

		if nextCursor == 0 || int64(len(matchedItems)) >= pageSize {
			return matchedItems, nextCursor, nil
		}
		cursor = nextCursor
	}
}
func RedisSearchCacheWithMultiPath(pattern string, contents []ContentQuery, cursor uint64, pageSize int64) ([]map[string]interface{}, uint64, error) {
	ctx := context.Background()
	matchedItems := []map[string]interface{}{}
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return matchedItems, 0, err
		}

		for _, key := range keys {
			cacheData, err := RDB.Get(ctx, key).Result()
			if err == nil {
				var data map[string]interface{}
				if json.Unmarshal([]byte(cacheData), &data) == nil {
					if matchAllJSONPaths(data, contents) {
						matchedItems = append(matchedItems, data)
					}
				}
			}
		}

		if nextCursor == 0 || int64(len(matchedItems)) >= pageSize {
			return matchedItems, nextCursor, nil
		}
		cursor = nextCursor
	}
}

func RedisDelByContentWithMultiPath(prefix string, contents []ContentQuery) (int, error) {
	ctx := context.Background()
	var cursor uint64 = 0
	deletedCount := 0
	for {
		keys, nextCursor, err := RDB.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return deletedCount, err
		}

		for _, key := range keys {
			cacheData, err := RDB.Get(ctx, key).Result()
			if err == nil {
				var data map[string]interface{}
				if json.Unmarshal([]byte(cacheData), &data) == nil {
					if matchAllJSONPaths(data, contents) {
						err = RDB.Del(ctx, key).Err()
						if err != nil {
							return deletedCount, err
						}
						deletedCount++
					}
				}
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return deletedCount, nil
}

func matchAllJSONPaths(data map[string]interface{}, contents []ContentQuery) bool {
	for _, content := range contents {
		if !matchJSONPath(data, content.Path, content.Value) {
			return false
		}
	}
	return true
}
