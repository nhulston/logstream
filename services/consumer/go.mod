module github.com/nhulston/logstream/services/consumer

go 1.24

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.12.0
	github.com/lib/pq v1.10.9
	github.com/nhulston/logstream/proto/gen v0.0.0
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.1 // indirect
)

replace github.com/nhulston/logstream/proto/gen => ../../proto/gen
