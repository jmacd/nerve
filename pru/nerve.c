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

#include <pru_rpmsg.h>
#include <pru_virtio_ids.h>
#include <rsc_types.h>
#include <string.h>

#include <am335x/pru_cfg.h>
#include <am335x/pru_ctrl.h>
#include <am335x/pru_intc.h>

volatile register uint32_t __R30; // output register for PRU
volatile register uint32_t __R31; // input/interrupt register for PRU

#include "edma.h"
#include "gpixio/include/control.h"

struct pru_rpmsg_transport rpmsg_transport;
char rpmsg_payload[RPMSG_BUF_SIZE];
uint16_t rpmsg_src, rpmsg_dst, rpmsg_len;

// dmaChannel is 0 is mapped to PRU interrupt channel 9 by default.
// @@@
const int dmaChannel = 0;
const uint32_t dmaChannelMask = (1 << 0);

volatile edmaParam *edma_param_entry;

dbl_pixel_t *frame_banks[2];
dbl_pixel_t *local_banks[2];

// Set up the pointers to each of the GPIO ports
uint32_t *const gpio0 = (uint32_t *)0x44e07000; // GPIO Bank 0  See Table 2.2 of TRM
uint32_t *const gpio1 = (uint32_t *)0x4804c000; // GPIO Bank 1
uint32_t *const gpio2 = (uint32_t *)0x481ac000; // GPIO Bank 2
uint32_t *const gpio3 = (uint32_t *)0x481ae000; // GPIO Bank 3

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

// sysevt 62 == tpcc_errint_pend_po1
#define SYSEVT_EDMA_CTRL_ERROR_TO_PRU 62

// sysevt 61 == tptc_errint_pend_po1
#define SYSEVT_EDMA_CHAN_ERROR_TO_PRU 61

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

// Definition for unused interrupts
#define HOST_UNUSED 255

// HI and LO are abbreviations used below.
#define HI 1
#define LO 0

// These are word-size offsets from the GPIO register base address.
#define GPIO_CLEARDATAOUT (0x190 / WORDSZ) // For clearing the GPIO registers
#define GPIO_SETDATAOUT (0x194 / WORDSZ)   // For setting the GPIO registers
#define GPIO_DATAOUT (0x13C / WORDSZ)      // For setting the GPIO registers

// Color bits used for flash(), warn(), park().
#define CBITS_BLACK 0x0
#define CBITS_RED 0x1
#define CBITS_GREEN 0x2
#define CBITS_BLUE 0x4
#define CBITS_YELLOW (CBITS_RED | CBITS_GREEN)
#define CBITS_CYAN (CBITS_GREEN | CBITS_BLUE)
#define CBITS_MAG (CBITS_RED | CBITS_BLUE)
#define CBITS_WHITE (CBITS_RED | CBITS_GREEN | CBITS_BLUE)

#pragma DATA_SECTION(my_irq_rsc, ".pru_irq_map")
#pragma RETAIN(my_irq_rsc)

#if 0
struct pru_irq_rsc my_irq_rsc = {
    // Mapping sysevts to a channel. Each pair contains a sysevt, channel.
    0, /* type = 0 */
    6, /* number of system events being mapped */
    {
        // Interrupts to and from the ARM (virtio).
        {SYSEVT_PRU_TO_ARM, 2, HOST_INTERRUPT_CHANNEL_PRU_TO_ARM},
        {SYSEVT_ARM_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},

        // Interrupts to and from the EDMA.
        {SYSEVT_EDMA_TO_PRU, 1, HOST_INTERRUPT_CHANNEL_EDMA_TO_PRU},
        {SYSEVT_PRU_TO_EDMA, 9, HOST_INTERRUPT_CHANNEL_PRU_TO_EDMA},

        // Error interrupts from EDMA on ARM->PRU channel.
        {SYSEVT_EDMA_CTRL_ERROR_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},
        {SYSEVT_EDMA_CHAN_ERROR_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},
    },
};
#else
// This is a list of interrupts going to the PRU.
struct pru_irq_rsc my_irq_rsc = {
    // Mapping sysevts going to the PRU core. Each pair contains a
    // sysevt, interrupt priority (typically matches host channel),
    // and host channel.
    0, /* type = 0 */
    4, /* number of system events being mapped */
    {
        // Interrupts to and from the ARM (virtio).
        {SYSEVT_ARM_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},

        // Interrupts to and from the EDMA.
        {SYSEVT_EDMA_TO_PRU, 1, HOST_INTERRUPT_CHANNEL_EDMA_TO_PRU},

        // Error interrupts from EDMA on ARM->PRU channel.
        {SYSEVT_EDMA_CTRL_ERROR_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},
        {SYSEVT_EDMA_CHAN_ERROR_TO_PRU, 0, HOST_INTERRUPT_CHANNEL_ARM_TO_PRU},
    },
};
#endif

