#include <pru_rpmsg.h>
#include <pru_virtio_ids.h>
#include <rsc_types.h>

#include <am335x/pru_cfg.h>
#include <am335x/pru_intc.h>

#define VIRTIO_CONFIG_S_DRIVER_OK 4

#define NOTIFY_ID 1
/*
 * Sizes of the virtqueues (expressed in number of buffers supported,
 * and must be power of 2)
 */
#define PRU_RPMSG_VQ0_SIZE 16
#define PRU_RPMSG_VQ1_SIZE 16

/* The feature bitmap for virtio rpmsg
 */
#define VIRTIO_RPMSG_F_NS 0 // name service notifications

/* This firmware supports name service notifications as one of its features */
#define RPMSG_PRU_C0_FEATURES (1 << VIRTIO_RPMSG_F_NS)

/* The PRU-ICSS system events used for RPMsg are defined in the Linux device
 * tree PRU0 uses system event 16 (To ARM) and 17 (From ARM) PRU1 uses system
 * event 18 (To ARM) and 19 (From ARM)
 */
// PRU 0
//#define TO_ARM_HOST 16
//#define FROM_ARM_HOST 17
// PRU 1
#define TO_ARM_HOST 18
#define FROM_ARM_HOST 19

/* Host-0 Interrupt sets bit 30 in register R31 */
#define HOST_INT ((uint32_t)1 << 30)

/* Copied from the internet! */
#define offsetof(st, m) ((uint32_t) & (((st *)0)->m))

/* Mapping sysevts to a channel. Each pair contains a sysevt, channel. */
// PRU 1?
struct ch_map pru_intc_map[] = {
    {18, 3},
    {19, 1},
};
// PRU 0?
/* struct ch_map pru_intc_map[] = { */
/*     {16, 2}, */
/*     {17, 0}, */
/* }; */

/* Definition for unused interrupts */
#define HOST_UNUSED 255

struct my_resource_table {
  struct resource_table base;

  uint32_t offset[3]; /* Should match 'num' in actual definition */

  struct fw_rsc_carveout carveout;

  struct fw_rsc_vdev rpmsg_vdev;
  struct fw_rsc_vdev_vring rpmsg_vring0;
  struct fw_rsc_vdev_vring rpmsg_vring1;

  struct fw_rsc_custom pru_ints;
};

#pragma DATA_SECTION(resourceTable, ".resource_table")
#pragma RETAIN(resourceTable)
struct my_resource_table resourceTable = {
    {
        1, /* Resource table version: only version 1 is supported by the current
              driver */
        3, /* number of entries in the table */
        0, 0, /* reserved, must be zero */
    },
    /* offsets to entries */
    {
        offsetof(struct my_resource_table, carveout),
        offsetof(struct my_resource_table, rpmsg_vdev),
        offsetof(struct my_resource_table, pru_ints),
    },
    /* carveout */
    {
        (uint32_t)TYPE_CARVEOUT, /* type */
        (uint32_t)0,             /* da */
        (uint32_t)0,             /* pa */
        (uint32_t)1 << 23,       /* len (8MB) */
        (uint32_t)0,             /* flags */
        (uint32_t)0,             /* reserved */
        "framebufs",
    },

    /* rpmsg vdev entry */
    {
        (uint32_t)TYPE_VDEV,             // type
        (uint32_t)VIRTIO_ID_RPMSG,       // id
        (uint32_t)NOTIFY_ID,             // notifyid
        (uint32_t)RPMSG_PRU_C0_FEATURES, // dfeatures
        (uint32_t)0,                     // gfeatures
        (uint32_t)0,                     // config_len
        (uint8_t)0,                      // status
        (uint8_t)2,                      // num_of_vrings, only two is supported
        {(uint8_t)0, (uint8_t)0},        // reserved
                                         /* no config data */
    },
    /* the two vrings */
    {
        0,                  // da, will be populated by host, can't pass it in
        16,                 // align (bytes),
        PRU_RPMSG_VQ0_SIZE, // num of descriptors
        0,                  // notifyid, will be populated, can't pass right now
        0                   // reserved
    },
    {
        0,                  // da, will be populated by host, can't pass it in
        16,                 // align (bytes),
        PRU_RPMSG_VQ1_SIZE, // num of descriptors
        0,                  // notifyid, will be populated, can't pass right now
        0                   // reserved
    },

    {
        TYPE_CUSTOM,
        TYPE_PRU_INTS,
        sizeof(struct fw_rsc_custom_ints),
        {
            /* PRU_INTS version */
            0x0000,

            /* Channel-to-host mapping, 255 for unused */

            // PRU 0
            /* 0, */
            /* HOST_UNUSED, */
            /* 2, */
            /* HOST_UNUSED, */

            // PRU 1?
            1,
            HOST_UNUSED,
            3,
            HOST_UNUSED,

            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            /* Number of evts being mapped to channels */
            (sizeof(pru_intc_map) / sizeof(struct ch_map)),
            /* Pointer to the structure containing mapped events */
            pru_intc_map,
        },
    },
};

/*
 * Using the name 'rpmsg-pru' will probe the rpmsg_pru driver found
 * at linux-x.y.z/drivers/rpmsg/rpmsg_pru.c
 */
