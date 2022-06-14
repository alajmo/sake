#!/bin/bash

set -eum

SAKE_PATH=$(dirname $(dirname $(realpath "$0")))
SAKE_EXAMPLE_PATH="./test/playground"
OUTPUT_FILE="$SAKE_PATH/res/output.json"
OUTPUT_GIF="$SAKE_PATH/res/output.gif"

_init() {
  cd "$SAKE_EXAMPLE_PATH"

  # remove previous artifacts
  rm "$OUTPUT_FILE" "$OUTPUT_GIF" -f
}

_simulate_commands() {
  # list of the commands we want to record
  local CMD='
    _mock() {
      # the | pv -qL 30 part is used to simulate typing
      echo "\$ $1" | pv -qL 40

      first_char=$(printf %.1s "$1")
      if test "$first_char" != "'#'"; then
        $1
      fi
    }

    clear
    export PS1="\$ "
    sleep 1s

    _mock "# Ping all hosts one by one"
    sleep 0.5s
    _mock "sake run ping --tags remote"
    sleep 1.3s
    clear

    _mock "# Ping all hosts in parallel"
    sleep 0.5s
    _mock "sake run ping --tags remote --parallel"
    sleep 1.3s
    clear

    sleep 0.5s
    _mock "# Query servers for some stats"
    sleep 1s
    _mock "sake run info --all --output table"
    sleep 2s
    clear
  '

  asciinema rec -c "$CMD" --idle-time-limit 100 --title mani --quiet "$OUTPUT_FILE" &
  fg %1
}

_generate_gif() {
  cd "$SAKE_PATH/res"

  # Convert to gif
  output_file=$(basename $OUTPUT_FILE)
  output_gif=$(basename $OUTPUT_GIF)
  docker run --rm -v "$PWD":/data asciinema/asciicast2gif -S 3 -h 30 "$output_file" "$output_gif"
}

_main() {
  _init
  _simulate_commands
  _generate_gif
}

_main

