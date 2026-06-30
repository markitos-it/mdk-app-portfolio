#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
ENVIRONMENT_FILE="bin/shared/environment.sh"
source $ENVIRONMENT_FILE

function log_info() {
    echo "[INFO] $*"
}
function log_error() {
    echo "[ERROR] $*" >&2
}

#:[.'.]:>-------------------------------------
show_banner
show_config
go run cmd/app/main.go
#:[.'.]:>-------------------------------------