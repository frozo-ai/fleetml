module github.com/fleetml/fleetml/agent

go 1.24

require (
	github.com/fleetml/fleetml/proto v0.0.0
	github.com/shirou/gopsutil/v3 v3.24.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.29.1
)

replace github.com/fleetml/fleetml/proto => ../proto
