#!/bin/sh

BONE=192.168.6.2

scp -q -p * debian@${BONE}:pru

ssh -q debian@${BONE} '(cd pru && ./build3.sh)'
