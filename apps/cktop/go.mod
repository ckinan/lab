module github.com/ckinan/lab/apps/cktop

go 1.25.0

require (
github.com/charmbracelet/bubbles v1.0.0
github.com/charmbracelet/bubbletea v1.3.10
github.com/charmbracelet/lipgloss v1.1.0
github.com/ckinan/lab/libs/proc v0.0.0
github.com/ckinan/lab/libs/util v0.0.0
)

replace (
github.com/ckinan/lab/libs/proc => ../../libs/proc
github.com/ckinan/lab/libs/util => ../../libs/util
)
