#ifndef SCHEDULE_UPDATER_H_
#define SCHEDULE_UPDATER_H_

#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "net/mac/tsch/tsch.h"
#include "net/mac/tsch/tsch-schedule.h"

#define SCHEDULE_UPDATER_MAX_CELLS 10

struct cell_id {
    uint8_t link_options;
    uint16_t channel;
    uint16_t timeslot;
};

struct schedule_updater_pkt {
    linkaddr_t neighbor_addr;
    uint8_t cell_ids_len;
    struct cell_id cell_ids[SCHEDULE_UPDATER_MAX_CELLS];
};

// Acknowledgment that the receiver must send immediatly after receiving an update_schedule packet
// The ACK is represented by the pkt_number
typedef uint8_t schedule_updater_ack;

uint16_t schedule_updater_pkt_size_needed(struct schedule_updater_pkt *pkt);

void schedule_updater_pkt_encode(uint8_t *dest, struct schedule_updater_pkt *pkt);

void schedule_updater_pkt_decode(struct schedule_updater_pkt *pkt, const uint8_t *pkt_raw);

void schedule_updater_pkt_add_cells(struct schedule_updater_pkt *pkt, struct tsch_slotframe *sf);

void schedule_updater_pkt_log(struct schedule_updater_pkt *pkt);

#endif /* SCHEDULE_UPDATER_H_ */
