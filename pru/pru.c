#include <rsc_types.h>

#define offsetof(st, m) \
    ((uint32_t)&(((st *)0)->m))

struct my_resource_table {
  struct resource_table base;

  uint32_t offset[1]; /* Should match 'num' in actual definition */

  struct fw_rsc_carveout carveout;
};

#pragma DATA_SECTION(resourceTable, ".resource_table")
#pragma RETAIN(resourceTable)
struct my_resource_table resourceTable = {
	1,	/* Resource table version: only version 1 is supported by the current driver */
	1,	/* number of entries in the table */
	0, 0,	/* reserved, must be zero */
	/* offsets to entries */
	{
		offsetof(struct my_resource_table, carveout),
	},
	/* carveout */
	{
	  (uint32_t) TYPE_CARVEOUT, /* type */
	  (uint32_t) 0, /* da */
	  (uint32_t) 0, /* pa */
	  (uint32_t) 1<<23, /* len (8MB) */
	  (uint32_t) 0, /* flags */
	  (uint32_t) 0, /* reserved */
	  "framebufs",
	},	  
};

#define CYCLES_PER_SECOND 200000000 /* PRU has 200 MHz clock */

// https://markayoder.github.io/PRUCookbook/05blocks/blocks.html#blocks_mapping_bits
#define P9_31 (1 << 0)  // blue
#define P9_29 (1 << 1)  // orange
#define P9_30 (1 << 2)  // green
#define P9_28 (1 << 3)  // red

#define R P9_28
#define G P9_30
#define B P9_31
#define O P9_29

volatile register uint32_t __R30; /* output register for PRU */

void main(void) {
    while (1) {
      __R30 |= R|G|B|O;
      __delay_cycles(CYCLES_PER_SECOND / 10);
      __R30 &= ~R;
      __R30 &= ~G;
      __R30 &= ~B;
      __R30 &= ~O;
      __delay_cycles(CYCLES_PER_SECOND / 10);
    }
}
