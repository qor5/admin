module github.com/qor5/admin/v3

go 1.24.0

require (
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/aws/aws-sdk-go-v2 v1.38.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.68.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.17.0
	github.com/go-chi/chi/v5 v5.2.2
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gosimple/slug v1.14.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/huandu/go-clone v1.7.3
	github.com/iancoleman/strcase v0.3.0
	github.com/jinzhu/inflection v1.0.0
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.10.9
	github.com/manifoldco/promptui v0.9.0
	github.com/markbates/goth v1.80.0
	github.com/mholt/archiver/v4 v4.0.0-alpha.9
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/ory/ladon v1.3.0
	github.com/oschwald/geoip2-golang v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.4.0
	github.com/qor5/confx v0.0.0-20250426065316-0d28db5b4d54
	github.com/qor5/imaging v1.6.4
	github.com/qor5/web v1.3.2
	github.com/qor5/web/v3 v3.0.12-0.20250618085230-3764d0e521a8
	github.com/qor5/x/v3 v3.2.1-0.20251126082016-f61128fc8187
	github.com/samber/lo v1.50.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/spf13/cast v1.7.1
	github.com/stretchr/testify v1.11.1
	github.com/sunfmin/reflectutils v1.0.6
	github.com/sunfmin/snippetgo v0.0.3
	github.com/theplant/bimg v1.1.1
	github.com/theplant/docgo v0.0.16
	github.com/theplant/gofixtures v1.1.3
	github.com/theplant/htmlgo v1.0.3
	github.com/theplant/inject v1.1.0
	github.com/theplant/osenv v0.0.2
	github.com/theplant/relay v0.8.0
	github.com/theplant/sliceutils v0.0.0-20200406042209-89153d988eb1
	github.com/theplant/testenv v0.2.1
	github.com/theplant/testingutils v0.0.2
	github.com/tnclong/go-que v0.0.0-20240226030728-4e1f3c8ec781
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6
	github.com/wcharczuk/go-chart/v2 v2.1.2
	github.com/yosssi/gohtml v0.0.0-20201013000340-ee4748c638f4
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/text v0.32.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
	gorm.io/datatypes v1.2.7
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.5.6
	gorm.io/gorm v1.30.1
)

require (
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	connectrpc.com/connect v1.18.1 // indirect
	connectrpc.com/cors v0.1.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/STARRY-S/zip v0.1.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.6 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.59 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.28 // indirect
	github.com/aws/aws-sdk-go-v2/feature/rds/auth v1.6.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.14 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.5.2 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dlclark/regexp2 v1.11.2 // indirect
	github.com/docker/docker v28.3.3+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-kit/kit v0.12.1-0.20220826005032-a7ba4fa4e289 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/form v3.1.4+incompatible // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.25.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.3.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.21.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/go-clone/generic v1.7.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jjeffery/errors v1.0.3 // indirect
	github.com/jjeffery/kv v0.8.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/markbates/going v1.0.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mdelapenya/tlscert v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nwaples/rardecode/v2 v2.2.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/ory/pagination v0.0.1 // indirect
	github.com/oschwald/maxminddb-golang v1.13.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/redis/go-redis/v9 v9.11.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shirou/gopsutil/v4 v4.25.6 // indirect
	github.com/shurcooL/github_flavored_markdown v0.0.0-20210228213109-c3a9aa474629 // indirect
	github.com/shurcooL/highlight_diff v0.0.0-20230708024848-22f825814995 // indirect
	github.com/shurcooL/highlight_go v0.0.0-20230708025100-33e05792540a // indirect
	github.com/shurcooL/octicon v0.0.0-20230705024016-66bff059edb8 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sorairolake/lzip-go v0.3.5 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/testcontainers/testcontainers-go v0.38.0 // indirect
	github.com/testcontainers/testcontainers-go/modules/redis v0.38.0 // indirect
	github.com/theplant/appkit v0.0.0-20250528023215-3d0d299dc4c6 // indirect
	github.com/theplant/validator v0.0.0-20210202101755-357a9daa8f5f // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/tidwall/gjson v1.17.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/wI2L/jsondiff v0.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/image v0.23.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
)

// replace github.com/qor5/web/v3 => ../web

// replace github.com/qor5/x/v3 => ../x
