#!/usr/bin/env sh

set -euo pipefail

: ${FLEETSTATE_SERVER=http://127.0.0.1:10080}

usage() {
    echo "Usage:"
    echo "$0 VIN"
    echo
    echo "Example:"
    echo "  FLEETSTATE_SERVER=http://fleetstate $0 THE1VIN" 
    echo
    echo "The default value of FLEETSTATE_SERVER env var is $FLEETSTATE_SERVER"
}

VIN=${1-}

case "$VIN" in
    -*) 
        usage
        exit 1
        ;;
    "")
        echo >&2 "$0: must specify VIN to watch"
        usage
        exit 1
        ;;
esac

curl -s --show-error "${FLEETSTATE_SERVER}/vehicle/${VIN}/stream"