control_t *global_ctrl;

#define NUM_RESOURCES 3

// my_resource_table describes the custom hardware settings used by
// this program.
struct my_resource_table {
  struct resource_table base;

  uint32_t offset[NUM_RESOURCES]; // Should match 'num' in actual definition

  struct fw_rsc_carveout framebufs;      // Resource 0
  struct fw_rsc_carveout controls;       // Resource 1
  struct fw_rsc_vdev rpmsg_vdev;         // Resource 2
  struct fw_rsc_vdev_vring rpmsg_vring0; // (cont)
  struct fw_rsc_vdev_vring rpmsg_vring1; // (cont)
};

#pragma DATA_SECTION(resourceTable, ".resource_table")
#pragma RETAIN(resourceTable)
// my_resource_table is (as I understand it) how the Linux kernel
// knows what it needs to start the firmware.
struct my_resource_table resourceTable = {
    // resource_table base
    {
        1,             // Resource table version: only version 1 is supported
        NUM_RESOURCES, // Number of entries in the table (equals length of offset field).
        0, 0,          // Reserved zero fields
    },
    // Entry offsets
    {
        offsetof(struct my_resource_table, framebufs),
        offsetof(struct my_resource_table, controls),
        offsetof(struct my_resource_table, rpmsg_vdev),
    },
    // Carveout 1
    {
        (uint32_t)TYPE_CARVEOUT,       // type
        (uint32_t)FW_RSC_ADDR_ANY,     // da
        (uint32_t)0,                   // pa
        (uint32_t)FRAMEBUF_TOTAL_SIZE, // len
        (uint32_t)0,                   // flags
        (uint32_t)0,                   // reserved
        "framebufs",
    },
    // Carveout 2
    {
        (uint32_t)TYPE_CARVEOUT,       // type
        (uint32_t)FW_RSC_ADDR_ANY,     // da
        (uint32_t)0,                   // pa
        (uint32_t)CONTROLS_TOTAL_SIZE, // len
        (uint32_t)0,                   // flags
        (uint32_t)0,                   // reserved
        "controls",
    },
    // RPMsg virtual device
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
    },
    // The two vring structs must be packed after the vdev entry.
    {
        FW_RSC_ADDR_ANY,    // da, will be populated by host, can't pass it in
        16,                 // align (bytes),
        PRU_RPMSG_VQ0_SIZE, // num of descriptors
        0,                  // notifyid, will be populated, can't pass right now
        0                   // reserved
    },
    {
        FW_RSC_ADDR_ANY,    // da, will be populated by host, can't pass it in
        16,                 // align (bytes),
        PRU_RPMSG_VQ1_SIZE, // num of descriptors
        0,                  // notifyid, will be populated, can't pass right now
        0                   // reserved
    },
};

// Set updates modifies a single bit of a GPIO register.
void set(uint32_t *gpio, int bit, int on) {
  if (on) {
    gpio[GPIO_SETDATAOUT] = 1 << bit;
  } else {
    gpio[GPIO_CLEARDATAOUT] = 1 << bit;
  }
}

// uledN toggles the 4 user-programmable LEDs (although the BBB starts
// with them bound to other events, you can echo none >
// /sys/class/leds/$led/trigger to disable triggers and make them
// available for use.
void uled1(int val) { set(gpio1, 21, val); }
void uled2(int val) { set(gpio1, 22, val); }
void uled3(int val) { set(gpio1, 23, val); }
void uled4(int val) { set(gpio1, 24, val); }

// clock, latch, and outputEnable of the HUB75 board (all connections)
// can be modified directly.
void clock(int val) { set(gpio1, 19, val); }
void latch(int val) { set(gpio1, 29, val); }
void outputEnable(int val) { set(gpio1, 28, val); }

// selA-selD allow setting one bit of the address selector.
void selA(int val) { set(gpio1, 12, val); }
void selB(int val) { set(gpio1, 13, val); }
void selC(int val) { set(gpio1, 14, val); }
void selD(int val) { set(gpio1, 15, val); }

