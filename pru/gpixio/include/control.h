// MIT License
//
// Copyright (C) Joshua MacDonald
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

#ifndef __CONTROL_H
#define __CONTROL_H

#define PRU_L4_FAST_SHARED_PRUSS_MEM 0x4a310000

#define WORDSZ sizeof(uint32_t)

#define FRAMEBUF_GPIOS 4
#define FRAMEBUF_SCANS 16
#define FRAMEBUF_WIDTH 64

// 16B
#define FRAMEBUF_PIXEL_SIZE (FRAMEBUF_GPIOS * WORDSZ)

// 1KB
#define FRAMEBUF_SCAN_SIZE (FRAMEBUF_WIDTH * FRAMEBUF_PIXEL_SIZE)

// 16KB
#define FRAMEBUF_FRAME_SIZE (FRAMEBUF_SCANS * FRAMEBUF_SCAN_SIZE)

// 4KB local
#define FRAMEBUF_PART_SIZE (1U << 12)

// 4 parts per frame
#define FRAMEBUF_PARTS_PER_FRAME (FRAMEBUF_FRAME_SIZE / FRAMEBUF_PART_SIZE)

// 4 scans per part
#define FRAMEBUF_SCANS_PER_PART (FRAMEBUF_SCANS / FRAMEBUF_PARTS_PER_FRAME)

// carveout 8MB
#define FRAMEBUF_TOTAL_SIZE (1U << 23)

// two 4MB banks
#define FRAMEBUF_BANK_SIZE (FRAMEBUF_TOTAL_SIZE / 2)

// 256 frames per bank
#define FRAMEBUF_FRAMES_PER_BANK (FRAMEBUF_BANK_SIZE / FRAMEBUF_FRAME_SIZE)

typedef struct control control_t;

struct control {
  volatile uint32_t framebufs_addr;
  volatile uint32_t framebufs_size;
  volatile uint32_t framecount;
  volatile uint32_t dma_wait;

  volatile uint32_t ready_bank;
  volatile uint32_t start_bank;
};

