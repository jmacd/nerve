/*
 * Copyright (C) 2015-2021 Texas Instruments Incorporated - http://www.ti.com/
 *
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 *	* Redistributions of source code must retain the above copyright
 *	  notice, this list of conditions and the following disclaimer.
 *
 *	* Redistributions in binary form must reproduce the above copyright
 *	  notice, this list of conditions and the following disclaimer in the
 *	  documentation and/or other materials provided with the
 *	  distribution.
 *
 *	* Neither the name of Texas Instruments Incorporated nor the names of
 *	  its contributors may be used to endorse or promote products derived
 *	  from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

// EDMA PARAM registers
typedef struct {
  uint32_t sam : 1;
  uint32_t dam : 1;
  uint32_t syncdim : 1;
  uint32_t static_set : 1;
  uint32_t : 4;
  uint32_t fwid : 3;
  uint32_t tccmode : 1;
  uint32_t tcc : 6;
  uint32_t : 2;
  uint32_t tcinten : 1;
  uint32_t itcinten : 1;
  uint32_t tcchen : 1;
  uint32_t itcchen : 1;
  uint32_t privid : 4;
  uint32_t : 3;
  uint32_t priv : 1;
} edmaParamOpt;

typedef struct {
  uint32_t acnt : 16;
  uint32_t bcnt : 16;
} edmaParamABcnt;

typedef struct {
  uint32_t srcbidx : 16;
  uint32_t dstbidx : 16;
} edmaParamBidx;

typedef struct {
  uint32_t link : 16;
  uint32_t bcntrld : 16;
} edmaParamLnkRld;

typedef struct {
  uint32_t srccidx : 16;
  uint32_t dstcidx : 16;
} edmaParamCidx;

typedef struct {
  uint32_t ccnt : 16;
  uint32_t : 16;
} edmaParamCcnt;

typedef struct {
  edmaParamOpt opt;
  uint32_t src;
  edmaParamABcnt abcnt;
  uint32_t dst;
  edmaParamBidx bidx;
  edmaParamLnkRld lnkrld;
  edmaParamCidx cidx;
  edmaParamCcnt ccnt;
} edmaParam;

#define WORDSZ sizeof(uint32_t)

// CM_PER_BASE is the Clock Module Peripheral base address.
// TRM 8.1.12.1
#define CM_PER_BASE ((volatile uint32_t *)(0x44E00000))

// Third-Party Transfer Controller Clock
// TRM 8.1.12.1
#define CM_PER_TPTC0_CLKCTRL (0x24 / WORDSZ)
#define CM_PER_TPCC_CLKCTRL (0xBC / WORDSZ)
#define CM_PER_CLK_ENABLED (0x2)

// EDMA constants

#define EDMA_BASE ((volatile uint32_t *)(0x49000000))

// Peripheral Identification Register Section 11.4.1.1
#define EDMA_PID 0

// EDMA3CC Configuration Register Section 11.4.1.2
#define EDMA_CCCFG (0x4 / WORDSZ)

// EDMA3CC System Configuration Register Section 15.1.3.2
#define EDMA_SYSCONFIG (0x10 / WORDSZ)

// Section 11.4.1.4 DMA Channel Mapping Registers
#define EDMA_DCHMAP_0 (0x100 / WORDSZ)  // DMA Channel Mapping Register 0
#define EDMA_DCHMAP_1 (0x104 / WORDSZ)  // DMA Channel Mapping Register 1
#define EDMA_DCHMAP_2 (0x108 / WORDSZ)  // DMA Channel Mapping Register 2
#define EDMA_DCHMAP_3 (0x10C / WORDSZ)  // DMA Channel Mapping Register 3
#define EDMA_DCHMAP_4 (0x110 / WORDSZ)  // DMA Channel Mapping Register 4
#define EDMA_DCHMAP_5 (0x114 / WORDSZ)  // DMA Channel Mapping Register 5
#define EDMA_DCHMAP_6 (0x118 / WORDSZ)  // DMA Channel Mapping Register 6
#define EDMA_DCHMAP_7 (0x11C / WORDSZ)  // DMA Channel Mapping Register 7
#define EDMA_DCHMAP_8 (0x120 / WORDSZ)  // DMA Channel Mapping Register 8
#define EDMA_DCHMAP_9 (0x124 / WORDSZ)  // DMA Channel Mapping Register 9
#define EDMA_DCHMAP_10 (0x128 / WORDSZ) // DMA Channel Mapping Register 10
#define EDMA_DCHMAP_11 (0x12C / WORDSZ) // DMA Channel Mapping Register 11
#define EDMA_DCHMAP_12 (0x130 / WORDSZ) // DMA Channel Mapping Register 12
#define EDMA_DCHMAP_13 (0x134 / WORDSZ) // DMA Channel Mapping Register 13
#define EDMA_DCHMAP_14 (0x138 / WORDSZ) // DMA Channel Mapping Register 14
#define EDMA_DCHMAP_15 (0x13C / WORDSZ) // DMA Channel Mapping Register 15
// ... (48 more)
// 100h to 1FCh DCHMAP_0..63 DMA Channel Mapping Registers 0-63

// Section 11.4.1.5 QDMA Channel Mapping Registers
#define EDMA_QCHMAP_0 (0x200 / WORDSZ) // QDMA Channel Mapping Register 0
#define EDMA_QCHMAP_1 (0x204 / WORDSZ) // QDMA Channel Mapping Register 1
#define EDMA_QCHMAP_2 (0x208 / WORDSZ) // QDMA Channel Mapping Register 2
#define EDMA_QCHMAP_3 (0x20C / WORDSZ) // QDMA Channel Mapping Register 3
#define EDMA_QCHMAP_4 (0x210 / WORDSZ) // QDMA Channel Mapping Register 4
#define EDMA_QCHMAP_5 (0x214 / WORDSZ) // QDMA Channel Mapping Register 5
#define EDMA_QCHMAP_6 (0x218 / WORDSZ) // QDMA Channel Mapping Register 6
#define EDMA_QCHMAP_7 (0x21C / WORDSZ) // QDMA Channel Mapping Register 7

// Section 11.4.1.6 DMA Queue Number Registers
#define EDMA_DMAQNUM_0 (0x240 / WORDSZ) // DMA Queue Number Register 0
#define EDMA_DMAQNUM_1 (0x244 / WORDSZ) // DMA Queue Number Register 1
#define EDMA_DMAQNUM_2 (0x248 / WORDSZ) // DMA Queue Number Register 2
#define EDMA_DMAQNUM_3 (0x24C / WORDSZ) // DMA Queue Number Register 3
#define EDMA_DMAQNUM_4 (0x250 / WORDSZ) // DMA Queue Number Register 4
#define EDMA_DMAQNUM_5 (0x254 / WORDSZ) // DMA Queue Number Register 5
#define EDMA_DMAQNUM_6 (0x258 / WORDSZ) // DMA Queue Number Register 6
#define EDMA_DMAQNUM_7 (0x25C / WORDSZ) // DMA Queue Number Register 7

#define EDMA_QDMAQNUM (0x260 / WORDSZ) // QDMA Queue Number Register Section 11.4.1.7
#define EDMA_QUEPRI (0x284 / WORDSZ)   // Queue Priority Register Section 11.4.1.8
#define EDMA_EMR (0x300 / WORDSZ)      // Event Missed Register Section 11.4.1.9
#define EDMA_EMRH (0x304 / WORDSZ)     // Event Missed Register High Section 11.4.1.10
#define EDMA_EMCR (0x308 / WORDSZ)     // Event Missed Clear Register Section 11.4.1.11
#define EDMA_EMCRH (0x30C / WORDSZ)    // Event Missed Clear Register High Section 11.4.1.12
#define EDMA_QEMR (0x310 / WORDSZ)     // QDMA Event Missed Register Section 11.4.1.13
#define EDMA_QEMCR (0x314 / WORDSZ)    // QDMA Event Missed Clear Register Section 11.4.1.14
#define EDMA_CCERR (0x318 / WORDSZ)    // EDMA3CC Error Register Section 11.4.1.15
#define EDMA_CCERRCLR (0x31C / WORDSZ) // EDMA3CC Error Clear Register Section 11.4.1.16

#define EDMA_DRAE0 (0x340 / WORDSZ)  // DMA Region Access Enable Register for Region 0 Section 11.4.1.18
#define EDMA_DRAEH0 (0x344 / WORDSZ) // DMA Region Access Enable Register High for Region 0 Section 11.4.1.19
#define EDMA_DRAE1 (0x348 / WORDSZ)  // DMA Region Access Enable Register for Region 1 Section 11.4.1.20
#define EDMA_DRAEH1 (0x34C / WORDSZ) // DMA Region Access Enable Register High for Region 1 Section 11.4.1.21
#define EDMA_DRAE2 (0x350 / WORDSZ)  // DMA Region Access Enable Register for Region 2 Section 11.4.1.22
#define EDMA_DRAEH2 (0x354 / WORDSZ) // DMA Region Access Enable Register High for Region 2 Section 11.4.1.23
#define EDMA_DRAE3 (0x358 / WORDSZ)  // DMA Region Access Enable Register for Region 3 Section 11.4.1.24
#define EDMA_DRAEH3 (0x35C / WORDSZ) // DMA Region Access Enable Register High for Region 3 Section 11.4.1.25
#define EDMA_DRAE4 (0x360 / WORDSZ)  // DMA Region Access Enable Register for Region 4 Section 11.4.1.26
#define EDMA_DRAEH4 (0x364 / WORDSZ) // DMA Region Access Enable Register High for Region 4 Section 11.4.1.27
#define EDMA_DRAE5 (0x368 / WORDSZ)  // DMA Region Access Enable Register for Region 5 Section 11.4.1.28
#define EDMA_DRAEH5 (0x36C / WORDSZ) // DMA Region Access Enable Register High for Region 5 Section 11.4.1.29
#define EDMA_DRAE6 (0x370 / WORDSZ)  // DMA Region Access Enable Register for Region 6 Section 11.4.1.30
#define EDMA_DRAEH6 (0x374 / WORDSZ) // DMA Region Access Enable Register High for Region 6 Section 11.4.1.31
#define EDMA_DRAE7 (0x378 / WORDSZ)  // DMA Region Access Enable Register for Region 7 Section 11.4.1.32
#define EDMA_DRAEH7 (0x37C / WORDSZ) // DMA Region Access Enable Register High for Region 7 Section 11.4.1.33

// 380h to 39Ch QRAE_0 to QRAE_7 QDMA Region Access Enable Registers for Region 0-7 Section 11.4.1.34

#define EDMA_Q0E0 (0x400 / WORDSZ)  // Event Queue 0 Entry 0 Register Section 11.4.1.35
#define EDMA_Q0E1 (0x404 / WORDSZ)  // Event Queue 0 Entry 1 Register Section 11.4.1.36
#define EDMA_Q0E2 (0x408 / WORDSZ)  // Event Queue 0 Entry 2 Register Section 11.4.1.37
#define EDMA_Q0E3 (0x40C / WORDSZ)  // Event Queue 0 Entry 3 Register Section 11.4.1.38
#define EDMA_Q0E4 (0x410 / WORDSZ)  // Event Queue 0 Entry 4 Register Section 11.4.1.39
#define EDMA_Q0E5 (0x414 / WORDSZ)  // Event Queue 0 Entry 5 Register Section 11.4.1.40
#define EDMA_Q0E6 (0x418 / WORDSZ)  // Event Queue 0 Entry 6 Register Section 11.4.1.41
#define EDMA_Q0E7 (0x41C / WORDSZ)  // Event Queue 0 Entry 7 Register Section 11.4.1.42
#define EDMA_Q0E8 (0x420 / WORDSZ)  // Event Queue 0 Entry 8 Register Section 11.4.1.43
#define EDMA_Q0E9 (0x424 / WORDSZ)  // Event Queue 0 Entry 9 Register Section 11.4.1.44
#define EDMA_Q0E10 (0x428 / WORDSZ) // Event Queue 0 Entry 10 Register Section 11.4.1.45
#define EDMA_Q0E11 (0x42C / WORDSZ) // Event Queue 0 Entry 11 Register Section 11.4.1.46
#define EDMA_Q0E12 (0x430 / WORDSZ) // Event Queue 0 Entry 12 Register Section 11.4.1.47
#define EDMA_Q0E13 (0x434 / WORDSZ) // Event Queue 0 Entry 13 Register Section 11.4.1.48
#define EDMA_Q0E14 (0x438 / WORDSZ) // Event Queue 0 Entry 14 Register Section 11.4.1.49
#define EDMA_Q0E15 (0x43C / WORDSZ) // Event Queue 0 Entry 15 Register Section 11.4.1.50
#define EDMA_Q1E0 (0x440 / WORDSZ)  // Event Queue 1 Entry 0 Register Section 11.4.1.51
#define EDMA_Q1E1 (0x444 / WORDSZ)  // Event Queue 1 Entry 1 Register Section 11.4.1.52
#define EDMA_Q1E2 (0x448 / WORDSZ)  // Event Queue 1 Entry 2 Register Section 11.4.1.53
#define EDMA_Q1E3 (0x44C / WORDSZ)  // Event Queue 1 Entry 3 Register Section 11.4.1.54
#define EDMA_Q1E4 (0x450 / WORDSZ)  // Event Queue 1 Entry 4 Register Section 11.4.1.55
#define EDMA_Q1E5 (0x454 / WORDSZ)  // Event Queue 1 Entry 5 Register Section 11.4.1.56
#define EDMA_Q1E6 (0x458 / WORDSZ)  // Event Queue 1 Entry 6 Register Section 11.4.1.57
#define EDMA_Q1E7 (0x45C / WORDSZ)  // Event Queue 1 Entry 7 Register Section 11.4.1.58
#define EDMA_Q1E8 (0x460 / WORDSZ)  // Event Queue 1 Entry 8 Register Section 11.4.1.59
#define EDMA_Q1E9 (0x464 / WORDSZ)  // Event Queue 1 Entry 9 Register Section 11.4.1.60
#define EDMA_Q1E10 (0x468 / WORDSZ) // Event Queue 1 Entry 10 Register Section 11.4.1.61
#define EDMA_Q1E11 (0x46C / WORDSZ) // Event Queue 1 Entry 11 Register Section 11.4.1.62
#define EDMA_Q1E12 (0x470 / WORDSZ) // Event Queue 1 Entry 12 Register Section 11.4.1.63
#define EDMA_Q1E13 (0x474 / WORDSZ) // Event Queue 1 Entry 13 Register Section 11.4.1.64
#define EDMA_Q1E14 (0x478 / WORDSZ) // Event Queue 1 Entry 14 Register Section 11.4.1.65
#define EDMA_Q1E15 (0x47C / WORDSZ) // Event Queue 1 Entry 15 Register Section 11.4.1.66
#define EDMA_Q2E0 (0x480 / WORDSZ)  // Event Queue 2 Entry 0 Register Section 11.4.1.67
#define EDMA_Q2E1 (0x484 / WORDSZ)  // Event Queue 2 Entry 1 Register Section 11.4.1.68
#define EDMA_Q2E2 (0x488 / WORDSZ)  // Event Queue 2 Entry 2 Register Section 11.4.1.69
#define EDMA_Q2E3 (0x48C / WORDSZ)  // Event Queue 2 Entry 3 Register Section 11.4.1.70
#define EDMA_Q2E4 (0x490 / WORDSZ)  // Event Queue 2 Entry 4 Register Section 11.4.1.71
#define EDMA_Q2E5 (0x494 / WORDSZ)  // Event Queue 2 Entry 5 Register Section 11.4.1.72
#define EDMA_Q2E6 (0x498 / WORDSZ)  // Event Queue 2 Entry 6 Register Section 11.4.1.73
#define EDMA_Q2E7 (0x49C / WORDSZ)  // Event Queue 2 Entry 7 Register Section 11.4.1.74
#define EDMA_Q2E8 (0x4A0 / WORDSZ)  // Event Queue 2 Entry 8 Register Section 11.4.1.75
#define EDMA_Q2E9 (0x4A4 / WORDSZ)  // Event Queue 2 Entry 9 Register Section 11.4.1.76
#define EDMA_Q2E10 (0x4A8 / WORDSZ) // Event Queue 2 Entry 10 Register Section 11.4.1.77
#define EDMA_Q2E11 (0x4AC / WORDSZ) // Event Queue 2 Entry 11 Register Section 11.4.1.78
#define EDMA_Q2E12 (0x4B0 / WORDSZ) // Event Queue 2 Entry 12 Register Section 11.4.1.79
#define EDMA_Q2E13 (0x4B4 / WORDSZ) // Event Queue 2 Entry 13 Register Section 11.4.1.80
#define EDMA_Q2E14 (0x4B8 / WORDSZ) // Event Queue 2 Entry 14 Register Section 11.4.1.81
#define EDMA_Q2E15 (0x4BC / WORDSZ) // Event Queue 2 Entry 15 Register Section 11.4.1.82

// Section 11.4.1.83 Queue Status Registers
#define EDMA_QSTAT_0 (0x600 / WORDSZ) // Queue Status Register 0
#define EDMA_QSTAT_1 (0x604 / WORDSZ) // Queue Status Register 1
#define EDMA_QSTAT_2 (0x608 / WORDSZ) // Queue Status Register 2

#define EDMA_QWMTHRA (0x620 / WORDSZ) // Queue Watermark Threshold A Register Section 11.4.1.84
#define EDMA_CCSTAT (0x640 / WORDSZ)  // EDMA3CC Status Register Section 11.4.1.85
#define EDMA_MPFAR (0x800 / WORDSZ)   // Memory Protection Fault Address Register Section 11.4.1.86
#define EDMA_MPFSR (0x804 / WORDSZ)   // Memory Protection Fault Status Register Section 11.4.1.87
#define EDMA_MPFCR (0x808 / WORDSZ)   // Memory Protection Fault Command Register Section 11.4.1.88
#define EDMA_MPPAG (0x80C / WORDSZ)   // Memory Protection Page Attribute Register Global Section 11.4.1.89

// 810h to 82Ch MPPA_0 to MPPA_7 Memory Protection Page Attribute Registers Section 11.4.1.90

#define EDMA_ER (0x1000 / WORDSZ)    // Event Register Section 11.4.1.91
#define EDMA_ERH (0x1004 / WORDSZ)   // Event Register High Section 11.4.1.92
#define EDMA_ECR (0x1008 / WORDSZ)   // Event Clear Register Section 11.4.1.93
#define EDMA_ECRH (0x100C / WORDSZ)  // Event Clear Register High Section 11.4.1.94
#define EDMA_ESR (0x1010 / WORDSZ)   // Event Set Register Section 11.4.1.95
#define EDMA_ESRH (0x1014 / WORDSZ)  // Event Set Register High Section 11.4.1.96
#define EDMA_CER (0x1018 / WORDSZ)   // Chained Event Register Section 11.4.1.97
#define EDMA_CERH (0x101C / WORDSZ)  // Chained Event Register High Section 11.4.1.98
#define EDMA_EER (0x1020 / WORDSZ)   // Event Enable Register Section 11.4.1.99
#define EDMA_EERH (0x1024 / WORDSZ)  // Event Enable Register High Section 11.4.1.100
#define EDMA_EECR (0x1028 / WORDSZ)  // Event Enable Clear Register Section 11.4.1.101
#define EDMA_EECRH (0x102C / WORDSZ) // Event Enable Clear Register High Section 11.4.1.102
#define EDMA_EESR (0x1030 / WORDSZ)  // Event Enable Set Register Section 11.4.1.103
#define EDMA_EESRH (0x1034 / WORDSZ) // Event Enable Set Register High Section 11.4.1.104
#define EDMA_SER (0x1038 / WORDSZ)   // Secondary Event Register Section 11.4.1.105
#define EDMA_SERH (0x103C / WORDSZ)  // Secondary Event Register High Section 11.4.1.106
#define EDMA_SECR (0x1040 / WORDSZ)  // Secondary Event Clear Register Section 11.4.1.107
#define EDMA_SECRH (0x1044 / WORDSZ) // Secondary Event Clear Register High Section 11.4.1.108
#define EDMA_IER (0x1050 / WORDSZ)   // Interrupt Enable Register Section 11.4.1.109
#define EDMA_IERH (0x1054 / WORDSZ)  // Interrupt Enable Register High Section 11.4.1.110
#define EDMA_IECR (0x1058 / WORDSZ)  // Interrupt Enable Clear Register Section 11.4.1.111
#define EDMA_IECRH (0x105C / WORDSZ) // Interrupt Enable Clear Register High Section 11.4.1.112
#define EDMA_IESR (0x1060 / WORDSZ)  // Interrupt Enable Set Register Section 11.4.1.113
#define EDMA_IESRH (0x1064 / WORDSZ) // Interrupt Enable Set Register High Section 11.4.1.114
#define EDMA_IPR (0x1068 / WORDSZ)   // Interrupt Pending Register Section 11.4.1.115
#define EDMA_IPRH (0x106C / WORDSZ)  // Interrupt Pending Register High Section 11.4.1.116
#define EDMA_ICR (0x1070 / WORDSZ)   // Interrupt Clear Register Section 11.4.1.117
#define EDMA_ICRH (0x1074 / WORDSZ)  // Interrupt Clear Register High Section 11.4.1.118
#define EDMA_IEVAL (0x1078 / WORDSZ) // Interrupt Evaluate Register Section 11.4.1.119
#define EDMA_QER (0x1080 / WORDSZ)   // QDMA Event Register Section 11.4.1.120
#define EDMA_QEER (0x1084 / WORDSZ)  // QDMA Event Enable Register Section 11.4.1.121
#define EDMA_QEECR (0x1088 / WORDSZ) // QDMA Event Enable Clear Register Section 11.4.1.122
#define EDMA_QEESR (0x108C / WORDSZ) // QDMA Event Enable Set Register Section 11.4.1.123
#define EDMA_QSER (0x1090 / WORDSZ)  // QDMA Secondary Event Register Section 11.4.1.124
#define EDMA_QSECR (0x1094 / WORDSZ) // QDMA Secondary Event Clear Register Section 11.4.1.125

// (TRM 11.3.3.1)
#define EDMA_PARAM_OFFSET (0x4000 / WORDSZ)
#define EDMA_PARAM_SIZE sizeof(edmaParam) // 32 bytes

// bit 5 is the valid strobe to generate system events with __R31
#define R31_INTERRUPT_ENABLE (1 << 5)
#define R31_INTERRUPT_OFFSET 16

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
