module github.com/qor/qor5

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.62
	github.com/disintegration/imaging v1.6.2
	github.com/go-chi/chi v1.5.4
	github.com/go-playground/form v3.1.4+incompatible
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/goplaid/web v1.1.14-0.20211012134701-57895edc9687
	github.com/goplaid/x v1.0.16-0.20211010021035-634f3e1c5b86
	github.com/gosimple/slug v1.9.0
	github.com/gosimple/unidecode v1.0.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jinzhu/configor v1.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.10.3
	github.com/markbates/goth v1.68.0
	github.com/qor/admin v0.0.0-20210421035046-739414767209 // indirect
	github.com/qor/media v0.0.0-20210601073757-402011f3b027 // indirect
	github.com/qor/oss v0.0.0-20210412121326-3c5583a62015
	github.com/qor/qor v0.0.0-20210513025647-811b8dd7cfcf
	github.com/qor/serializable_meta v0.0.0-20180510060738-5fd8542db417
	github.com/stretchr/testify v1.7.0
	github.com/sunfmin/reflectutils v1.0.2
	github.com/theplant/bimg v1.1.1
	github.com/theplant/gofixtures v1.1.0
	github.com/theplant/htmlgo v1.0.3
	github.com/theplant/testingutils v0.0.0-20190603093022-26d8b4d95c61
	goji.io v2.0.2+incompatible
	golang.org/x/text v0.3.7
	gorm.io/driver/postgres v1.1.1
	gorm.io/gorm v1.21.15
)

// replace github.com/goplaid/web => ../../goplaid/web
// replace github.com/goplaid/x => ../../goplaid/x
// replace github.com/sunfmin/reflectutils => ../../reflectutils
