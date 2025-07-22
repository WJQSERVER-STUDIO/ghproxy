package auth

import (
	"fmt"
	"ghproxy/config"
	"os"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func ReadIPFilterList(cfg *config.Config) (whitelist []string, blacklist []string, err error) {
	if cfg.IPFilter.IPFilterFile == "" {
		return nil, nil, nil
	}

	// 检查文件是否存在, 不存在则创建空json
	if _, err := os.Stat(cfg.IPFilter.IPFilterFile); os.IsNotExist(err) {
		if err := CreateEmptyIPFilterFile(cfg.IPFilter.IPFilterFile); err != nil {
			return nil, nil, fmt.Errorf("failed to create empty IP filter file: %w", err)
		}
	}

	data, err := os.ReadFile(cfg.IPFilter.IPFilterFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read IP filter file: %w", err)
	}

	var ipFilterData struct {
		AllowList []string `json:"allow"`
		BlockList []string `json:"block"`
	}
	if err := json.Unmarshal(data, &ipFilterData); err != nil {
		return nil, nil, fmt.Errorf("invalid IP filter file format: %w", err)
	}

	return ipFilterData.AllowList, ipFilterData.BlockList, nil
}

// 创建空列表json
func CreateEmptyIPFilterFile(filePath string) error {
	emptyData := struct {
		AllowList []string `json:"allow"`
		BlockList []string `json:"block"`
	}{
		AllowList: []string{},
		BlockList: []string{},
	}

	jsonData, err := json.Marshal(emptyData, jsontext.Multiline(true), jsontext.WithIndent("  "))
	if err != nil {
		return fmt.Errorf("failed to marshal empty IP filter data: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write empty IP filter file: %w", err)
	}
	return nil
}
