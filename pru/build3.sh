#!/bin/sh

trap cleanup 1 2 3 6

PID=""

cleanup()
{
  kill -9 $PID
  exit 1
}

ps ax | grep user.out | awk '{print $1}' | xargs kill -9

echo "Stopping ..."
echo stop > /sys/class/remoteproc/remoteproc1/state

#make clean
rm -rf output
mkdir output

make output/pru.o PROC=pru TARGET=pru CHIP=AM335x

make output/pru.out PROC=pru TARGET=pru CHIP=AM335x

cp output/pru.out /lib/firmware/blink4-fw

echo blink4-fw > /sys/class/remoteproc/remoteproc1/firmware

sleep 1
echo "Starting ..."
echo start > /sys/class/remoteproc/remoteproc1/state

sleep 1
echo "Running user ..."

gcc user.c -o user.out
#./user.out&
#PID=$$
#echo "Running... $PID"
sleep 3600