void pause() {
  //__delay_cycles(100);
}

// setRow sets the 4-bit address into the row selector GPIO bits.
//
// Note latchRows starts with outputEnable(HI) and this ends with
// outputEnable(LO).  The raising/lowering of enableOutput is paired.
void setRow(uint32_t on) {
  // 0xf because 4 address lines.  If this were a x64 panel (1/32
  // scan) use 0x1f for 5 address lines.
  uint32_t off = on ^ 0xf;

  pause();
  // Selector bits start at position 12 in gpio1
  gpio1[GPIO_SETDATAOUT] = on << 12;
  gpio1[GPIO_CLEARDATAOUT] = off << 12;
}

// toggleClock raises and lowers the HUB75 clock signal.
void toggleClock() {
  pause();
  clock(HI);
  pause();
  clock(LO);
}

// largeRows turns off the output before latching.  setRow will re-enable it.
void latchRows(uint32_t row) {
  pause();
  outputEnable(HI);
  pause();
  latch(HI);
  pause();
  latch(LO);
  pause();
  outputEnable(LO);
}

// setPix writes 4 GPIO words.  they are expected to have the correct
// row selector bits set (as well as clock, latch, and OE all low).
void setPix(dbl_pixel_t *pixel) {
  pause();
  gpio0[GPIO_DATAOUT] = pixel->gpv0.word;
  gpio1[GPIO_DATAOUT] = pixel->gpv1.word;
  gpio2[GPIO_DATAOUT] = pixel->gpv2.word;
  gpio3[GPIO_DATAOUT] = pixel->gpv3.word;
}

// flash is used to display a solid 1-bit color as a warning or
// debugging aid.  if howmany is -1, it stays here indefinitely.
void flash(int cbits, int howmany) {
  uint32_t row;

  dbl_pixel_t pixel;
  memset(&pixel, 0, sizeof(pixel));

  if (cbits & CBITS_GREEN) {
    pixel.gpv1.bits.j3_g1 = 1;
    pixel.gpv0.bits.j3_g2 = 1;
    pixel.gpv2.bits.j1_g1 = 1;
    pixel.gpv2.bits.j1_g2 = 1;
  }

  if (cbits & CBITS_BLUE) {
    pixel.gpv0.bits.j3_b1 = 1;
    pixel.gpv0.bits.j3_b2 = 1;
    pixel.gpv2.bits.j1_b1 = 1;
    pixel.gpv0.bits.j1_b2 = 1;
  }

  if (cbits & CBITS_RED) {
    pixel.gpv0.bits.j3_r1 = 1;
    pixel.gpv1.bits.j3_r2 = 1;
    pixel.gpv2.bits.j1_r1 = 1;
    pixel.gpv0.bits.j1_r2 = 1;
  }

  int count;
  for (count = 0; howmany < 0 || count < howmany; count++) {
    //__delay_cycles(100000);
    for (row = 0; row < FRAMEBUF_SCANS; row++) {
      uint32_t pix;

      pixel.gpv1.bits.rowSelect = row;

      for (pix = 0; pix < FRAMEBUF_WIDTH; pix++) {
        setPix(&pixel);
        toggleClock();
      }
      latchRows(row);
    }
  }
}

// warn flashes a ~1 second warning
void warn(int cbits) { flash(cbits, 5000); }

// park stops the program w/ a certain color
void park(int cbits) { flash(cbits, -1); }

// setup_param initializes all the EDMA PaRAM entries to 0.
// then it sets the constant fields for a single PaRAM entry
// corresponding with the single DMA channel in use.
//
// Note this is used at runtime and is not an efficient approach to
// DMA.  TODO: Really, the application here calls for a continuous DMA
// loop!  Every 4th row is currently a little brighter than the others
// because of the extra time spent in dma_wait().
void setup_param() {
  // Setup and store PaRAM set for transfer.
  uint16_t paramOffset = EDMA_PARAM_OFFSET;
  paramOffset += ((dmaChannel * EDMA_PARAM_SIZE) / WORDSZ);

  edma_param_entry = (volatile edmaParam *)(EDMA_BASE + paramOffset);

  edma_param_entry->lnkrld.link = 0xFFFF;
  edma_param_entry->lnkrld.bcntrld = 0x0000;
  edma_param_entry->opt.static_set = 1;

  // Transfer complete interrupt enable.
  edma_param_entry->opt.tcinten = 1;

  // Intermediate transfer completion chaining enable.
  // not needed, used for splitting the transfer
  // edma_param_entry->opt.itcchen = 1;
  edma_param_entry->opt.tcc = dmaChannel;

  edma_param_entry->ccnt.ccnt = 1;
  edma_param_entry->abcnt.acnt = 1 << 12;
  edma_param_entry->abcnt.bcnt = 1;
  edma_param_entry->bidx.srcbidx = 0;
  edma_param_entry->bidx.dstbidx = 0;
  edma_param_entry->src = resourceTable.framebufs.pa;
  edma_param_entry->dst = PRU_L4_FAST_SHARED_PRUSS_MEM;
}

