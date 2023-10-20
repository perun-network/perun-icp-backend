#!/bin/sh

# Send a SIGTERM signal to all processes that have 'dfx' in their command line
pkill -f dfx
dfx_status=$?

# Send a SIGTERM signal to all processes that have 'icx' in their command line
pkill -f icx
icx_status=$?

if [ $dfx_status -ne 0 ] || [ $icx_status -ne 0 ]; then
    echo "Error stopping DFX or ICX..."
    exit 1
else
    echo "DFX and ICX processes stopped."
fi