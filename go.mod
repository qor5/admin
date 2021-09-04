module github.com/qor/qor5

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.62
	github.com/goplaid/web v1.1.6-0.20210904141118-704ee18371ae
	github.com/goplaid/x v1.0.10-0.20210904141651-86ea8556ab49
	github.com/jinzhu/configor v1.2.1 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/qor/admin v0.0.0-20210421035046-739414767209 // indirect
	github.com/qor/media v0.0.0-20210601073757-402011f3b027
	github.com/qor/oss v0.0.0-20210412121326-3c5583a62015
	github.com/qor/qor v0.0.0-20210513025647-811b8dd7cfcf // indirect
	github.com/sunfmin/reflectutils v1.0.0
	github.com/theplant/htmlgo v1.0.1
	golang.org/x/text v0.3.6
)

//replace github.com/goplaid/web => ../../goplaid/web
//replace github.com/goplaid/x => ../../goplaid/x
