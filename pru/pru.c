#include <pru_rpmsg.h>
#include <pru_virtio_ids.h>
#include <rsc_types.h>

#include <am335x/pru_cfg.h>
#include <am335x/pru_ctrl.h>
#include <am335x/pru_intc.h>

volatile register uint32_t __R30; /* output register for PRU */
volatile register uint32_t __R31; /* input/interrupt register for PRU */

#define WORDSZ sizeof(uint32_t)

#include "control.h"
#include "edma.h"

struct pru_rpmsg_transport rpmsg_transport;
char rpmsg_payload[RPMSG_BUF_SIZE];
uint16_t rpmsg_src, rpmsg_dst, rpmsg_len;

// Set in resourceTable.rpmsg_vdev.status when the kernel is ready.
#define VIRTIO_CONFIG_S_DRIVER_OK ((uint32_t)1 << 2)

// Sizes of the virtqueues (expressed in number of buffers supported,
// and must be power of 2)
#define PRU_RPMSG_VQ0_SIZE 16
#define PRU_RPMSG_VQ1_SIZE 16

// The feature bitmap for virtio rpmsg
#define VIRTIO_RPMSG_F_NS 0 // name service notifications

// This firmware supports name service notifications as one of its features.
#define RPMSG_PRU_C0_FEATURES (1 << VIRTIO_RPMSG_F_NS)

// sysevt 16 == pr1_pru_mst_intr[0]_intr_req
#define SYSEVT_PRU_TO_ARM 16

// sysevt 17 == pr1_pru_mst_intr[1]_intr_req
#define SYSEVT_ARM_TO_PRU 17

// sysevt 18 == pr1_pru_mst_intr[2]_intr_req
#define SYSEVT_PRU_TO_EDMA 18

// sysevt 63 == tpcc_int_pend_po1
#define SYSEVT_EDMA_TO_PRU 63

// Chanel 2 is the first (of 8) PRU interrupt output channels.
#define HOST_INTERRUPT_CHANNEL_PRU_TO_ARM 2

// Channel 0 is the first (of 2) PRU interrupt input channels.
#define HOST_INTERRUPT_CHANNEL_ARM_TO_PRU 0

// Chanel 9 is the last (of 8) PRU interrupt output channels,
// a.k.a. pr1_host[7] maps to DMA channel 0 (see TRM 11.3.20).
#define HOST_INTERRUPT_CHANNEL_PRU_TO_EDMA 9

// Channel 1 is the second (of 2) PRU interrupt input channels.
#define HOST_INTERRUPT_CHANNEL_EDMA_TO_PRU 1

// Interrupt inputs set bits 30 and 31 in register R31.
#define PRU_R31_INTERRUPT_FROM_ARM ((uint32_t)1 << 30)  // Fixed, equals channel 0
#define PRU_R31_INTERRUPT_FROM_EDMA ((uint32_t)1 << 31) // Fixed, equals channel 1

// (From the internet!)
#define offsetof(st, m) ((uint32_t) & (((st *)0)->m))

// Mapping sysevts to a channel. Each pair contains a sysevt, channel.
struct ch_map pru_intc_map[] = {
    {SYSEVT_PRU_TO_ARM, HOST_INTERRUPT_CHANNEL_PRU_TO_ARM},
    {SYSEVT_ARM_TO_PRU, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},

    {SYSEVT_EDMA_TO_PRU, HOST_INTERRUPT_CHANNEL_EDMA_TO_PRU},
    {SYSEVT_PRU_TO_EDMA, HOST_INTERRUPT_CHANNEL_PRU_TO_EDMA},
};

/* Definition for unused interrupts */
#define HOST_UNUSED 255

struct my_resource_table {
  struct resource_table base;

  uint32_t offset[4]; /* Should match 'num' in actual definition */

  struct fw_rsc_carveout framebufs;
  struct fw_rsc_carveout controls;

  struct fw_rsc_vdev rpmsg_vdev;
  struct fw_rsc_vdev_vring rpmsg_vring0;
  struct fw_rsc_vdev_vring rpmsg_vring1;

  struct fw_rsc_custom pru_ints;
};

