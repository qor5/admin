package media_library_view

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
	SampleArgsText              func(id string) string
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
}
