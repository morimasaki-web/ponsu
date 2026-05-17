package notify

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateBody string
		data         TemplateData
		want         string
		wantErr      bool
	}{
		{
			name:         "タイトルと申請者名を埋め込める",
			templateBody: "新しい申請が届きました: {{.Title}}（申請者: {{.SubmitterName}}）",
			data:         TemplateData{Title: "備品購入申請", SubmitterName: "山田太郎"},
			want:         "新しい申請が届きました: 備品購入申請（申請者: 山田太郎）",
		},
		{
			name:         "タイトルのみのテンプレート",
			templateBody: "申請が承認されました: {{.Title}}",
			data:         TemplateData{Title: "備品購入申請"},
			want:         "申請が承認されました: 備品購入申請",
		},
		{
			name:         "変数なしのテンプレート",
			templateBody: "通知が届きました",
			data:         TemplateData{},
			want:         "通知が届きました",
		},
		{
			name:         "不正なテンプレート構文はエラー",
			templateBody: "{{.Title",
			data:         TemplateData{Title: "テスト"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.templateBody, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	customBody := "カスタム通知: {{.Title}}"

	tests := []struct {
		name           string
		customTemplate *string
		eventType      string
		want           string
	}{
		{
			name:           "カスタムテンプレートがあればそちらを使う",
			customTemplate: &customBody,
			eventType:      "RequestSubmitted",
			want:           customBody,
		},
		{
			name:           "カスタムテンプレートがnilならデフォルトを使う",
			customTemplate: nil,
			eventType:      "RequestSubmitted",
			want:           defaultTemplates["RequestSubmitted"],
		},
		{
			name:           "カスタムテンプレートが空文字ならデフォルトを使う",
			customTemplate: strPtr(""),
			eventType:      "RequestApproved",
			want:           defaultTemplates["RequestApproved"],
		},
		{
			name:           "未知のイベントタイプはフォールバック文字列を返す",
			customTemplate: nil,
			eventType:      "UnknownEvent",
			want:           "通知: {{.Title}}",
		},
		{
			name:           "RequestRejectedのデフォルトテンプレートを使う",
			customTemplate: nil,
			eventType:      "RequestRejected",
			want:           defaultTemplates["RequestRejected"],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTemplate(tt.customTemplate, tt.eventType)
			if got != tt.want {
				t.Errorf("GetTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
