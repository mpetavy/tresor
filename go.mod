module github.com/mpetavy/tresor

go 1.16

require (
	github.com/chai2010/tiff v0.0.0-20200705094435-2b8a7f42fe29
	github.com/disintegration/imaging v1.6.2
	github.com/dsoprea/go-exif/v3 v3.0.0-20210625224831-a6301f85c82b
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/fatih/structs v1.1.0
	github.com/fogleman/gg v1.3.0
	github.com/go-pg/pg v8.0.7+incompatible
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/labstack/echo/v4 v4.9.0 // indirect
	github.com/lib/pq v1.10.2
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mpetavy/common v1.4.36
	github.com/mpetavy/go-dicom v0.0.0-20210302105037-44b79120da96
	github.com/onsi/ginkgo v1.13.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/unidoc/unipdf/v3 v3.26.1
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.7.0
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/net v0.0.0-20220930213112-107f3e3c3b0b // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mellium.im/sasl v0.2.1 // indirect
)

//replace github.com/mpetavy/common => ../common
