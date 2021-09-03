module github.com/qor/qor5

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.62
	github.com/goplaid/web v1.1.5
	github.com/goplaid/x v1.0.10-0.20210903060327-e186ee64edce
	github.com/jinzhu/gorm v1.9.16
	github.com/qor/media v0.0.0-20210601073757-402011f3b027
	github.com/qor/oss v0.0.0-20210412121326-3c5583a62015
	github.com/sunfmin/reflectutils v1.0.0
	github.com/theplant/htmlgo v1.0.1
)

//replace github.com/goplaid/web => ../../goplaid/web
//replace github.com/goplaid/x => ../../goplaid/x
