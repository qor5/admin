module github.com/qor/qor5

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.62
	github.com/disintegration/imaging v1.6.2
	github.com/go-chi/chi v1.5.4
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/goplaid/multipartestutils v0.0.3
	github.com/goplaid/web v1.1.25
	github.com/goplaid/x v1.0.23-0.20220128031632-b8f2afe15d7e
	github.com/gorilla/sessions v1.2.0 // indirect
	github.com/gosimple/slug v1.9.0
	github.com/gosimple/unidecode v1.0.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jinzhu/configor v1.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.10.3
	github.com/markbates/goth v1.68.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/qor/oss v0.0.0-20210412121326-3c5583a62015
	github.com/spf13/cast v1.4.1
	github.com/stretchr/testify v1.7.0
	github.com/sunfmin/reflectutils v1.0.3
	github.com/theplant/bimg v1.1.1
	github.com/theplant/docgo v0.0.7
	github.com/theplant/gofixtures v1.1.0
	github.com/theplant/htmlgo v1.0.3
	github.com/theplant/sliceutils v0.0.0-20200406042209-89153d988eb1
	github.com/theplant/testingutils v0.0.0-20190603093022-26d8b4d95c61
	github.com/tnclong/go-que v0.0.0-20201111043106-1fc5fa2b9761
	goji.io v2.0.2+incompatible
	golang.org/x/text v0.3.7
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/postgres v1.2.1
	gorm.io/gorm v1.22.2
)

// replace github.com/goplaid/web => ../goplaid/web
// replace github.com/goplaid/x => ../../goplaid/x
