/*
 * Copyright (C) 2015 Texas Instruments Incorporated - http://www.ti.com/
 *
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 *	* Redistributions of source code must retain the above copyright
 *	  notice, this list of conditions and the following disclaimer.
 *
 *	* Redistributions in binary form must reproduce the above copyright
 *	  notice, this list of conditions and the following disclaimer in the
 *	  documentation and/or other materials provided with the
 *	  distribution.
 *
 *	* Neither the name of Texas Instruments Incorporated nor the names of
 *	  its contributors may be used to endorse or promote products derived
 *	  from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/poll.h>
#include <unistd.h>

#include "control.h"

#define MAX_BUFFER_SIZE 512
char readBuf[MAX_BUFFER_SIZE];

#define DEVICE_NAME "/dev/rpmsg_pru30"

#define XSTR(x) STR(x)
#define STR(x) #x

#define PRINTCFG(x) printf(#x ": %u\n", x)

void start_dma(uint32_t localTargetBank, uint32_t currentBank, uint32_t currentFrame, uint32_t currentPart) {
  if (currentBank == 1 && currentFrame == FRAMEBUF_FRAMES_PER_BANK - 1 && currentPart == FRAMEBUF_PARTS_PER_FRAME - 1) {
    currentBank = 0;
    currentFrame = 0;
    currentPart = 0;
  } else {
    currentPart++;
  }
  uint32_t dst;
  uint32_t src;
  dst = PRU_L4_FAST_SHARED_PRUSS_MEM + (localTargetBank * FRAMEBUF_PART_SIZE);
  src = (currentBank * FRAMEBUF_BANK_SIZE) + (currentFrame * FRAMEBUF_FRAME_SIZE) + (currentPart * FRAMEBUF_PART_SIZE);

  printf("copy dst 0x%x src 0x%x size 0x%x\n", dst, src, FRAMEBUF_PART_SIZE);
}

void demoPrint(void) {
  // For two banks
  uint32_t localno = 0;
  uint32_t bankno;

  start_dma(0, 1, FRAMEBUF_FRAMES_PER_BANK - 1, FRAMEBUF_PARTS_PER_FRAME - 1);

  uint32_t pixaddr = 0;

  for (bankno = 0; bankno < 2; bankno++) {
    // For 256 frames per bank
    uint32_t frame;

    for (frame = 0; frame < FRAMEBUF_FRAMES_PER_BANK; frame++) {

      uint32_t part;
      uint32_t row = 0;

      // For 4 parts per frame
      for (part = 0; part < FRAMEBUF_PARTS_PER_FRAME; part++) {

        localno ^= 1;

        // Start a DMA to fill the next local bank.
        printf("start dma local=%u bank=%u frame=%u part=%u\n", localno, bankno, frame, part);
        start_dma(localno, bankno, frame, part);

        // For 4 scans per part
        uint32_t scan;
        for (scan = 0; scan < FRAMEBUF_SCANS_PER_PART; scan++, row++) {

          printf("start row %u addr 0x%x\n", row, pixaddr);

          uint32_t pix;
          for (pix = 0; pix < 64; pix++) {
            pixaddr += 16;
          }
        }
      }
    }
  }

  printf("finish addr 0x%x\n", pixaddr);
}

int main(void) {
#if 0
  demoPrint();
#endif
#if 1
  struct pollfd pollfds[1];
  int i;
  int result = 0;

  printf("Configured: %s\n", DEVICE_NAME);
  PRINTCFG(FRAMEBUF_PIXEL_SIZE);
  PRINTCFG(FRAMEBUF_SCAN_SIZE);
  PRINTCFG(FRAMEBUF_FRAME_SIZE);
  PRINTCFG(FRAMEBUF_PART_SIZE);
  PRINTCFG(FRAMEBUF_PARTS_PER_FRAME);
  PRINTCFG(FRAMEBUF_SCANS_PER_PART);
  PRINTCFG(FRAMEBUF_TOTAL_SIZE);
  PRINTCFG(FRAMEBUF_BANK_SIZE);
  PRINTCFG(FRAMEBUF_FRAMES_PER_BANK);

  /* Open the rpmsg_pru character device file */
  pollfds[0].fd = open(DEVICE_NAME, O_RDWR);

  /*
   * If the RPMsg channel doesn't exist yet the character device
   * won't either.
   * Make sure the PRU firmware is loaded and that the rpmsg_pru
   * module is inserted.
   */
  if (pollfds[0].fd < 0) {
    printf("Failed to open %s\n", DEVICE_NAME);
    return -1;
  }

  /* The RPMsg channel exists and the character device is opened */
  printf("Opened %s\n", DEVICE_NAME);

  /* Send 'hello world!' to the PRU through the RPMsg channel */
  result = write(pollfds[0].fd, "hello world!", 13);
  if (result == 0) {
    printf("Could not send to PRU\n");
    return -1;
  }

  result = read(pollfds[0].fd, readBuf, MAX_BUFFER_SIZE);
  if (result == 0) {
    printf("Could not read from PRU\n");
    return -1;
  }
  uint32_t addr;
  memcpy(&addr, readBuf, 4);
  printf("Message %d received from PRU (%d bytes) %x\n", i, result, addr);

  /* Received all the messages the example is complete */
  printf("Closing %s\n", DEVICE_NAME);

  int fd = open("/dev/mem", O_RDWR, 0);

  uint32_t ctrlPtr = (uint32_t)mmap(NULL, CONTROLS_TOTAL_SIZE, PROT_READ | PROT_WRITE, MAP_SHARED, fd, addr);

  printf("Control mapped at addr=%x\n", ctrlPtr);
  control_t *ctrl = (control_t *)ctrlPtr;

  uint32_t framebufsPtr =
      (uint32_t)mmap(NULL, FRAMEBUF_TOTAL_SIZE, PROT_READ | PROT_WRITE, MAP_SHARED, fd, (uint32_t)ctrl->framebufs);

  printf("Framebufs mapped at addr=%x\n", framebufsPtr);
  uint32_t *framebufs = (uint32_t *)framebufsPtr;

  while (1) {
    memset((void *)framebufs, 0, FRAMEBUF_TOTAL_SIZE);
    uint32_t frame;
    pixel_t *pixptr = (pixel_t *)framebufs;

    for (frame = 0; frame < 2 * FRAMEBUF_FRAMES_PER_BANK; frame++) {

      uint32_t row;

      for (row = 0; row < FRAMEBUF_SCANS; row++) {
        uint32_t pix;

        // For 64 pixels width
        for (pix = 0; pix < 64; pix++) {
          int r = rand();

          pixptr->gpv1.bits.rowSelect = row;
          pixptr->gpv1.bits.inputClock = 0;
          pixptr->gpv1.bits.outputEnable = 0;
          pixptr->gpv1.bits.inputLatch = 0;
          pixptr->gpv0.bits.j3_r1 = (r & 0x1) != 0;
          pixptr->gpv1.bits.j3_g1 = (r & 0x2) != 0;
          pixptr->gpv0.bits.j3_b1 = (r & 0x4) != 0;
          pixptr->gpv1.bits.j3_r2 = (r & 0x8) != 0;
          pixptr->gpv0.bits.j3_g2 = (r & 0x10) != 0;
          pixptr->gpv0.bits.j3_b2 = (r & 0x20) != 0;
          pixptr->gpv2.bits.j1_r1 = (r & 0x40) != 0;
          pixptr->gpv2.bits.j1_g1 = (r & 0x80) != 0;
          pixptr->gpv2.bits.j1_b1 = (r & 0x100) != 0;
          pixptr->gpv0.bits.j1_r2 = (r & 0x200) != 0;
          pixptr->gpv2.bits.j1_g2 = (r & 0x400) != 0;
          pixptr->gpv0.bits.j1_b2 = (r & 0x800) != 0;
          pixptr++;
        }
      }
    }
  }

  // *framebufs = rand();
  // for (int i = 0; i < (FRAMEBUF_TOTAL_SIZE / WORDSZ); i++) {
  //   framebufs[i] = rand();
  // }

  // uint32_t *start = framebufs;
  // uint32_t *limit = framebufs + (FRAMEBUF_TOTAL_SIZE / WORDSZ);
  // while (start < limit) {
  //   *start++ = rand();
  // }

  uint32_t last_value;
  while (1) {
    uint32_t current = ctrl->framecount;

    if (ctrl->dma_wait != 0) {
      printf("dma_wait: %u\n", ctrl->dma_wait);
    }

    if (last_value != 0) {
      uint32_t diff = current - last_value;
      printf("frames/sec: %.1f\n", (double)diff);
    }
    last_value = current;
    sleep(1);
  }

  /* Close the rpmsg_pru character device file */
  close(pollfds[0].fd);

  return 0;
#endif
}