// setup_dma_channel_zero tries to reset and clear any pending
// interrupt and error states before we start.
void setup_dma_channel_zero() {
  // Zero the PaRAM entries.
  memset((void *)(EDMA_BASE + EDMA_PARAM_OFFSET / WORDSZ), 0, EDMA_PARAM_SIZE * EDMA_PARAM_NUM);

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
  EDMA_BASE[SHADOW1(EDMAREG_SECR)] |= dmaChannelMask;
  EDMA_BASE[SHADOW1(EDMAREG_ICR)] |= dmaChannelMask;

  // Disable the queue watermark (in case that's triggering CCERR?)
  EDMA_BASE[EDMA_QWMTHRA] = 0x11 | (0x11 << 8) | (0x11 << 16);

  // Enable channel interrupt.
  EDMA_BASE[SHADOW1(EDMAREG_IESR)] |= dmaChannelMask;

  // Clear channel controller errors.  (Indiscriminantly. TODO: there
  // are 2 distinct kinds of error here.)
  EDMA_BASE[EDMA_CCERRCLR] |= 0xffffffff;

  // Enable channel for an event trigger.
  EDMA_BASE[SHADOW1(EDMAREG_EESR)] |= dmaChannelMask;

  // Clear event missed register.
  EDMA_BASE[EDMA_EMCR] |= dmaChannelMask;

  setup_param();
}

// start_dma calculates the address of the next block to transfer.
// The inputs are bank, frame, and frame part that the PRU is
// currently writing to the GPIOs.  This logic computes the next
// bank (checking for rollover) and begins a transfer.
//
// TODO: This is a heavyweight setup, we should be able to use
// transfer linking/chaining for a continuous loop.
uint32_t start_dma(uint32_t nextLocalIndex, uint32_t currentBank, uint32_t currentFrame, uint32_t currentPart) {

  int startOfBank = currentFrame == 0 && currentPart == 0;
  int endOfBank = currentFrame == FRAMEBUF_FRAMES_PER_BANK - 1 && currentPart == FRAMEBUF_PARTS_PER_FRAME - 1;

  if (startOfBank) {
    global_ctrl->start_bank = currentBank;
  }

  if (endOfBank) {
    currentBank = global_ctrl->ready_bank % 2; // % 2 is for safety
    currentFrame = 0;
    currentPart = 0;
  } else {
    currentPart++;
  }

  setup_param();

  edma_param_entry->dst = PRU_L4_FAST_SHARED_PRUSS_MEM + (nextLocalIndex * FRAMEBUF_PART_SIZE);
  edma_param_entry->src = resourceTable.framebufs.pa + (currentBank * FRAMEBUF_BANK_SIZE) +
                          (currentFrame * FRAMEBUF_FRAME_SIZE) + (currentPart * FRAMEBUF_PART_SIZE);

  edma_param_entry->ccnt.ccnt = 1;
  edma_param_entry->abcnt.acnt = 1 << 12;
  edma_param_entry->abcnt.bcnt = 1;
  edma_param_entry->opt.tcc = dmaChannel;

  // The equivalent blocking memory transfer: TODO: @@@ lol it's
  // nearly as fast as the slowest DMA method, hard to get working.
  memcpy((void *)edma_param_entry->dst, (void *)edma_param_entry->src, FRAMEBUF_PART_SIZE);

  return currentBank;

  // Trigger transfer.  (4.4.1.2.2 Event Interface Mapping)
  // This is pr1_pru_mst_intr[2]_intr_req, system event 18
  // __R31 = R31_INTERRUPT_ENABLE | (SYSEVT_PRU_TO_EDMA - R31_INTERRUPT_OFFSET);
  // return currentBank;
}

