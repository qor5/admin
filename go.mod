module github.com/qor5/admin/v3

go 1.22

require (
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/aws/aws-sdk-go v1.51.21
	github.com/disintegration/imaging v1.6.2
	github.com/dustin/go-humanize v1.0.1
	github.com/go-chi/chi/v5 v5.0.12
	github.com/gocarina/gocsv v0.0.0-20231116093920-b87c2d0e983a
	github.com/google/go-cmp v0.6.0
	github.com/gosimple/slug v1.14.0
	github.com/gosimple/unidecode v1.0.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/iancoleman/strcase v0.3.0
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.10.9
	github.com/markbates/goth v1.79.0
	github.com/mholt/archiver/v4 v4.0.0-alpha.8
	github.com/ory/ladon v1.2.0
	github.com/pquerna/otp v1.4.0
	github.com/qor/oss v0.0.0-20230717083721-c04686f83630
	github.com/qor5/ui/v3 v3.0.0
	github.com/qor5/web/v3 v3.0.0
	github.com/qor5/x/v3 v3.0.0
	github.com/sunfmin/reflectutils v1.0.3
	github.com/theplant/bimg v1.1.1
	github.com/theplant/gofixtures v1.1.0
	github.com/theplant/htmlgo v1.0.3
	github.com/theplant/sliceutils v0.0.0-20200406042209-89153d988eb1
	github.com/theplant/testingutils v0.0.0-20240326065615-ab2586803ce4
	github.com/thoas/go-funk v0.9.3
	github.com/tnclong/go-que v0.0.0-20240226030728-4e1f3c8ec781
	github.com/ua-parser/uap-go v0.0.0-20240113215029-33f8e6d47f38
	github.com/wcharczuk/go-chart/v2 v2.1.1
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	goji.io/v3 v3.0.0
	golang.org/x/text v0.14.0
	gorm.io/driver/postgres v1.5.7
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.25.9
)

require (
	cloud.google.com/go/compute v1.25.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/blend/go-sdk v1.20220411.3 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.5.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.2.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/markbates/going v1.0.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/nwaples/rardecode/v2 v2.0.0-beta.2 // indirect
	github.com/ory/pagination v0.0.1 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/image v0.15.0 // indirect
	golang.org/x/oauth2 v0.19.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/qor5/web/v3 => ../../qor5/web

replace github.com/qor5/ui/v3 => ../../qor5/ui

replace github.com/qor5/x/v3 => ../../qor5/x
