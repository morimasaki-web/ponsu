// Package config は環境変数からアプリ設定を読み込み、接続情報などを組み立てる。
// このファイルはローカル開発向けに .env/.env.local を読み込む補助を提供する。
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func dotenvPathsFromEnv() []string {
	v := strings.TrimSpace(os.Getenv("PONSU_DOTENV_FILES"))
	if v == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ';'
	})
	paths := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		paths = append(paths, p)
	}
	return paths
}

func readDotenvFile(path string) (map[string]string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Windows の「UTF-8 (BOM付き)」で保存された .env 系を許容する。
	src = bytes.TrimPrefix(src, utf8BOM)

	m, err := godotenv.UnmarshalBytes(src)
	if err != nil {
		// godotenv のエラーは入力の断片を含むことがあるため、内容をログに出さない。
		return nil, fmt.Errorf("invalid dotenv file: %s", path)
	}
	return m, nil
}

// LoadDotenvLocal はカレントディレクトリの .env と .env.local を読み込む。
// 読み込み順は .env → .env.local で、同じキーは .env.local を優先する。
// ただし、OS環境変数として既に設定されているキーは上書きしない。
//
// PONSU_DOTENV_FILES が設定されている場合は、そのパス一覧（カンマ/セミコロン区切り）を順に読み込む。
// この場合、デフォルトの .env/.env.local は読み込まない（秘密情報をワークスペース外へ置けるようにするため）。
//
// 戻り値 loaded には、実際に読み込めたファイル名（例: .env.local）が入る。
func LoadDotenvLocal() (loaded []string, err error) {
	paths := dotenvPathsFromEnv()
	if len(paths) == 0 {
		paths = []string{".env", ".env.local"}
	}

	merged := map[string]string{}
	for _, path := range paths {
		m, readErr := readDotenvFile(path)
		if readErr != nil {
			if errors.Is(readErr, fs.ErrNotExist) {
				continue
			}
			// godotenv は fs.ErrNotExist を包まないケースもあるため補助判定
			if os.IsNotExist(readErr) {
				continue
			}
			return loaded, readErr
		}
		for k, v := range m {
			merged[k] = v
		}
		loaded = append(loaded, path)
	}

	for k, v := range merged {
		if _, ok := os.LookupEnv(k); ok {
			continue
		}
		_ = os.Setenv(k, v)
	}

	return loaded, nil
}
