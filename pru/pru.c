#include <rsc_types.h>

struct resource_table_hdr {
  struct resource_table header;

  uint32_t offset[3];

  struct {
    struct fw_rsc_hdr header;
    struct fw_rsc_carveout carveout;
  } carveout;

  struct {
    struct fw_rsc_hdr header;
    struct fw_rsc_trace trace;
  } trace;

  struct {
    struct fw_rsc_hdr header;
    struct fw_rsc_vdev vdev;
    struct fw_rsc_vdev_vring vrings[2];
    uint8_t config[0xc];
  } vdev;
};

const struct resource_table_hdr resource_table
__attribute__((used, section (".resource_table"))) = {
  .header = {
    .ver = 1,
    .num = ARRAY_SIZE(resource_table.offset), /* Number of resources */
  },

  .offset[0] = offsetof(struct resource_table_hdr, carveout),
  .offset[1] = offsetof(struct resource_table_hdr, trace),
  .offset[2] = offsetof(struct resource_table_hdr, vdev),

  .carveout = {
    .header = {
      .type = RSC_CARVEOUT,
    },
    .carveout = {
      .da = 0xf4000000,
      .len = 0x2000,
      .name = "firmware",
    },
  },

  /* Trace resource to printf() into */
  .trace = {
    .header = {
      .type = RSC_TRACE,
    },
    .trace = {
      .da = (uint32_t)trace_buf,
      .len = TRACE_BUFFER_SIZE,
      .name = "trace",
    },
  },

  /* VirtIO device */
  .vdev = {
    .header = {
      .type = RSC_VDEV,
    },
    .vdev = {
      .id = VIRTIO_ID_RPROC_SERIAL,
      .notifyid = 0,
      .dfeatures = 0,
      .config_len = 0xc,
      .num_of_vrings = 2,
    },
    .vrings = {
      [0] = {
	.align = 0x10,
	.num = 0x4,
	.notifyid = 0,
      },
      [1] = {
	.align = 0x10,
	.num = 0x4,
	.notifyid = 0,
      },
    },
  },
};

