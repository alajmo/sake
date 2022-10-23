#!/bin/bash

set -euo pipefail

function parse_options() {
  SAVE=
  while [[ $# -gt 0 ]]; do
    case "${1}" in
      --save|-s)
        SAVE=YES
        shift
        ;;
      *)
        printf "Unknown flag: ${1}\n\n"
        help
        exit 1
        ;;
    esac
  done
}

function __main__() {
  parse_options $@
    # --style=basic
    if [[ "$SAVE" ]]; then
        hyperfine -N --runs 10 '../dist/sake run ping -s server-9' > ./profiles/ping-no-key
        hyperfine -N --runs 10 '../dist/sake run ping -t reachable' > ./profiles/ping
        hyperfine -N --runs 10 '../dist/sake run ping -p -t reachable' > ./profiles/ping-parallel
        hyperfine -N --runs 10 '../dist/sake run d -t reachable' > ./profiles/nested
        hyperfine -N --runs 10 '../dist/sake run d -p -t reachable' > ./profiles/nested-parallel
        hyperfine -N --runs 10 '../dist/sake list servers' > ./profiles/list-servers
    else
        hyperfine -N --runs 10 '../dist/sake run ping -s server-9'
        hyperfine -N --runs 10 '../dist/sake run ping -t reachable'
        hyperfine -N --runs 10 '../dist/sake run ping -p -t reachable'
        hyperfine -N --runs 10 '../dist/sake run d -t reachable'
        hyperfine -N --runs 10 '../dist/sake run d -p -t reachable'
        hyperfine -N --runs 10 '../dist/sake list servers'
    fi
}

__main__ $@
