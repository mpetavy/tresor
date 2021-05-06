module github.com/mpetavy/tresor

go 1.16

require (
	github.com/chai2010/tiff v0.0.0-20200331144629-b5b74f075872
	github.com/disintegration/imaging v1.6.2
	github.com/fatih/structs v1.1.0
	github.com/fogleman/gg v1.3.0
	github.com/go-pg/pg v8.0.6+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/lib/pq v1.7.0
	github.com/mpetavy/common v1.2.6
	github.com/mpetavy/go-dicom v0.0.0-20200607105844-561ed6d653d4
	github.com/onsi/ginkgo v1.13.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/unidoc/unipdf/v3 v3.7.1
	go.mongodb.org/mongo-driver v1.3.4
	golang.org/x/image v0.0.0-20200609002522-3f4726a040e8
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mellium.im/sasl v0.2.1 // indirect
)

//replace github.com/mpetavy/common => ../common