// wait_dma as you see, has some bugs.  Most likely, the problems
// originate from the use of a 4KB transfer above, which means the
// transfer controller's queues are filling to their high watermark.
// We see CCERR, EMR as well as several kernel-level issues related
// to DMA completion events!
uint32_t wait_dma(uint32_t *restart) {
  uint32_t wait = 0;

  // Note: understand how "omap_intc_handle_irq: spurious irq!" comes about (kernel 4.19?)
  // Note: kernel is unhappy with "virtio_rpmsg_bus virtio0: msg received with no recipient"

  if (EDMA_BASE[EDMA_CCERR] != 0) {
    park(CBITS_CYAN);
  }

  if (EDMA_BASE[EDMA_EMR]) {
    warn(CBITS_YELLOW);
  }
  if (EDMA_BASE[EDMA_EMRH]) {
    warn(CBITS_CYAN);
  }

  if (__R31 & PRU_R31_INTERRUPT_FROM_ARM) {

    // Clear the interrupt event.  It could be one of two kinds of
    // error from the EDMA controller or it could be the ARM kicking.
    if (CT_INTC.SECR1_bit.ENA_STS_63_32 & (1 << (SYSEVT_EDMA_CTRL_ERROR_TO_PRU - 32))) {
      warn(CBITS_BLUE);
      warn(CBITS_RED);
      CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_CTRL_ERROR_TO_PRU;
    } else if (CT_INTC.SECR1_bit.ENA_STS_63_32 & (1 << (SYSEVT_EDMA_CHAN_ERROR_TO_PRU - 32))) {
      park(CBITS_BLUE);
      CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_CHAN_ERROR_TO_PRU;
    } else if (CT_INTC.SECR0_bit.ENA_STS_31_0 & (1 << SYSEVT_ARM_TO_PRU)) {
      // This means the control program restarted, needs to know carveout addresses.
      *restart = 1;

      /* warn(CBITS_BLUE); */
      warn(CBITS_GREEN);
      CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_ARM_TO_PRU;
    } else {
      /* warn(CBITS_BLUE); */
      /* warn(CBITS_WHITE); */
    }
  } else {
    // warn(CBITS_GREEN);
    // warn(CBITS_WHITE);
  }

#if 1
  // Note: deferred this until past the restart signal
  if (1) {
    //*restart = 1;
    return 0;
  }
#endif

  while (!(__R31 & PRU_R31_INTERRUPT_FROM_EDMA)) {
    wait++;
    warn(CBITS_GREEN);
    warn(CBITS_YELLOW);
  }

  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_TO_PRU;

  EDMA_BASE[SHADOW1(EDMAREG_ICR)] = dmaChannelMask;

  return wait;
}

// reset_hardware_state enables clears PRU-shared memory, starts the
// cycle counter, clears system events we're going to listen for,
// resets the GPIO bits, etc.
void reset_hardware_state() {
  // Allow OCP master port access by the PRU.
  CT_CFG.SYSCFG_bit.STANDBY_INIT = 0;

  // Enable PRU0 cycle counter.
  PRU0_CTRL.CTRL_bit.CTR_EN = 1;

  // Clear the system event mapped to the two input interrupts.
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_ARM_TO_PRU;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_PRU_TO_ARM;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_TO_PRU;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_PRU_TO_EDMA;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_CTRL_ERROR_TO_PRU;
  CT_INTC.SICR_bit.STS_CLR_IDX = SYSEVT_EDMA_CHAN_ERROR_TO_PRU;

  // Enable the EDMA (Transfer controller, Channel controller) clocks.
  CM_PER_BASE[CM_PER_TPTC0_CLKCTRL] = CM_PER_CLK_ENABLED;
  CM_PER_BASE[CM_PER_TPCC_CLKCTRL] = CM_PER_CLK_ENABLED;

  // Reset gpio output.
  const uint32_t allbits = 0x00000000;
  gpio0[GPIO_CLEARDATAOUT] = allbits;
  gpio1[GPIO_CLEARDATAOUT] = allbits;
  gpio2[GPIO_CLEARDATAOUT] = allbits;
  gpio3[GPIO_CLEARDATAOUT] = allbits;

  // Reset the local shared memory buffer.
  memset((void *)0x10000, 0, 0x3000);

  pause();
  latch(LO);
  pause();
  clock(LO);
  pause();
  outputEnable(HI);
  pause();
  setRow(0);
  pause();
  outputEnable(LO);
}

