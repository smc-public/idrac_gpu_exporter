module github.com/mrlhansen/idrac_exporter

go 1.21

require (
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/common v0.62.0
	github.com/xhit/go-str2duration/v2 v2.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)

replace github.com/mrlhansen/idrac_exporter => ../idrac_exporter
