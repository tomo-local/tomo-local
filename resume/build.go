package main

// TODO: build.config.json を読み込み、sections の定義に従って README.md を生成するスクリプト
//
// 期待する動作:
//   1. build.config.json を読み込む
//   2. output に指定されたファイル (README.md) を生成先として開く
//   3. title を出力する
//   4. sections を順に処理する
//      - heading が定義されている場合は出力する
//      - file が定義されている場合はそのファイルの内容を出力する
//      - files が定義されている場合は各ファイルの内容を順に出力する
//        (企業間のセパレーター "---" をどう扱うかは任意)
//   5. すべてのセクションを結合して output ファイルに書き出す
//
// 実行方法 (想定):
//   cd doc && go run build.go

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Section struct {
	Heading string   `json:"heading"`
	File    string   `json:"file"`
	Files   []string `json:"files"`
}

type Config struct {
	Output   string    `json:"output"`
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

// loadConfig は指定したパスの JSON ファイルを読み込み Config を返す
func loadConfig(path string) (config *Config, err error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	config = &Config{}
	err = json.Unmarshal(content, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// build は Config に従ってマークダウンを生成し、文字列として返す
// baseDir は各 file/files パスの基準ディレクトリ
func build(config Config, baseDir string) (output string, err error) {
	outputs := []string{config.Title + "\n"}

	for _, section := range config.Sections {
		if section.Heading != "" {
			outputs = append(outputs, section.Heading+"\n")
		}
		if section.File != "" {
			content, err := os.ReadFile(baseDir + "/" + section.File)
			if err != nil {
				return "", err
			}
			outputs = append(outputs, string(content))
		}
		for _, file := range section.Files {
			content, err := os.ReadFile(baseDir + "/" + file)
			if err != nil {
				return "", err
			}
			outputs = append(outputs, string(content))
		}
	}

	output = ""
	for _, part := range outputs {
		output += part + "\n"
	}

	return output, nil
}

func main() {
	configPath := "build.config.json"

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	baseDir := filepath.Dir(configPath)

	output, err := build(*config, baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(config.Output, []byte(output), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %s\n", config.Output)
}
