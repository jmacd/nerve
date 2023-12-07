#!/bin/sh

HOST=linux.local

scp -q -r -p BB-PRU-DMA.dts jmacd@${HOST}:ti-linux-kernel-dev/KERNEL/arch/arm/boot/dts/overlays

#ssh -q jmacd@${HOST} '(cd ti-linux-kernel-dev/KERNEL)'
