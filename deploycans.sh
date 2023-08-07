#!/bin/bash

# Define the function to create the ledger argument
createLedgerArg() {
    local ICP_PERUN_MINT_ACC="433bd8e9dd65bdfb34259667578e749136f3e0ea1566e10af1e0dd324cbd9144"
    local ICP_PERUN_USERA_ACC="97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
    local ICP_PERUN_USERB_ACC="40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"

    echo "(record {minting_account = \"$ICP_PERUN_MINT_ACC\"; initial_values = vec { record { \"$ICP_PERUN_USERA_ACC\"; record { e8s=10_000_000} }; record { \"$ICP_PERUN_USERB_ACC\"; record { e8s=10_000_000 } }}; send_whitelist = vec {}})"
}

# Define the function to deploy Perun
deployPerun() {
    local execPath=$1

    path=$(which dfx)
    if [ -z "$path" ]; then
        echo "Error: dfx not found in PATH"
        return 1
    fi

    echo "Deploying Perun"
    cd $execPath
    deployMsg=$($path deploy icp_perun 2>&1)
    status=$?

    if [ $status -ne 0 ]; then
        echo "Error deploying icp_perun:\n$deployMsg\n"
        return $status
    else
        echo "$deployMsg"
        return 0
    fi
}

# Define the function to deploy the ledger
deployLedger() {
    local execPath=$1
    local ledgerArg=$2

    path=$(which dfx)
    if [ -z "$path" ]; then
        echo "Error: dfx not found in PATH"
        return 1
    fi

    echo "Deploying the Ledger with the following parameters: $ledgerArg"
    cd $execPath
    outputLedger=$($path deploy ledger --argument "$ledgerArg" 2>&1)
    status=$?

    if [ $status -ne 0 ]; then
        echo "Error deploying ledger:\n$outputLedger\n"
        return $status
    else
        echo "$outputLedger"
        return 0
    fi
}

# Define the function to deploy Perun
deployPerun() {
    local execPath=$1

    path=$(which dfx)
    if [ -z "$path" ]; then
        echo "Error: dfx not found in PATH"
        return 1
    fi

    echo "Deploying Perun"
    # cd $execPath
    deployMsg=$($path deploy icp_perun 2>&1)
    status=$?

    if [ $status -ne 0 ]; then
        echo "Error deploying icp_perun:\n$deployMsg\n"
        return $status
    else
        echo "$deployMsg"
        return 0
    fi
}

# Call the functions
ledgerArg=$(createLedgerArg)
./startdfx.sh
execPath="./userdata"  # Change this to the directory where dfx command needs to be executed

deployLedger $execPath "$ledgerArg"
deployPerun $execPath



