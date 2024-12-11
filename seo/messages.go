package seo

type Messages struct {
	Variable                      string
	VariableDescription           string
	Basic                         string
	Title                         string
	Description                   string
	Keywords                      string
	OpenGraphInformation          string
	OpenGraphTitle                string
	OpenGraphDescription          string
	OpenGraphURL                  string
	OpenGraphType                 string
	OpenGraphImageURL             string
	OpenGraphImage                string
	OpenGraphMetadata             string
	Seo                           string
	Customize                     string
	SEOPreview                    string
	BlankOpenGraphInformationTips string
}

var Messages_en_US = &Messages{
	Variable:                      "Variables Setting",
	Basic:                         "Basic",
	Title:                         "Title",
	Description:                   "Description",
	Keywords:                      "Keywords",
	OpenGraphInformation:          "Open Graph Information",
	OpenGraphTitle:                "Open Graph Title",
	OpenGraphDescription:          "Open Graph Description",
	OpenGraphURL:                  "Open Graph URL",
	OpenGraphType:                 "Open Graph Type",
	OpenGraphImageURL:             "Open Graph Image URL",
	OpenGraphImage:                "Open Graph Image",
	OpenGraphMetadata:             "Open Graph Metadata",
	Seo:                           "SEO",
	Customize:                     "Customize",
	SEOPreview:                    "SEO Preview",
	BlankOpenGraphInformationTips: `The "Open Graph Information" are blank. The default values will be used on the page.`,
}

var Messages_zh_CN = &Messages{
	Variable:                      "变量设置",
	Basic:                         "基本信息",
	Title:                         "标题",
	Description:                   "描述",
	Keywords:                      "关键词",
	OpenGraphInformation:          "OG 信息",
	OpenGraphTitle:                "OG 标题",
	OpenGraphDescription:          "OG 描述",
	OpenGraphURL:                  "OG 链接",
	OpenGraphType:                 "OG 类型",
	OpenGraphImageURL:             "OG 图片链接",
	OpenGraphImage:                "OG 图片",
	OpenGraphMetadata:             "OG 元数据",
	Seo:                           "搜索引擎优化",
	Customize:                     "自定义",
	SEOPreview:                    "SEO 预览",
	BlankOpenGraphInformationTips: `"OG 信息"w为空。页面将使用默认值。`,
}
