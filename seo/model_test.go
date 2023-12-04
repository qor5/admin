package seo

import (
	"context"
	"testing"
)

func TestSettingHTMLComponent(t *testing.T) {
	tests := []struct {
		name    string
		setting Setting
		tags    map[string]string
		want    string
	}{
		{
			name: "Render the SEO html",
			setting: Setting{
				Title:                "title",
				Description:          "description",
				Keywords:             "keyword",
				OpenGraphTitle:       "og title",
				OpenGraphDescription: "og description",
				OpenGraphURL:         "http://dev.qor5.com/product/1",
				OpenGraphType:        "",
				OpenGraphImageURL:    "http://dev.qor5.com/product/1/og.jpg",
			},
			tags: map[string]string{},
			want: `
			<title>title</title>
			<meta name='description' content='description'>
			<meta name='keywords' content='keyword'>
			<meta property='og:title' name='og:title' content='og title'>
			<meta property='og:description' name='og:description' content='og description'>
			<meta property='og:type' name='og:type' content='website'>
			<meta property='og:image' name='og:image' content='http://dev.qor5.com/product/1/og.jpg'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>`,
		},

		{
			name: "Render the SEO html using the tag data",
			setting: Setting{
				Title:                "title",
				Description:          "description",
				Keywords:             "keyword",
				OpenGraphTitle:       "og title",
				OpenGraphDescription: "og description",
				OpenGraphURL:         "http://dev.qor5.com/product/1",
				OpenGraphType:        "",
				OpenGraphImageURL:    "http://dev.qor5.com/product/1/og.jpg",
			},
			tags: map[string]string{
				"og:type":       "product",
				"twitter:image": "http://dev.qor5.com/product/1/twitter.jpg",
			},
			want: `
			<title>title</title>
			<meta name='description' content='description'>
			<meta name='keywords' content='keyword'>
			<meta property='og:title' name='og:title' content='og title'>
			<meta property='og:description' name='og:description' content='og description'>
			<meta property='og:type' name='og:type' content='product'>
			<meta property='og:image' name='og:image' content='http://dev.qor5.com/product/1/og.jpg'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
			<meta property='twitter:image' name='twitter:image' content='http://dev.qor5.com/product/1/twitter.jpg'>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.setting.HTMLComponent(tt.tags).MarshalHTML(context.TODO()); !metaEqual(string(got), tt.want) {
				t.Errorf("Setting.HTMLComponent() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
