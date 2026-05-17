package notify

import (
	"bytes"
	"fmt"
	"text/template"
)

// デフォルトテンプレート（組織がカスタマイズしていない場合）
var defaultTemplates = map[string]string{
	"RequestSubmitted": "新しい申請が届きました: {{.Title}}（申請者: {{.SubmitterName}}）",
	"RequestApproved":  "申請が承認されました: {{.Title}}",
	"RequestRejected":  "申請が却下されました: {{.Title}}",
}

// TemplateData はテンプレートに渡すデータ
type TemplateData struct {
	Title         string
	SubmitterName string
}

// RenderTemplate はテンプレート文字列にデータを埋め込む
func RenderTemplate(templateBody string, data TemplateData) (string, error) {
	tmpl, err := template.New("notification").Parse(templateBody)
	if err != nil {
		return "", fmt.Errorf("テンプレートのパースに失敗: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("テンプレートの実行に失敗: %w", err)
	}

	return buf.String(), nil
}

// GetTemplate はカスタムテンプレートがなければデフォルトを返す
func GetTemplate(customTemplate *string, eventType string) string {
	if customTemplate != nil && *customTemplate != "" {
		return *customTemplate
	}
	if def, ok := defaultTemplates[eventType]; ok {
		return def
	}
	return "通知: {{.Title}}"
}
