module github.com/qor5/admin/v3/utils/testflow/gentool

go 1.22.5

require (
	github.com/gobuffalo/flect v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/qor5/admin/v3 v3.0.1-0.20240424102851-d75759576158
	github.com/qor5/web/v3 v3.0.7
	github.com/sergi/go-diff v1.3.1
	mvdan.cc/gofumpt v0.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/orcaman/concurrent-map/v2 v2.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/qor5/admin/v3 => ../../../

// replace github.com/qor5/web/v3 => ../../../../web