// init_test_buffer sets the PRU-local pixels to blue and the
// framebuffer to a red-green-blue checkerboard pattern.
void init_test_buffer() {
  // Turn off CLK, OE, LATCH pins and set the correct row number
  // for each pixel.
  uint32_t pix, row;
  dbl_pixel_t *pixptr = (dbl_pixel_t *)0x10000; // Base 8kB of PRU shared storage.

  // The PRU-local storage is initialized with all blue pixels.  If
  // for some reason the DMA is not succesful, these pixels represent
  // rows 0-15 blue, rows 16-31 dark.
  for (row = 0; row < (2 * FRAMEBUF_SCANS_PER_PART); row++) {
    for (pix = 0; pix < 64; pix++) {
      pixptr->gpv1.bits.rowSelect = row;
      pixptr->gpv1.bits.inputClock = 0;
      pixptr->gpv1.bits.outputEnable = 0;
      pixptr->gpv1.bits.inputLatch = 0;

      pixptr->gpv2.bits.j1_r1 = 0;
      pixptr->gpv2.bits.j1_g1 = 0;
      pixptr->gpv2.bits.j1_b1 = 1;
      pixptr->gpv0.bits.j1_r2 = 0;
      pixptr->gpv2.bits.j1_g2 = 0;
      pixptr->gpv0.bits.j1_b2 = 1;

      pixptr->gpv0.bits.j2_r1 = 0;
      pixptr->gpv2.bits.j2_g1 = 0;
      pixptr->gpv0.bits.j2_b1 = 1;
      pixptr->gpv2.bits.j2_r2 = 0;
      pixptr->gpv2.bits.j2_g2 = 0;
      pixptr->gpv2.bits.j2_b2 = 1;

      pixptr->gpv0.bits.j3_r1 = 0;
      pixptr->gpv1.bits.j3_g1 = 0;
      pixptr->gpv0.bits.j3_b1 = 1;
      pixptr->gpv1.bits.j3_r2 = 0;
      pixptr->gpv0.bits.j3_g2 = 0;
      pixptr->gpv0.bits.j3_b2 = 1;

      pixptr->gpv0.bits.j4_r1 = 0;
      pixptr->gpv0.bits.j4_g1 = 0;
      pixptr->gpv1.bits.j4_b1 = 1;
      pixptr->gpv3.bits.j4_r2 = 0;
      pixptr->gpv3.bits.j4_g2 = 0;
      pixptr->gpv0.bits.j4_b2 = 1;

      pixptr->gpv2.bits.j5_r1 = 0;
      pixptr->gpv0.bits.j5_g1 = 0;
      pixptr->gpv0.bits.j5_b1 = 1;
      pixptr->gpv0.bits.j5_r2 = 0;
      pixptr->gpv0.bits.j5_g2 = 0;
      pixptr->gpv2.bits.j5_b2 = 1;

      pixptr->gpv2.bits.j6_r1 = 0;
      pixptr->gpv2.bits.j6_g1 = 0;
      pixptr->gpv2.bits.j6_b1 = 1;
      pixptr->gpv2.bits.j6_r2 = 0;
      pixptr->gpv2.bits.j6_g2 = 0;
      pixptr->gpv2.bits.j6_b2 = 1;

      pixptr->gpv2.bits.j7_r1 = 0;
      pixptr->gpv2.bits.j7_g1 = 0;
      pixptr->gpv2.bits.j7_b1 = 1;
      pixptr->gpv2.bits.j7_r2 = 0;
      pixptr->gpv3.bits.j7_g2 = 0;
      pixptr->gpv2.bits.j7_b2 = 1;

      pixptr->gpv3.bits.j8_r1 = 0;
      pixptr->gpv3.bits.j8_g1 = 0;
      pixptr->gpv3.bits.j8_b1 = 1;
      pixptr->gpv3.bits.j8_r2 = 0;
      pixptr->gpv0.bits.j8_g2 = 0;
      pixptr->gpv3.bits.j8_b2 = 1;

      pixptr++;
    }
  }

  // Zero the framebuffer before setting the test pattern.
  memset((void *)resourceTable.framebufs.pa, 0, FRAMEBUF_TOTAL_SIZE);
  uint32_t bankno;
  for (bankno = 0; bankno < 2; bankno++) {
    // For 256 frames per bank
    uint32_t frame;
    dbl_pixel_t *pixptr = (dbl_pixel_t *)(resourceTable.framebufs.pa + bankno * FRAMEBUF_BANK_SIZE);

    for (frame = 0; frame < FRAMEBUF_FRAMES_PER_BANK; frame++) {

      uint32_t row;

      for (row = 0; row < FRAMEBUF_SCANS; row++) {
        uint32_t pix;

        // This draws a blue checkerboard pattern with alternating red
        // and green squares.
        for (pix = 0; pix < 64; pix++) {
          pixptr->gpv1.bits.rowSelect = row;
          pixptr->gpv1.bits.inputClock = 0;
          pixptr->gpv1.bits.outputEnable = 0;
          pixptr->gpv1.bits.inputLatch = 0;

          uint32_t quad = (pix >> 4) & 1;

          pixptr->gpv2.bits.j1_r1 = 1 ^ quad;
          pixptr->gpv2.bits.j1_g1 = 0;
          pixptr->gpv2.bits.j1_b1 = 0 ^ quad;
          pixptr->gpv0.bits.j1_r2 = 0;
          pixptr->gpv2.bits.j1_g2 = 0 ^ quad;
          pixptr->gpv0.bits.j1_b2 = 1 ^ quad;

          pixptr->gpv0.bits.j2_r1 = 1 ^ quad;
          pixptr->gpv2.bits.j2_g1 = 0;
          pixptr->gpv0.bits.j2_b1 = 0 ^ quad;
          pixptr->gpv2.bits.j2_r2 = 0;
          pixptr->gpv2.bits.j2_g2 = 0 ^ quad;
          pixptr->gpv2.bits.j2_b2 = 1 ^ quad;

          pixptr->gpv0.bits.j3_r1 = 1 ^ quad;
          pixptr->gpv1.bits.j3_g1 = 0;
          pixptr->gpv0.bits.j3_b1 = 0 ^ quad;
          pixptr->gpv1.bits.j3_r2 = 0;
          pixptr->gpv0.bits.j3_g2 = 0 ^ quad;
          pixptr->gpv0.bits.j3_b2 = 1 ^ quad;

          pixptr->gpv0.bits.j4_r1 = 1 ^ quad;
          pixptr->gpv0.bits.j4_g1 = 0;
          pixptr->gpv1.bits.j4_b1 = 0 ^ quad;
          pixptr->gpv3.bits.j4_r2 = 0;
          pixptr->gpv3.bits.j4_g2 = 0 ^ quad;
          pixptr->gpv0.bits.j4_b2 = 1 ^ quad;

          pixptr->gpv2.bits.j5_r1 = 1 ^ quad;
          pixptr->gpv0.bits.j5_g1 = 0;
          pixptr->gpv0.bits.j5_b1 = 0 ^ quad;
          pixptr->gpv0.bits.j5_r2 = 0;
          pixptr->gpv0.bits.j5_g2 = 0 ^ quad;
          pixptr->gpv2.bits.j5_b2 = 1 ^ quad;

          pixptr->gpv2.bits.j6_r1 = 1 ^ quad;
          pixptr->gpv2.bits.j6_g1 = 0;
          pixptr->gpv2.bits.j6_b1 = 0 ^ quad;
          pixptr->gpv2.bits.j6_r2 = 0;
          pixptr->gpv2.bits.j6_g2 = 0 ^ quad;
          pixptr->gpv2.bits.j6_b2 = 1 ^ quad;

          pixptr->gpv2.bits.j7_r1 = 1 ^ quad;
          pixptr->gpv2.bits.j7_g1 = 0;
          pixptr->gpv2.bits.j7_b1 = 0 ^ quad;
          pixptr->gpv2.bits.j7_r2 = 0;
          pixptr->gpv3.bits.j7_g2 = 0 ^ quad;
          pixptr->gpv2.bits.j7_b2 = 1 ^ quad;

          pixptr->gpv3.bits.j8_r1 = 1 ^ quad;
          pixptr->gpv3.bits.j8_g1 = 0;
          pixptr->gpv3.bits.j8_b1 = 0 ^ quad;
          pixptr->gpv3.bits.j8_r2 = 0;
          pixptr->gpv0.bits.j8_g2 = 0 ^ quad;
          pixptr->gpv3.bits.j8_b2 = 1 ^ quad;

          pixptr++;
        }
      }
    }
  }
}

