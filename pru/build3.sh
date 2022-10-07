#!/bin/sh

#export PRU_CGT=/usr/lib/ti/pru-software-support-package

#export PRU_CGT=/usr/share/ti/cgt-pru
#export TARGET=pru
#make

#clpru -I /usr/lib/ti/pru-software-support-package/include -I /usr/share/ti/cgt-pru/include pru.c

#export PRU_CGT=/usr/share/ti/cgt-pru
make output/pru.o PROC=pru

