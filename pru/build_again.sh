#!/bin/sh

make clean
rm -rf output
mkdir output

echo stop > /sys/class/remoteproc/remoteproc1/state

make output/pru.o PROC=pru TARGET=pru CHIP=AM335x
make output/pru.out PROC=pru TARGET=pru CHIP=AM335x

cp output/pru.out /lib/firmware/blink4-fw

echo blink4-fw > /sys/class/remoteproc/remoteproc1/firmware
