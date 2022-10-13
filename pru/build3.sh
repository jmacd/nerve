#!/bin/sh

make output/pru.o PROC=pru TARGET=pru CHIP=AM335x

make output/pru.out PROC=pru TARGET=pru CHIP=AM335x

cp output/pru.out /lib/firmware/blink4-fw

echo "Stopping ..."
echo stop > /sys/class/remoteproc/remoteproc1/state

sleep 1
echo blink4-fw > /sys/class/remoteproc/remoteproc1/firmware

sleep 1
echo "Starting ..."
echo start > /sys/class/remoteproc/remoteproc1/state

sleep 1
echo "Running user ..."

gcc user.c -o user.out
./user.out
