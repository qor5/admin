module github.com/qor/qor5

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.62
	github.com/disintegration/imaging v1.6.2
	github.com/goplaid/web v1.1.7
	github.com/goplaid/x v1.0.11-0.20210908120943-5d255c785149
	github.com/gosimple/slug v1.9.0
	github.com/gosimple/unidecode v1.0.0
	github.com/jinzhu/configor v1.2.1 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.0
	github.com/qor/admin v0.0.0-20210421035046-739414767209 // indirect
	github.com/qor/media v0.0.0-20210601073757-402011f3b027 // indirect
	github.com/qor/oss v0.0.0-20210412121326-3c5583a62015
	github.com/qor/qor v0.0.0-20210513025647-811b8dd7cfcf
	github.com/qor/serializable_meta v0.0.0-20180510060738-5fd8542db417
	github.com/sunfmin/reflectutils v1.0.0
	github.com/theplant/bimg v1.1.1
	github.com/theplant/htmlgo v1.0.1
	golang.org/x/text v0.3.6
)

//replace github.com/goplaid/web => ../../goplaid/web
//replace github.com/goplaid/x => ../../goplaid/x