// wait_for_virtio_ready waits for Linux drivers to be ready for RPMsg communication.
void wait_for_virtio_ready() {
  volatile uint8_t *status = &resourceTable.rpmsg_vdev.status;
  while (!(*status & VIRTIO_CONFIG_S_DRIVER_OK)) {
    // Wait
  }
}

// setup_transport opens the RPMsg channel to the ARM host.
void setup_transport() {
  // Using the name 'rpmsg-pru' will probe the rpmsg_pru driver found
  // at linux/drivers/rpmsg/rpmsg_pru.c
  char channel_name[32] = "rpmsg-pru";
  const int channel_port = 30;

  // Initialize two vrings using system events on dedicated channels.
  pru_rpmsg_init(&rpmsg_transport, &resourceTable.rpmsg_vring0, &resourceTable.rpmsg_vring1, SYSEVT_PRU_TO_ARM,
                 SYSEVT_ARM_TO_PRU);

  // Create the RPMsg channel between the PRU and the ARM.
  while (pru_rpmsg_channel(RPMSG_NS_CREATE, &rpmsg_transport, channel_name, channel_port) != PRU_RPMSG_SUCCESS) {
  }
}

// send_to_arm sends the carveout addresses to the ARM.
void send_to_arm() {
  if (pru_rpmsg_receive(&rpmsg_transport, &rpmsg_src, &rpmsg_dst, rpmsg_payload, &rpmsg_len) != PRU_RPMSG_SUCCESS) {
    return;
  }
  memcpy(rpmsg_payload, &resourceTable.controls.pa, 4);
  while (pru_rpmsg_send(&rpmsg_transport, rpmsg_dst, rpmsg_src, rpmsg_payload, 4) != PRU_RPMSG_SUCCESS) {
  }
}

