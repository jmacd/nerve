#!/bin/sh

LINUX=192.168.0.40

scp -q * jmacd@${LINUX}:pru

ssh -q jmacd@${LINUX} '(cd pru && ./build2.sh)'
