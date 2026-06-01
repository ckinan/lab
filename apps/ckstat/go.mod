module github.com/ckinan/lab/apps/ckstat

go 1.25.0

require (
github.com/ckinan/lab/libs/proc v0.0.0
github.com/ckinan/lab/libs/util v0.0.0
)

replace (
github.com/ckinan/lab/libs/proc => ../../libs/proc
github.com/ckinan/lab/libs/util => ../../libs/util
)
