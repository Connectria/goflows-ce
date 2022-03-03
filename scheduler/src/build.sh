#!/bin/bash 

# -----------------------------------------------------------------------------
# build_processor.sh - build the GoFlows Processor (goflows-processor)
# -----------------------------------------------------------------------------

unset GO111MODULE
go clean 
go mod tidy 
go build -a -o ../goflows-scheduler

# -----------------------------------------------------------------------------
