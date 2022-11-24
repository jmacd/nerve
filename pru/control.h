#ifndef __CONTROL_H
#define __CONTROL_H

#define PRU_L4_FAST_SHARED_PRUSS_MEM 0x4a310000

// Using fpp/capes/bbb/panels/Octoscroller.json as a reference.

// J1
// gp2 |= (1U << 2);  // J1 r1 (P8-07)
// gp2 |= (1U << 3);  // J1 g1 (P8-08)
// gp2 |= (1U << 5);  // J1 b1 (P8-09)
// gp0 |= (1U << 23); // J1 r2 (P8-13)
// gp2 |= (1U << 4);  // J1 g2 (P8-10)
// gp0 |= (1U << 26); // J1 b2 (P8-14)

// J3
// gp0 |= 1U << 30; // r1 (P9-11)
// gp1 |= 1U << 18; // g1 (P9-14)
// gp0 |= 1U << 31; // b1 (P9-13)
// gp1 |= 1U << 16; // r2 (P9-15)
// gp0 |= 1U << 3;  // g2 (P9-21)
// gp0 |= 1U << 5;  // b2 (P9-17)

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
  volatile uint32_t dma_wait;
};

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned __bit0 : 1;  // 0
    unsigned __bit1 : 1;  // 1
    unsigned __bit2 : 1;  // 2
    unsigned j3_g2 : 1;   // 3
    unsigned __bit4 : 1;  // 4
    unsigned j3_b2 : 1;   // 5
    unsigned __bit6 : 1;  // 6
    unsigned __bit7 : 1;  // 7
    unsigned __bit8 : 1;  // 8
    unsigned __bit9 : 1;  // 9
    unsigned __bit10 : 1; // 10
    unsigned __bit11 : 1; // 11
    unsigned __bit12 : 1; // 12
    unsigned __bit13 : 1; // 13
    unsigned __bit14 : 1; // 14
    unsigned __bit15 : 1; // 15
    unsigned __bit16 : 1; // 16
    unsigned __bit17 : 1; // 17
    unsigned __bit18 : 1; // 18
    unsigned __bit19 : 1; // 19
    unsigned __bit20 : 1; // 20
    unsigned __bit21 : 1; // 21
    unsigned __bit22 : 1; // 22
    unsigned j1_r2 : 1;   // 23
    unsigned __bit24 : 1; // 24
    unsigned __bit25 : 1; // 25
    unsigned j1_b2 : 1;   // 26
    unsigned __bit27 : 1; // 27
    unsigned __bit28 : 1; // 28
    unsigned __bit29 : 1; // 29
    unsigned j3_r1 : 1;   // 30
    unsigned j3_b1 : 1;   // 31
  } bits;
} gpio0_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned __bit0 : 1;  // 0
    unsigned __bit1 : 1;  // 1
    unsigned __bit2 : 1;  // 2
    unsigned __bit3 : 1;  // 3
    unsigned __bit4 : 1;  // 4
    unsigned __bit5 : 1;  // 5
    unsigned __bit6 : 1;  // 6
    unsigned __bit7 : 1;  // 7
    unsigned __bit8 : 1;  // 8
    unsigned __bit9 : 1;  // 9
    unsigned __bit10 : 1; // 10
    unsigned __bit11 : 1; // 11

    unsigned rowSelect : 4; // 15:12

    unsigned j3_r2 : 1;   // 16
    unsigned __bit17 : 1; // 17
    unsigned j3_g1 : 1;   // 18

    unsigned inputClock : 1; // 19

    unsigned __bit20 : 1; // 20

    unsigned uled0 : 1; // 21
    unsigned uled1 : 1; // 22
    unsigned uled2 : 1; // 23
    unsigned uled3 : 1; // 24

    unsigned __bit25 : 1; // 25
    unsigned __bit26 : 1; // 26
    unsigned __bit27 : 1; // 27

    unsigned outputEnable : 1; // 28
    unsigned inputLatch : 1;   // 29

    unsigned __bit30 : 1; // 30
    unsigned __bit31 : 1; // 31
  } bits;
} gpio1_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned __bit0 : 1;  // 0
    unsigned __bit1 : 1;  // 1
    unsigned j1_r1 : 1;   // 2
    unsigned j1_g1 : 1;   // 3
    unsigned j1_g2 : 1;   // 4
    unsigned j1_b1 : 1;   // 5
    unsigned __bit6 : 1;  // 6
    unsigned __bit7 : 1;  // 7
    unsigned __bit8 : 1;  // 8
    unsigned __bit9 : 1;  // 9
    unsigned __bit10 : 1; // 10
    unsigned __bit11 : 1; // 11
    unsigned __bit12 : 1; // 12
    unsigned __bit13 : 1; // 13
    unsigned __bit14 : 1; // 14
    unsigned __bit15 : 1; // 15
    unsigned __bit16 : 1; // 16
    unsigned __bit17 : 1; // 17
    unsigned __bit18 : 1; // 18
    unsigned __bit19 : 1; // 19
    unsigned __bit20 : 1; // 20
    unsigned __bit21 : 1; // 21
    unsigned __bit22 : 1; // 22
    unsigned __bit23 : 1; // 23
    unsigned __bit24 : 1; // 24
    unsigned __bit25 : 1; // 25
    unsigned __bit26 : 1; // 26
    unsigned __bit27 : 1; // 27
    unsigned __bit28 : 1; // 28
    unsigned __bit29 : 1; // 29
    unsigned __bit30 : 1; // 30
    unsigned __bit31 : 1; // 31
  } bits;
} gpio2_t;

typedef union {
  volatile uint32_t word;

  volatile struct {
    unsigned __bit0 : 1;  // 0
    unsigned __bit1 : 1;  // 1
    unsigned __bit2 : 1;  // 2
    unsigned __bit3 : 1;  // 3
    unsigned __bit4 : 1;  // 4
    unsigned __bit5 : 1;  // 5
    unsigned __bit6 : 1;  // 6
    unsigned __bit7 : 1;  // 7
    unsigned __bit8 : 1;  // 8
    unsigned __bit9 : 1;  // 9
    unsigned __bit10 : 1; // 10
    unsigned __bit11 : 1; // 11
    unsigned __bit12 : 1; // 12
    unsigned __bit13 : 1; // 13
    unsigned __bit14 : 1; // 14
    unsigned __bit15 : 1; // 15
    unsigned __bit16 : 1; // 16
    unsigned __bit17 : 1; // 17
    unsigned __bit18 : 1; // 18
    unsigned __bit19 : 1; // 19
    unsigned __bit20 : 1; // 20
    unsigned __bit21 : 1; // 21
    unsigned __bit22 : 1; // 22
    unsigned __bit23 : 1; // 23
    unsigned __bit24 : 1; // 24
    unsigned __bit25 : 1; // 25
    unsigned __bit26 : 1; // 26
    unsigned __bit27 : 1; // 27
    unsigned __bit28 : 1; // 28
    unsigned __bit29 : 1; // 29
    unsigned __bit30 : 1; // 30
    unsigned __bit31 : 1; // 31
  } bits;
} gpio3_t;

#define CONTROLS_TOTAL_SIZE sizeof(control_t)

struct pixel {
  gpio0_t gpv0;
  gpio1_t gpv1;
  gpio2_t gpv2;
  gpio3_t gpv3;
};

typedef struct pixel pixel_t;

#endif