#define CHAN_NAME "rpmsg-pru"
#define CHAN_DESC "Channel 30"
#define CHAN_PORT 30

char payload[RPMSG_BUF_SIZE];

// Set up the pointers to each of the GPIO ports
uint32_t *gpio0 = (uint32_t *)0x44e07000; // GPIO Bank 0  See Table 2.2 of TRM
uint32_t *gpio1 = (uint32_t *)0x4804c000; // GPIO Bank 1
uint32_t *gpio2 = (uint32_t *)0x481ac000; // GPIO Bank 2
uint32_t *gpio3 = (uint32_t *)0x481ae000; // GPIO Bank 3

#define GPIO_CLEARDATAOUT (0x190 / 4) // For clearing the GPIO registers
#define GPIO_SETDATAOUT (0x194 / 4)   // For setting the GPIO registers
#define GPIO_DATAOUT (0x138 / 4)      // For reading the GPIO registers

volatile register uint32_t __R30; /* output register for PRU */
volatile register uint32_t __R31; /* input register for PRU */

// Clock is P8-43 (pruout) R30 bit 2
#define CLOCK_SHIFT 2

// Latch is P8-46 (pruout) R30 bit 1
#define LATCH_SHIFT 1

// OE is P8-45 (pruout) R30 bit 0
#define OE_SHIFT 0

// Delay in cycles
#define DELAY 10

void setRow(uint32_t row);
void clock();
void setPix0();
void setPix1();
void latchRows();

void main(void) {

  struct pru_rpmsg_transport transport;
  uint16_t src, dst, len;
  volatile uint8_t *status;

  /* Allow OCP master port access by the PRU so the PRU can read external
   * memories */
  CT_CFG.SYSCFG_bit.STANDBY_INIT = 0;

  /* Clear the status of the PRU-ICSS system event that the ARM will use to
   * 'kick' us */
  CT_INTC.SICR_bit.STS_CLR_IDX = FROM_ARM_HOST;

  /* Make sure the Linux drivers are ready for RPMsg communication */
  status = &resourceTable.rpmsg_vdev.status;
  while (!(*status & VIRTIO_CONFIG_S_DRIVER_OK))
    ;

  /* Initialize the RPMsg transport structure */
  pru_rpmsg_init(&transport, &resourceTable.rpmsg_vring0,
                 &resourceTable.rpmsg_vring1, TO_ARM_HOST, FROM_ARM_HOST);

  /* Create the RPMsg channel between the PRU and ARM user space using the
   * transport structure. */
  while (pru_rpmsg_channel(RPMSG_NS_CREATE, &transport, CHAN_NAME, CHAN_DESC,
                           CHAN_PORT) != PRU_RPMSG_SUCCESS)
    ;

  while (1) {
    /* Check bit 31 of register R31 to see if the ARM has kicked us */
    if (__R31 & HOST_INT) {
      /* Clear the event status */
      CT_INTC.SICR_bit.STS_CLR_IDX = FROM_ARM_HOST;
      /* Receive all available messages, multiple messages can be sent per kick
       */
      if (pru_rpmsg_receive(&transport, &src, &dst, payload, &len) ==
          PRU_RPMSG_SUCCESS) {
        break;
      }
    }
  }
  // Send the carveout address to the ARM program.
  memcpy(payload, &resourceTable.carveout.pa, 4);
  pru_rpmsg_send(&transport, dst, src, payload, 4);

  // Initialize the carveout (testing)
  uint32_t *start = (uint32_t *)resourceTable.carveout.pa;
  uint32_t *limit = (uint32_t *)(resourceTable.carveout.pa + (1 << 23));

  volatile uint32_t *shared = (uint32_t *)resourceTable.carveout.pa;

  for (; shared < limit; shared++) {
    *shared = (shared - start);
  }

  // Begin display loop
  uint32_t i, row;

  while (1) {
    for (row = 0; row < 16; row++) {
      setRow(row);

      for (i = 0; i < 64; i++) {
        setPix0();
        clock();
        setPix1();
        clock();
      }

      latchRows();
    }
  }
}

void setRow(uint32_t on) {
  // 0xf because 4 lines.  If this were a x64 panel (1/32 scan) use 0x1f.
  uint32_t off = ~on & 0xf;

  // Selector bits start at position 13 in gpio2
  gpio2[GPIO_SETDATAOUT] = on << 13;
  gpio2[GPIO_CLEARDATAOUT] = off << 13;
}

void clock() {
  __R30 |= 1 << CLOCK_SHIFT;
  __delay_cycles(DELAY);
  __R30 &= ~(1 << CLOCK_SHIFT);
  __delay_cycles(DELAY);
}

void latchRows() {
  __R30 |= 1 << OE_SHIFT;
  __delay_cycles(DELAY);
  __R30 |= 1 << LATCH_SHIFT;
  __delay_cycles(DELAY);
  __R30 &= ~(1 << LATCH_SHIFT);
  __delay_cycles(DELAY);
  __R30 &= ~(1 << OE_SHIFT);
  __delay_cycles(DELAY);
}

void setPix0() {
  gpio2[GPIO_SETDATAOUT] = (1 << 9); // red
}

void setPix1() {
  gpio2[GPIO_SETDATAOUT] = (1 << 23); // blue
}
