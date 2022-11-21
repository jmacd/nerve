#ifndef __CONTROL_H
#define __CONTROL_H

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
  volatile uint32_t *framebufs;
  volatile uint32_t framecount;
  volatile uint32_t latch_wait;
  volatile uint32_t dma_wait;
};

struct gpio1_bits {};

typedef gpio1_t union {
  volatile uint32_t GPIO1;

  volatile struct {
    unsigned bit0 : 1;  // 0
    unsigned bit1 : 1;  // 1
    unsigned bit2 : 1;  // 2
    unsigned bit3 : 1;  // 3
    unsigned bit4 : 1;  // 4
    unsigned bit5 : 1;  // 5
    unsigned bit6 : 1;  // 6
    unsigned bit7 : 1;  // 7
    unsigned bit8 : 1;  // 8
    unsigned bit9 : 1;  // 9
    unsigned bit10 : 1; // 10
    unsigned bit11 : 1; // 11

    unsigned rowSelect : 4; // 15:12

    unsigned j3_r2 : 1; // 16
    unsigned bit17 : 1; // 17
    unsigned j3_g1 : 1; // 18

    unsigned inputClock : 1; // 19

    unsigned bit20 : 1; // 20

    unsigned uled0 : 1; // 21
    unsigned uled1 : 1; // 22
    unsigned uled2 : 1; // 23
    unsigned uled3 : 1; // 24

    unsigned bit25 : 1; // 25
    unsigned bit26 : 1; // 26
    unsigned bit27 : 1; // 27

    unsigned outputEnable : 1; // 28
    unsigned inputLatch : 1;   // 29

    unsigned bit30 : 1; // 30
    unsigned bit31 : 1; // 31
  } GPIO1_bit;
};

#define CONTROLS_TOTAL_SIZE sizeof(control_t)

struct pixel {
  uint32_t gpv0;
  uint32_t gpv1;
  uint32_t gpv2;
  uint32_t gpv3;
};

typedef struct pixel pixel_t;

#endif