#pragma DATA_SECTION(resourceTable, ".resource_table")
#pragma RETAIN(resourceTable)
struct my_resource_table resourceTable = {
    {
        1,    /* Resource table version: only version 1 is supported by the current
                 driver */
        4,    /* number of entries in the table */
        0, 0, /* reserved, must be zero */
    },
    /* offsets to entries */
    {
        offsetof(struct my_resource_table, framebufs),
        offsetof(struct my_resource_table, controls),
        offsetof(struct my_resource_table, rpmsg_vdev),
        offsetof(struct my_resource_table, pru_ints),
    },
    /* carveout 1 */
    {
        (uint32_t)TYPE_CARVEOUT,       /* type */
        (uint32_t)0,                   /* da */
        (uint32_t)0,                   /* pa */
        (uint32_t)FRAMEBUF_TOTAL_SIZE, /* len */
        (uint32_t)0,                   /* flags */
        (uint32_t)0,                   /* reserved */
        "framebufs",
    },
    /* carveout 2 */
    {
        (uint32_t)TYPE_CARVEOUT,       /* type */
        (uint32_t)0,                   /* da */
        (uint32_t)0,                   /* pa */
        (uint32_t)CONTROLS_TOTAL_SIZE, /* len */
        (uint32_t)0,                   /* flags */
        (uint32_t)0,                   /* reserved */
        "controls",
    },

    /* rpmsg vdev entry */
    {
        (uint32_t)TYPE_VDEV,             // type
        (uint32_t)VIRTIO_ID_RPMSG,       // id
        (uint32_t)0,                     // notifyid
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
            PRU_INTS_VER0,

            // See TRM 4.4.2.1.  Two interrupt input channels.
            HOST_INTERRUPT_CHANNEL_ARM_TO_PRU,
            HOST_INTERRUPT_CHANNEL_EDMA_TO_PRU,

            // Two used output interrupt channels.
            HOST_INTERRUPT_CHANNEL_PRU_TO_ARM,

            // Six unused interrupt output channels.
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,
            HOST_UNUSED,

            HOST_INTERRUPT_CHANNEL_PRU_TO_EDMA,

            // Number of evts being mapped to channels.
            (sizeof(pru_intc_map) / sizeof(struct ch_map)),

            // The structure containing mapped events.
            pru_intc_map,
        },
    },
};

// Set up the pointers to each of the GPIO ports
uint32_t *gpio0 = (uint32_t *)0x44e07000; // GPIO Bank 0  See Table 2.2 of TRM
uint32_t *gpio1 = (uint32_t *)0x4804c000; // GPIO Bank 1
uint32_t *gpio2 = (uint32_t *)0x481ac000; // GPIO Bank 2
uint32_t *gpio3 = (uint32_t *)0x481ae000; // GPIO Bank 3

#define GPIO_CLEARDATAOUT (0x190 / WORDSZ) // For clearing the GPIO registers
#define GPIO_SETDATAOUT (0x194 / WORDSZ)   // For setting the GPIO registers
#define GPIO_DATAOUT (0x13C / WORDSZ)      // For setting the GPIO registers

// DMA completion interrupt use tpcc_int_pend_po1

// EDMA system event 0 and 1 correspond with pr1_host[7] and pr1_host[6]
// and pr1_host[0:7] maps to channels 2-9 on the PRU.
// => EDMA event 0 == PRU channel 9
// => EDMA event 1 == PRU channel 8
// For both of these, use the low register set (not the high register set).
const int dmaChannel = 0;
const uint32_t dmaChannelMask = (1 << 0);

void setup_dma_channel_zero() {
  // Map Channel 0 to PaRAM 0
  // DCHMAP_0 == DMA Channel 0 mapping to PaRAM set number 0.
  EDMA_BASE[EDMA_DCHMAP_0] = dmaChannel;

  // Setup EDMA region access for Shadow Region 1
  // DRAE1 == DMA Region Access Enable shadow region 1.
  EDMA_BASE[EDMA_DRAE1] |= dmaChannelMask;

  // Setup channel to submit to EDMA TC0. Note DMAQNUM0 is for DMAQNUM0
  // configures the channel controller for channels 0-7, the 0 in
  // 0xfffffff0 corresponds with "E0" of DMAQNUM0 (TRM 11.4.1.6), i.e., DMA
  // channel 0 maps to queue 0.
  EDMA_BASE[EDMA_DMAQNUM_0] &= 0xFFFFFFF0;

  // Clear interrupt and secondary event registers.
  EDMA_BASE[EDMA_SECR] |= dmaChannelMask;
  EDMA_BASE[EDMA_ICR] |= dmaChannelMask;

  // Enable channel interrupt.
  EDMA_BASE[EDMA_IESR] |= dmaChannelMask;

  // Enable channel for an event trigger.
  EDMA_BASE[EDMA_EESR] |= dmaChannelMask;

  // Clear event missed register.
  EDMA_BASE[EDMA_EMCR] |= dmaChannelMask;
}

