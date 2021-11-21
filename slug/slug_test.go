package slug

import (
	"testing"
)

func Test_slug(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{name: "Replace space with -", arg: "test title slug", want: "test-title-slug"},
		{name: "Replace special char with -", arg: "test&title*~slug", want: "test-title-slug"},
		{name: "Convert uppercase to lowercase", arg: "TestSlug", want: "testslug"},
		{name: "Convert other languages", arg: "测试标题", want: "ce-shi-biao-ti"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slug(tt.arg); got != tt.want {
				t.Errorf("slug() = %v, want %v", got, tt.want)
			}
		})
	}
}
