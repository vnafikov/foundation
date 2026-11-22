module github.com/vnafikov/foundation

go 1.27.4

require (
	github.com/coder/websocket v1.8.15
	github.com/jackc/pgx/v5 v5.13.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mitchellh/panicwrap v1.0.0
	github.com/vnafikov/common v0.0.0-20261122222030-b0e4e64beaef
	google.golang.org/grpc v1.85.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.4 // indirect
	github.com/vnafikov/modelmeta-gen v0.0.0-20260521160609-bf935cbcbe25 // indirect
	golang.org/x/exp v0.0.0-20261121134650-e1094eeca1b5 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/telemetry v0.0.0-20261119160235-508b12de1089 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/tools v0.53.0 // indirect
	golang.org/x/vuln v1.10.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261118221215-c8830f628b3f // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

tool (
	github.com/vnafikov/modelmeta-gen
	golang.org/x/vuln/cmd/govulncheck
)
