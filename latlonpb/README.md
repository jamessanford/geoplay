When updating `.proto` files, make sure to `go generate` before submitting.

This has some dependencies:

-	install `protoc` from the google protobuf package
-	`go get -u github.com/golang/protobuf/protoc-gen-go`
-	`go get -u github.com/gogo/letmegrpc/protoc-gen-letmegrpc`