void start_dma() {
  uint16_t paramOffset;
  edmaParam params;
  volatile edmaParam *ptr;

  // Setup and store PaRAM set for transfer.
  paramOffset = EDMA_PARAM_OFFSET;
  paramOffset += ((dmaChannel * EDMA_PARAM_SIZE) / WORDSZ);

  params.lnkrld.link = 0xFFFF;
  params.lnkrld.bcntrld = 0x0000;
  params.opt.tcc = dmaChannel;
  params.opt.tcinten = 1;
  params.opt.itcchen = 1;
  params.ccnt.ccnt = 1;
  params.abcnt.acnt = 100;
  params.abcnt.bcnt = 1;
  params.bidx.srcbidx = 1;
  params.bidx.dstbidx = 1;
  params.src = 0x4A310000;
  params.dst = 0x4A310100;

  ptr = (volatile edmaParam *)(EDMA_BASE + paramOffset);
  *ptr = params;

  // Trigger transfer.  (4.4.1.2.2 Event Interface Mapping)
  // This is pr1_pru_mst_intr[2]_intr_req, system event 18
  __R31 = R31_INTERRUPT_ENABLE | (SYSEVT_PRU_TO_EDMA - R31_INTERRUPT_OFFSET);
}

void wait_dma() {
  // Wait for completion interrupt.
  while (!(EDMA_BASE[EDMA_IPR] & dmaChannelMask)) {
  }
}

// Delay in cycles
#define DELAY 0 // 1000

void sleep();
void setRow(uint32_t row);
void testPix();
void toggleClock();
void setPix(pixel_t *pixel);
void latchRows();

#define HI 1
#define LO 0

void set(uint32_t *gpio, int bit, int on) {
  if (on) {
    gpio[GPIO_SETDATAOUT] = 1 << bit;
  } else {
    gpio[GPIO_CLEARDATAOUT] = 1 << bit;
  }
}

void uled1(int val) { set(gpio1, 21, val); }
void uled2(int val) { set(gpio1, 22, val); }
void uled3(int val) { set(gpio1, 23, val); }
void uled4(int val) { set(gpio1, 24, val); }

void clock(int val) { set(gpio1, 19, val); }
void latch(int val) { set(gpio1, 29, val); }
void outputEnable(int val) { set(gpio1, 28, val); }

void selA(int val) { set(gpio1, 12, val); }
void selB(int val) { set(gpio1, 13, val); }
void selC(int val) { set(gpio1, 14, val); }
void selD(int val) { set(gpio1, 15, val); }

void sleep() {
  // __delay_cycles(DELAY);
}

void setRow(uint32_t on) {
  // 0xf because 4 address lines.  If this were a x64 panel (1/32
  // scan) use 0x1f for 5 address lines.
  uint32_t off = on ^ 0xf;

  // Selector bits start at position 12 in gpio1
  gpio1[GPIO_SETDATAOUT] = on << 12;
  gpio1[GPIO_CLEARDATAOUT] = off << 12;
}

void toggleClock() {
  sleep();
  clock(HI);
  sleep();
  clock(LO);
}

void latchRows() {
  outputEnable(HI);
  sleep();
  latch(HI);
  sleep();
  latch(LO);
  sleep();
  outputEnable(LO);
  sleep();
}

// Using fpp/capes/bbb/panels/Octoscroller.json as a reference.

// J1
// gp2 |= (1U << 2);  // J1 r1 (P8-gp2)
// 07 |= (1U << 3);  // J1 g1 (P8-08)
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

void setPix(pixel_t *pixel) {
  gpio0[GPIO_DATAOUT] = pixel->gpv0;
  gpio1[GPIO_DATAOUT] = pixel->gpv1;
  gpio2[GPIO_DATAOUT] = pixel->gpv2;
  gpio3[GPIO_DATAOUT] = pixel->gpv3;
}

