module github.com/ckinan/lab/apps/ckagent

go 1.25.0

require (
	github.com/VictoriaMetrics/metrics v1.43.2
	github.com/ckinan/lab/libs/goruntime v0.0.0
	github.com/ckinan/lab/libs/proc v0.0.0
)

replace github.com/ckinan/lab/libs/proc => ../../libs/proc
replace github.com/ckinan/lab/libs/goruntime => ../../libs/goruntime
