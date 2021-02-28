#!/bin/bash

set -xe
set -o pipefail

bindir=~/bin
unitdir=~/.config/systemd/user
configdir=~/.config/mbrt/backsched

# build and install the binary
go build ./cmd/backsched
install -d "${bindir}"
install -T -m 755 backsched "${bindir}/backsched"
rm backsched

# install the config dir
install -d "${configdir}"
install -T -m 600 config-default.jsonnet "${configdir}/config.jsonnet"
install -T -m 600 backsched.libsonnet "${configdir}/backsched.libsonnet"

# install the service file
install -d "${unitdir}"
install -m 644 backsched-check.service "${unitdir}"
install -m 644 backsched-check.timer "${unitdir}"

# enable the timer and reload
systemctl --user daemon-reload
systemctl --user enable backsched-check.timer
