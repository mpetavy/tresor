module github.com/mpetavy/tresor

go 1.16

require (
	github.com/chai2010/tiff v0.0.0-20200705094435-2b8a7f42fe29
	github.com/disintegration/imaging v1.6.2
	github.com/fatih/structs v1.1.0
	github.com/fogleman/gg v1.3.0
	github.com/go-pg/pg v8.0.7+incompatible
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/lib/pq v1.10.2
	github.com/mpetavy/common v1.4.6
	github.com/mpetavy/go-dicom v0.0.0-20210302105037-44b79120da96
	github.com/onsi/ginkgo v1.13.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/unidoc/unipdf/v3 v3.26.1
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.7.0
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mellium.im/sasl v0.2.1 // indirect
)

//replace github.com/mpetavy/common => ../common
