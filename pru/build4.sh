#!/bin/sh

#BONE=beaglebone.local
#BONE=presskit.local
#BONE=192.168.6.2
BONE=nervekit.local

scp -q -r -p * debian@${BONE}:nerve

ssh -q debian@${BONE} '(cd nerve && ./build3.sh)'