// Using fpp/capes/bbb/panels/Octoscroller.json as a reference.
// Using J1 and J3 for testing.

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned _bit0 : 1;  // 0
    unsigned _bit1 : 1;  // 1
    unsigned j4_r1 : 1;  // 2
    unsigned j3_g2 : 1;  // 3
    unsigned j4_b2 : 1;  // 4
    unsigned j3_b2 : 1;  // 5
    unsigned _bit6 : 1;  // 6
    unsigned _bit7 : 1;  // 7
    unsigned j5_g2 : 1;  // 8
    unsigned j5_r2 : 1;  // 9
    unsigned j5_b1 : 1;  // 10
    unsigned j5_g1 : 1;  // 11
    unsigned _bit12 : 1; // 12
    unsigned _bit13 : 1; // 13
    unsigned j8_g2 : 1;  // 14
    unsigned j4_g1 : 1;  // 15
    unsigned _bit16 : 1; // 16
    unsigned _bit17 : 1; // 17
    unsigned _bit18 : 1; // 18
    unsigned _bit19 : 1; // 19
    unsigned _bit20 : 1; // 20
    unsigned _bit21 : 1; // 21
    unsigned j2_b1 : 1;  // 22
    unsigned j1_r2 : 1;  // 23
    unsigned _bit24 : 1; // 24
    unsigned _bit25 : 1; // 25
    unsigned j1_b2 : 1;  // 26
    unsigned j2_r1 : 1;  // 27
    unsigned _bit28 : 1; // 28
    unsigned _bit29 : 1; // 29
    unsigned j3_r1 : 1;  // 30
    unsigned j3_b1 : 1;  // 31
  } bits;
} gpio0_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned _bit0 : 1;        // 0
    unsigned _bit1 : 1;        // 1
    unsigned _bit2 : 1;        // 2
    unsigned _bit3 : 1;        // 3
    unsigned _bit4 : 1;        // 4
    unsigned _bit5 : 1;        // 5
    unsigned _bit6 : 1;        // 6
    unsigned _bit7 : 1;        // 7
    unsigned _bit8 : 1;        // 8
    unsigned _bit9 : 1;        // 9
    unsigned _bit10 : 1;       // 10
    unsigned _bit11 : 1;       // 11
    unsigned rowSelect : 4;    // 15:12
    unsigned j3_r2 : 1;        // 16
    unsigned j4_b1 : 1;        // 17
    unsigned j3_g1 : 1;        // 18
    unsigned inputClock : 1;   // 19
    unsigned _bit20 : 1;       // 20
    unsigned _bit21 : 1;       // 21
    unsigned _bit22 : 1;       // 22
    unsigned _bit23 : 1;       // 23
    unsigned _bit24 : 1;       // 24
    unsigned _bit25 : 1;       // 25
    unsigned _bit26 : 1;       // 26
    unsigned _bit27 : 1;       // 27
    unsigned outputEnable : 1; // 28
    unsigned inputLatch : 1;   // 29
    unsigned _bit30 : 1;       // 30
    unsigned _bit31 : 1;       // 31
  } bits;
} gpio1_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned _bit0 : 1;  // 0
    unsigned j2_g1 : 1;  // 1
    unsigned j1_r1 : 1;  // 2
    unsigned j1_g1 : 1;  // 3
    unsigned j1_g2 : 1;  // 4
    unsigned j1_b1 : 1;  // 5
    unsigned j7_r2 : 1;  // 6
    unsigned j7_b2 : 1;  // 7
    unsigned j7_b1 : 1;  // 8
    unsigned j7_g1 : 1;  // 9
    unsigned j6_g2 : 1;  // 10
    unsigned j7_r1 : 1;  // 11
    unsigned j6_b2 : 1;  // 12
    unsigned j6_r2 : 1;  // 13
    unsigned j6_b1 : 1;  // 14
    unsigned j6_g1 : 1;  // 15
    unsigned j6_r1 : 1;  // 16
    unsigned j5_b2 : 1;  // 17
    unsigned _bit18 : 1; // 18
    unsigned _bit19 : 1; // 19
    unsigned _bit20 : 1; // 20
    unsigned _bit21 : 1; // 21
    unsigned j2_r2 : 1;  // 22
    unsigned j2_g2 : 1;  // 23
    unsigned j2_b2 : 1;  // 24
    unsigned j5_r1 : 1;  // 25
    unsigned _bit26 : 1; // 26
    unsigned _bit27 : 1; // 27
    unsigned _bit28 : 1; // 28
    unsigned _bit29 : 1; // 29
    unsigned _bit30 : 1; // 30
    unsigned _bit31 : 1; // 31
  } bits;
} gpio2_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned _bit0 : 1;  // 0
    unsigned _bit1 : 1;  // 1
    unsigned _bit2 : 1;  // 2
    unsigned _bit3 : 1;  // 3
    unsigned _bit4 : 1;  // 4
    unsigned _bit5 : 1;  // 5
    unsigned _bit6 : 1;  // 6
    unsigned _bit7 : 1;  // 7
    unsigned _bit8 : 1;  // 8
    unsigned _bit9 : 1;  // 9
    unsigned _bit10 : 1; // 10
    unsigned _bit11 : 1; // 11
    unsigned _bit12 : 1; // 12
    unsigned _bit13 : 1; // 13
    unsigned j8_r2 : 1;  // 14
    unsigned j8_b1 : 1;  // 15
    unsigned j8_g1 : 1;  // 16
    unsigned j8_r1 : 1;  // 17
    unsigned j7_g2 : 1;  // 18
    unsigned j4_g2 : 1;  // 19
    unsigned j8_b2 : 1;  // 20
    unsigned j4_r2 : 1;  // 21
    unsigned _bit22 : 1; // 22
    unsigned _bit23 : 1; // 23
    unsigned _bit24 : 1; // 24
    unsigned _bit25 : 1; // 25
    unsigned _bit26 : 1; // 26
    unsigned _bit27 : 1; // 27
    unsigned _bit28 : 1; // 28
    unsigned _bit29 : 1; // 29
    unsigned _bit30 : 1; // 30
    unsigned _bit31 : 1; // 31
  } bits;
} gpio3_t;

#define CONTROLS_TOTAL_SIZE sizeof(control_t)

struct dbl_pixel {
  gpio0_t gpv0;
  gpio1_t gpv1;
  gpio2_t gpv2;
  gpio3_t gpv3;
};

typedef struct dbl_pixel dbl_pixel_t;

#endif
