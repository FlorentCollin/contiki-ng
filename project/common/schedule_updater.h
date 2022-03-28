#ifndef SCHEDULE_UPDATER_H_
#define SCHEDULE_UPDATER_H_

#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "net/mac/tsch/tsch.h"
#include "net/mac/tsch/tsch-schedule.h"

#define SCHEDULE_UPDATER_SLOTFRAME_HANDLE 2
#define SCHEDULE_UPDATER_SLOTFRAME_SIZE 7

#define SCHEDULE_UPDATER_MAX_CELLS 20

#define UPDATE_PKT_TYPE_OFFSET 0
#define UPDATE_PKT_NEIGHBOR_ADDR_OFFSET 1
#define UPDATE_PKT_CELLS_COUNT_OFFSET 1 + sizeof(uip_ip6addr_t)
#define CELL_SIZE (sizeof(uint8_t) + 2 * sizeof(uint16_t))
#define UPDATE_PKT_CELL_START(cell_number) UPDATE_PKT_CELLS_COUNT_OFFSET + 1 + cell_number * CELL_SIZE

enum schedule_updater_pkt_type {
    schedule_updater_pkt_type_update,
    schedule_updater_pkt_type_update_complete,
};

struct cell {
    uint8_t link_options;
    uint16_t channel;
    uint16_t timeslot;
};

struct schedule_updater_pkt {
    enum schedule_updater_pkt_type type;
    uip_ip6addr_t neighbor_addr;
    uint8_t cell_count;
    struct cell cells[SCHEDULE_UPDATER_MAX_CELLS];
};

uint16_t schedule_updater_pkt_size_needed(struct schedule_updater_pkt *pkt);

void schedule_updater_pkt_encode(uint8_t *dest, struct schedule_updater_pkt *pkt);

enum schedule_updater_pkt_type update_pkt_type(const uint8_t *pkt_raw);

uip_ip6addr_t update_pkt_neighbor_addr(const uint8_t *pkt_raw);

uint8_t update_pkt_cells_count(const uint8_t *pkt_raw);

uint8_t update_pkt_cell_link_options(const uint8_t *pkt_raw, uint8_t cell_number);

uint16_t update_pkt_cell_timeslot(const uint8_t *pkt_raw, uint8_t cell_number);

uint16_t update_pkt_cell_channel(const uint8_t *pkt_raw, uint8_t cell_number);

void update_pkt_add_cells(const uint8_t *pkt_raw, struct tsch_slotframe *sf);

void update_pkt_dispatch(const uint8_t *pkt);

void update_pkt_log(const uint8_t *pkt);

void update_pkt_log_type(const uint8_t *pkt);

#endif /* SCHEDULE_UPDATER_H_ */
