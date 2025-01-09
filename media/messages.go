package media

import "fmt"

type Messages struct {
	Crop                        string
	CropImage                   string
	ChooseFile                  string
	Delete                      string
	ChooseAFile                 string
	Search                      string
	UploadFiles                 string
	Cropping                    string
	DescriptionUpdated          string
	DescriptionForAccessibility string
	OrderBy                     string
	UploadedAt                  string
	UploadedAtDESC              string
	All                         string
	Images                      string
	Videos                      string
	Files                       string
	SampleArgsText              func(id string) string

	Copy                                  string
	CopyUpdated                           string
	Rename                                string
	RenameUpdated                         string
	Name                                  string
	NewFolder                             string
	UpdateDescription                     string
	UpdateDescriptionTextFieldPlaceholder string
	ChooseFolder                          string
	MoveTo                                string
	MovedFailed                           string
	MovedSuccess                          string
	Folders                               string
	UploadFile                            string
	DeleteObjects                         func(v int) string
	MediaLibrary                          string
	UnSupportFileType                     string
	CopyImageURL                          string
}

var Messages_en_US = &Messages{
	Crop:                        "Crop",
	CropImage:                   "Crop Image",
	ChooseFile:                  "Choose File",
	Delete:                      "Delete",
	ChooseAFile:                 "Choose a File",
	Search:                      "Search",
	UploadFiles:                 "Upload files",
	Cropping:                    "Cropping",
	DescriptionUpdated:          "Description Updated",
	DescriptionForAccessibility: "description for accessibility",
	OrderBy:                     "Order By",
	UploadedAt:                  "Date Uploaded",
	UploadedAtDESC:              "Date Uploaded (DESC)",
	All:                         "All",
	Images:                      "Images",
	Videos:                      "Videos",
	Files:                       "Files",

	Copy:                                  "Copy",
	CopyUpdated:                           "Copy Updated",
	Rename:                                "Rename",
	RenameUpdated:                         "Rename Updated",
	Name:                                  "Name",
	NewFolder:                             "New Folder",
	UpdateDescription:                     "Update Description",
	UpdateDescriptionTextFieldPlaceholder: "Description",
	ChooseFolder:                          "Choose Folder",
	MoveTo:                                "Move to",
	MovedFailed:                           "Moved Failed",
	MovedSuccess:                          "Moved Success",
	Folders:                               "Folders",
	UploadFile:                            "Upload File",
	DeleteObjects: func(v int) string {
		return fmt.Sprintf(`Are you sure you want to delete %v objects`, v)
	},

	MediaLibrary:      "Media Library",
	UnSupportFileType: "UnSupport FileType",
	CopyImageURL:      "copy image URL",
}

var Messages_zh_CN = &Messages{
	Crop:                        "剪裁",
	CropImage:                   "剪裁图片",
	ChooseFile:                  "选择文件",
	Delete:                      "删除",
	ChooseAFile:                 "选择一个文件",
	Search:                      "搜索",
	UploadFiles:                 "上传多个文件",
	Cropping:                    "正在剪裁...",
	DescriptionUpdated:          "描述更新成功",
	DescriptionForAccessibility: "图片描述",
	OrderBy:                     "排序",
	UploadedAt:                  "上传时间",
	UploadedAtDESC:              "上传时间 (降序)",
	All:                         "全部",
	Images:                      "图片",
	Videos:                      "视频",
	Files:                       "文件",

	Copy:                                  "拷贝",
	CopyUpdated:                           "拷贝成功",
	Rename:                                "重命名",
	RenameUpdated:                         "重命名成功",
	Name:                                  "名称",
	NewFolder:                             "新文件夹",
	UpdateDescription:                     "更新描述",
	UpdateDescriptionTextFieldPlaceholder: "描述",
	ChooseFolder:                          "选择文件夹",
	MoveTo:                                "移动到",
	MovedFailed:                           "移动失败",
	MovedSuccess:                          "移动成功",
	Folders:                               "文件夹",
	UploadFile:                            "上传文件",
	DeleteObjects: func(v int) string {
		return fmt.Sprintf(`是否删除 %v 个条目`, v)
	},
	MediaLibrary:      "媒体库",
	UnSupportFileType: "不支持的文件类型",
	CopyImageURL:      "拷贝图片链接",
}

var Messages_ja_JP = &Messages{
	Crop:                        "トリミング",
	CropImage:                   "画像をトリミング",
	ChooseFile:                  "ファイルを選択",
	Delete:                      "削除",
	ChooseAFile:                 "ファイルを選択",
	Search:                      "検索",
	UploadFiles:                 "ファイルをアップロード",
	Cropping:                    "トリミング中",
	DescriptionUpdated:          "説明を更新しました",
	DescriptionForAccessibility: "画像の説明",
	OrderBy:                     "並び替え",
	UploadedAt:                  "アップロード日時",
	UploadedAtDESC:              "アップロード日時 (降順)",
	All:                         "すべて",
	Images:                      "画像",
	Videos:                      "動画",
	Files:                       "ファイル",

	Copy:                                  "コピー",
	CopyUpdated:                           "コピーが更新されました",
	Rename:                                "名前を変更する",
	RenameUpdated:                         "名前の変更が成功しました",
	Name:                                  "名称",
	NewFolder:                             "新規フォルダ",
	UpdateDescription:                     "画像の説明を編集する",
	UpdateDescriptionTextFieldPlaceholder: "画像の説明",
	ChooseFolder:                          "フォルダを選択",
	MoveTo:                                "移動する",
	MovedFailed:                           "移動に失敗しました",
	MovedSuccess:                          "移動に成功しました",
	Folders:                               "フォルダ",
	UploadFile:                            "ファイルをアップロード",
	DeleteObjects: func(v int) string {
		return fmt.Sprintf(`Are you sure you want to delete %v objects`, v)
	},
	MediaLibrary:      "メディアライブラリ",
	UnSupportFileType: "サポートされていないファイル形式",
	CopyImageURL:      "画像のURLをコピー",
}