// setup_controls builds the control struct.  The address of this is
// passed to the ARM as a 4-byte write.
control_t *setup_controls() {
  global_ctrl = (control_t *)resourceTable.controls.pa;
  memset(global_ctrl, 0, sizeof(control_t));
  global_ctrl->framebufs_addr = resourceTable.framebufs.pa;
  global_ctrl->framebufs_size = FRAMEBUF_TOTAL_SIZE;
  return global_ctrl;
}

void main(void) {
  reset_hardware_state();

  wait_for_virtio_ready();

  control_t *ctrl = setup_controls();

  setup_transport();

  setup_dma_channel_zero();

  init_test_buffer();

  frame_banks[0] = (dbl_pixel_t *)(resourceTable.framebufs.pa);
  frame_banks[1] = (dbl_pixel_t *)(resourceTable.framebufs.pa + FRAMEBUF_BANK_SIZE);

  local_banks[0] = (dbl_pixel_t *)0x10000;
  local_banks[1] = (dbl_pixel_t *)0x11000;

  uint32_t unused;

  // Fill the first bank.
  start_dma(0, 1, FRAMEBUF_FRAMES_PER_BANK - 1, FRAMEBUF_PARTS_PER_FRAME - 1);
  wait_dma(&unused);

  // For two banks
  uint32_t localIndex = 0;
  uint32_t currentBank = 0;
  while (1) {
    // For 256 frames per bank
    uint32_t frame;
    uint32_t restart_signaled = 0;

    for (frame = 0; frame < FRAMEBUF_FRAMES_PER_BANK; frame++) {
      uint32_t row = 0;
      uint32_t part;

      // For 4 parts per frame
      for (part = 0; part < FRAMEBUF_PARTS_PER_FRAME; part++) {

        dbl_pixel_t *pixptr = local_banks[localIndex];

        localIndex ^= 1;

        // Start a DMA to fill the next local bank.
        currentBank = start_dma(localIndex, currentBank, frame, part);

        // For 4 scans per part
        uint32_t scan;
        for (scan = 0; scan < FRAMEBUF_SCANS_PER_PART; scan++, row++) {

          uint32_t pix;
          // For 64 pixels width
          for (pix = 0; pix < 64; pix++) {

            // Set 2 pixels
            setPix(pixptr++);
            toggleClock();
          }

          latchRows(row);

          // Slow down to see what's happening.
          // __delay_cycles(10000000);
        }
        wait_dma(&restart_signaled);
      }

      ctrl->framecount++;

      if (restart_signaled != 0) {
        send_to_arm();
      }
    }
  }
}
