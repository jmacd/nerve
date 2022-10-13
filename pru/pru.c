#include <rsc_types.h>
#include <pru_rpmsg.h>
#include <pru_virtio_ids.h>
#include <am335x/pru_cfg.h> 
#include <am335x/pru_intc.h> 

#define VIRTIO_CONFIG_S_DRIVER_OK 4

#define NOTIFY_ID 1
/*
 * Sizes of the virtqueues (expressed in number of buffers supported,
 * and must be power of 2)
 */
#define PRU_RPMSG_VQ0_SIZE	16
#define PRU_RPMSG_VQ1_SIZE	16

/* The feature bitmap for virtio rpmsg
 */
#define VIRTIO_RPMSG_F_NS	0		//name service notifications

/* This firmware supports name service notifications as one of its features */
#define RPMSG_PRU_C0_FEATURES	(1 << VIRTIO_RPMSG_F_NS)

/* The PRU-ICSS system events used for RPMsg are defined in the Linux device tree
 * PRU0 uses system event 16 (To ARM) and 17 (From ARM)
 * PRU1 uses system event 18 (To ARM) and 19 (From ARM)
 */
#define TO_ARM_HOST   16	
#define FROM_ARM_HOST 17

/* Host-0 Interrupt sets bit 30 in register R31 */
#define HOST_INT			((uint32_t) 1 << 30)	

/* Copied from the internet! */
#define offsetof(st, m) \
    ((uint32_t)&(((st *)0)->m))

/* Mapping sysevts to a channel. Each pair contains a sysevt, channel. */
// Note: wrong PRU
/* struct ch_map pru_intc_map[] = { {18, 3}, */
/* 				 {19, 1}, */
/* }; */
struct ch_map pru_intc_map[] = { 
	{16, 2},
	{17, 0},
};

/* Definition for unused interrupts */
#define HOST_UNUSED		255

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
	    0,
	    HOST_UNUSED,
	    2,
	    HOST_UNUSED,

	    // Note: wrong PRU
            /* HOST_UNUSED, */
            /* 1, */
            /* HOST_UNUSED, */
            /* 3, */

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

/*
 * Using the name 'rpmsg-pru' will probe the rpmsg_pru driver found
 * at linux-x.y.z/drivers/rpmsg/rpmsg_pru.c
 */
#define CHAN_NAME			"rpmsg-pru"
#define CHAN_DESC			"Channel 30"
#define CHAN_PORT			30

char payload[RPMSG_BUF_SIZE];

volatile register uint32_t __R30; /* output register for PRU */
volatile register uint32_t __R31; /* input register for PRU */

void test(void);

void main(void) {
  test();
  
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

void test(void) {
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
      while (pru_rpmsg_receive(&transport, &src, &dst, payload, &len) ==
             PRU_RPMSG_SUCCESS) {
        /* Echo the message back to the same address from which we just received
         */
        pru_rpmsg_send(&transport, dst, src, payload, len);
      }
    }
  }
}
