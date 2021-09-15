module github.com/arcology-network/eshing-svc

go 1.13

require (
	github.com/arcology-network/3rd-party v0.9.2-0.20210626004852-924da2642860
	github.com/arcology-network/common-lib v0.9.2-0.20210907015240-00e4f072f9b8
	github.com/arcology-network/component-lib v0.9.3-0.20210915014210-d8d4e578b73f
	github.com/arcology-network/concurrenturl v0.0.0-20210908071701-a1f430c90b99
	github.com/prometheus/client_golang v1.6.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.15.0
)

// replace (
// 	github.com/arcology-network/common-lib => ../common-lib/
// 	github.com/arcology-network/component-lib => ../component-lib/
// 	github.com/arcology-network/concurrentlib => ../concurrentlib/
// 	github.com/arcology-network/concurrenturl => ../concurrenturl/
// 	github.com/arcology-network/urlarbitrator-engine => ../urlarbitrator-engine/
// 	github.com/arcology-network/vm-adaptor => ../vm-adaptor/
// )
