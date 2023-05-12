#include "fast.h"

#include <arm_neon.h>

#include <stddef.h>

void CopyPixels(uint8_t *input, uint32_t *output) {
  // uint8x8x4_t vld4_u8 (const uint8_t *)
  // Form of expected instruction(s): vld4.8 {d0, d1, d2, d3}, [r0]

  uint8x16x4_t x = vld4q_u8(input);
  x.vals[0][0] = y;
  x.vals[1][1] = 2;
  x.vals[2][2] = 3;
  x.vals[3][15] = 4;
}
