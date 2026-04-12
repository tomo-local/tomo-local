package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// --- loadConfig のテスト ---

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(t *testing.T, dir string) string // config ファイルのパスを返す
		wantErr bool
		want    *Config
	}{
		{
			name: "フィールドを正しくパースする",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				path := filepath.Join(dir, "build.config.json")
				writeJSON(t, path, Config{
					Output: "README.md",
					Title:  "# 職務経歴書",
					Sections: []Section{
						{Heading: "## 職務要約", File: "parts/summary.md"},
					},
				})
				return path
			},
			want: &Config{
				Output: "README.md",
				Title:  "# 職務経歴書",
				Sections: []Section{
					{Heading: "## 職務要約", File: "parts/summary.md"},
				},
			},
		},
		{
			name: "filesの配列をパースする",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				path := filepath.Join(dir, "build.config.json")
				writeJSON(t, path, Config{
					Output: "README.md",
					Title:  "# Title",
					Sections: []Section{
						{
							Heading: "## 職務経歴 (詳細)",
							Files:   []string{"companies/a.md", "companies/b.md"},
						},
					},
				})
				return path
			},
			want: &Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{
						Heading: "## 職務経歴 (詳細)",
						Files:   []string{"companies/a.md", "companies/b.md"},
					},
				},
			},
		},
		{
			name: "ファイルが存在しない場合はエラーを返す",
			setup: func(t *testing.T, dir string) string {
				return "/nonexistent/path/build.config.json"
			},
			wantErr: true,
		},
		{
			name: "不正なJSONの場合はエラーを返す",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				path := filepath.Join(dir, "build.config.json")
				os.WriteFile(path, []byte("{ invalid json"), 0644)
				return path
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := tc.setup(t, dir)

			got, err := loadConfig(path)

			if tc.wantErr {
				if err == nil {
					t.Error("error を期待したが nil だった")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しない error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got  %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

// --- build のテスト ---

// 期待する出力フォーマット:
//   - title、heading、ファイル内容をそれぞれ \n\n で結合
//   - 末尾に \n を付ける

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		files   map[string]string // dir 内に作成するファイル (相対パス → 内容)
		config  Config
		wantErr bool
		want    string
	}{
		{
			name:  "セクションなし: titleのみ出力される",
			files: map[string]string{},
			config: Config{
				Output:   "README.md",
				Title:    "# 職務経歴書",
				Sections: []Section{},
			},
			want: "# 職務経歴書\n\n",
		},
		{
			name:  "heading + file: 見出しとファイル内容が順に出力される",
			files: map[string]string{"summary.md": "要約の内容です"},
			config: Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{Heading: "## 職務要約", File: "summary.md"},
				},
			},
			want: "# Title\n\n## 職務要約\n\n要約の内容です\n",
		},
		{
			name: "files配列: 全ファイルが定義順に出力される",
			files: map[string]string{
				"company_a.md": "### 企業A\n内容A",
				"company_b.md": "### 企業B\n内容B",
			},
			config: Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{
						Heading: "## 職務経歴 (詳細)",
						Files:   []string{"company_a.md", "company_b.md"},
					},
				},
			},
			want: "# Title\n\n## 職務経歴 (詳細)\n\n### 企業A\n内容A\n### 企業B\n内容B\n",
		},
		{
			name:  "headingなし: ファイル内容のみ出力される",
			files: map[string]string{"content.md": "見出しなしの内容"},
			config: Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{File: "content.md"},
				},
			},
			want: "# Title\n\n見出しなしの内容\n",
		},
		{
			name: "複数セクション: 定義順に出力される",
			files: map[string]string{
				"summary.md": "要約",
				"skills.md":  "スキル",
			},
			config: Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{Heading: "## 職務要約", File: "summary.md"},
					{Heading: "## スキルレベル", File: "skills.md"},
				},
			},
			want: "# Title\n\n## 職務要約\n\n要約\n## スキルレベル\n\nスキル\n",
		},
		{
			name:  "存在しないfileはエラーを返す",
			files: map[string]string{},
			config: Config{
				Output: "README.md",
				Title:  "# Title",
				Sections: []Section{
					{Heading: "## Section", File: "nonexistent.md"},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tc.files {
				writeFile(t, filepath.Join(dir, name), content)
			}

			got, err := build(tc.config, dir)

			if tc.wantErr {
				if err == nil {
					t.Error("error を期待したが nil だった")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しない error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

// --- ヘルパー ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("writeJSON marshal: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("writeJSON write: %v", err)
	}
}
