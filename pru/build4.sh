#!/bin/sh

BONE=beaglebone.local

scp -q -r -p * debian@${BONE}:pru

ssh -q debian@${BONE} '(cd pru && ./build3.sh)'
