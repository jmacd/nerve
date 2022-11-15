#ifndef __CONTROL_H
#define __CONTROL_H

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
  uint32_t *framebufs;

  uint32_t framecount;
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
