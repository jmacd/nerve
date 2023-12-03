#!/bin/sh

#BONE=presskit.local
BONE=nervekit.local

scp -q -r -p * debian@${BONE}:nerve

ssh -q debian@${BONE} '(cd nerve && ./build3.sh)'
