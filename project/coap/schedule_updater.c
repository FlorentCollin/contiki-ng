#include "schedule_updater.h"

#include "contiki.h"
#include "sys/log.h"
#include "sys/node-id.h"

#define LOG_MODULE "schedule_updater"
#define LOG_LEVEL LOG_LEVEL_INFO

uint16_t schedule_updater_pkt_size_needed(struct schedule_updater_pkt *pkt) {
    return sizeof(pkt->neighbor_addr)              /* */
           + sizeof(pkt->cell_ids_len)             /* */
           + sizeof(uint8_t) * pkt->cell_ids_len   /* cell_id.link_options */
           + sizeof(uint16_t) * pkt->cell_ids_len  /* cell_id.channel */
           + sizeof(uint16_t) * pkt->cell_ids_len; /* cell_id.timeslot */
}

void schedule_updater_pkt_encode(uint8_t *dest, struct schedule_updater_pkt *pkt) {
    memcpy(&dest[0], &pkt->neighbor_addr.u16, 8);
    dest[8] = pkt->cell_ids_len;
    int index = 9;
    for (int i = 0; i < pkt->cell_ids_len; i++) {
        dest[index++] = pkt->cell_ids[i].link_options;
        dest[index++] = pkt->cell_ids[i].timeslot;
        dest[index++] = pkt->cell_ids[i].channel;
    }
}

void schedule_updater_pkt_decode(struct schedule_updater_pkt *pkt, const uint8_t *pkt_raw) {
    memcpy(&pkt->neighbor_addr, &pkt_raw[0], 8);
    pkt->cell_ids_len = pkt_raw[8];
    int index = 9;
    for (uint8_t i = 0; i < pkt->cell_ids_len; i++) {
        pkt->cell_ids[i].link_options = pkt_raw[index++];
        pkt->cell_ids[i].timeslot = pkt_raw[index++];
        pkt->cell_ids[i].channel = pkt_raw[index++];
    }
}

void schedule_updater_pkt_add_cells(struct schedule_updater_pkt *pkt, struct tsch_slotframe *sf) {
    for (uint8_t i = 0; i < pkt->cell_ids_len; i++) {
        struct cell_id cell = pkt->cell_ids[i];
        const linkaddr_t *neighbor_addr  = &pkt->neighbor_addr;
        struct tsch_link *err = tsch_schedule_add_link(sf, cell.link_options, LINK_TYPE_NORMAL, 
                neighbor_addr, cell.timeslot, cell.channel, 1);
        if (err == NULL) {
            LOG_INFO("Error while adding a new link\n");
        }
    }
}

void schedule_updater_pkt_log(struct schedule_updater_pkt *pkt) {
    LOG_INFO("schedule_updater_pkt log:\n");
    //LOG_INFO("  pkt->neighbor_addr = ");
    //LOG_INFO_6ADDR(&pkt->neighbor_addr);
    //LOG_PRINT("\n");
    LOG_INFO("  pkt->cell_ids_len = %d\n", pkt->cell_ids_len);
    LOG_INFO("  pkt->cell_ids =\n");
    for (int i = 0; i < pkt->cell_ids_len; i++) {
        LOG_INFO("       (%d) link_options = %d\n", i, pkt->cell_ids[i].link_options);
        LOG_INFO("       (%d) timeslot = %d\n", i, pkt->cell_ids[i].timeslot);
        LOG_INFO("       (%d) channel = %d\n", i, pkt->cell_ids[i].channel);
        LOG_INFO("\n");
    }
}