void reset_hardware_state() {
  // Allow OCP master port access by the PRU.
  CT_CFG.SYSCFG_bit.STANDBY_INIT = 0;

  // Enable PRU0 cycle counter.
  PRU0_CTRL.CTRL_bit.CTR_EN = 1;

  // Clear the system event mapped to the two input interrupts.
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_ARM_TO_PRU;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_TO_PRU;

  // Enable the EDMA (Transfer controller, Channel controller) clocks.
  CM_PER_BASE[CM_PER_TPTC0_CLKCTRL] = CM_PER_CLK_ENABLED;
  CM_PER_BASE[CM_PER_TPCC_CLKCTRL] = CM_PER_CLK_ENABLED;

  // Reset gpio output.
  const uint32_t allbits = 0xffffffff;
  gpio0[GPIO_CLEARDATAOUT] = allbits;
  gpio1[GPIO_CLEARDATAOUT] = allbits;
  gpio2[GPIO_CLEARDATAOUT] = allbits;
  gpio3[GPIO_CLEARDATAOUT] = allbits;
}

void wait_for_virtio_ready() {
  // Make sure the Linux drivers are ready for RPMsg communication
  volatile uint8_t *status = &resourceTable.rpmsg_vdev.status;
  while (!(*status & VIRTIO_CONFIG_S_DRIVER_OK)) {
    // Wait
  }
}

void setup_transport() {
  // Using the name 'rpmsg-pru' will probe the rpmsg_pru driver found
  // at linux/drivers/rpmsg/rpmsg_pru.c
  char *const channel_name = "rpmsg-pru";
  char *const channel_desc = "Channel 30";
  const int channel_port = 30;

  // Initialize two vrings using system events on dedicated channels.
  pru_rpmsg_init(&rpmsg_transport, &resourceTable.rpmsg_vring0, &resourceTable.rpmsg_vring1, SYSEVT_PRU_TO_ARM,
                 SYSEVT_ARM_TO_PRU);

  // Create the RPMsg channel between the PRU and the ARM.
  while (pru_rpmsg_channel(RPMSG_NS_CREATE, &rpmsg_transport, channel_name, channel_desc, channel_port) !=
         PRU_RPMSG_SUCCESS) {
  }
}

void wait_for_arm() {
  while (1) {
    // Check register R31 for the ARM interrupt.
    if (__R31 & PRU_R31_INTERRUPT_FROM_ARM) {

      // Clear the event status *\/
      CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_ARM_TO_PRU;

      // Receive all available messages.
      if (pru_rpmsg_receive(&rpmsg_transport, &rpmsg_src, &rpmsg_dst, rpmsg_payload, &rpmsg_len) == PRU_RPMSG_SUCCESS) {
        break;
      }
    }
  }
}

// Send the carveout addresses to the ARM.
void send_to_arm() {
  memcpy(rpmsg_payload, &resourceTable.controls.pa, 4);
  while (pru_rpmsg_send(&rpmsg_transport, rpmsg_dst, rpmsg_src, rpmsg_payload, 4) != PRU_RPMSG_SUCCESS) {
  }
}

control_t *setup_controls() {
  control_t *ctrl = (control_t *)resourceTable.controls.pa;
  ctrl->framebufs = (uint32_t *)resourceTable.framebufs.pa;
  return ctrl;
}

void main(void) {
  reset_hardware_state();

  wait_for_virtio_ready();

  control_t *ctrl = setup_controls();

  setup_transport();

  setup_dma_channel_zero();

  wait_for_arm();

  send_to_arm();

  // start_dma();
  // wait_dma();

  // Begin display loop
  while (1) {
    pixel_t *pixptr = (pixel_t *)resourceTable.framebufs.pa;

    uint32_t frame;
    for (frame = 0; frame < FRAMEBUF_NUM_FRAMES; frame++) {

      uint32_t row;
      for (row = 0; row < 16; row++) {
        setRow(row);

        uint32_t pix;
        for (pix = 0; pix < 64; pix++) {
          setPix(pixptr++);
          toggleClock();
        }

        latchRows();
      }
      ctrl->framecount++;
    }
  }
}
