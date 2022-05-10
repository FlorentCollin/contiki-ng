#include "schedule_updater.h"

#include "contiki.h"
#include "sys/log.h"
#include "sys/node-id.h"
#include "net/ipv6/uip-ds6-nbr.h"
#include "net/mac/tsch/tsch-schedule.h"
#include "sys/ctimer.h"

#define LOG_MODULE "schedule_updater"
#define LOG_LEVEL LOG_LEVEL_INFO

uint16_t schedule_updater_pkt_size_needed(struct schedule_updater_pkt *pkt) {
    return sizeof(pkt->type) 
           + sizeof(pkt->neighbor_addr)
           + sizeof(pkt->cell_count)
           + sizeof(uint8_t) * pkt->cell_count
           + sizeof(uint16_t) * pkt->cell_count
           + sizeof(uint16_t) * pkt->cell_count;
}

void schedule_updater_pkt_encode(uint8_t *dest, struct schedule_updater_pkt *pkt) {
    dest[UPDATE_PKT_TYPE_OFFSET] = pkt->type;
    memcpy(&dest[UPDATE_PKT_NEIGHBOR_ADDR_OFFSET], &pkt->neighbor_addr, 8);
    dest[UPDATE_PKT_CELLS_COUNT_OFFSET] = pkt->cell_count;
    int index = UPDATE_PKT_CELL_START(0);
    int i;
    for (i = 0; i < pkt->cell_count; i++) {
        dest[index++] = pkt->cells[i].link_options;
        memcpy(&dest[index], &pkt->cells[i].timeslot, sizeof(pkt->cells[i].timeslot));
        index += sizeof(pkt->cells[i].timeslot);
        memcpy(&dest[index], &pkt->cells[i].channel, sizeof(pkt->cells[i].channel));
        index += sizeof(pkt->cells[i].channel);
    }
}

enum schedule_updater_pkt_type update_pkt_type(const uint8_t *pkt_raw) {
    return pkt_raw[UPDATE_PKT_TYPE_OFFSET];
}

linkaddr_t update_pkt_neighbor_addr(const uint8_t *pkt_raw) {
    linkaddr_t neighbor_addr;
    memcpy(&neighbor_addr, &pkt_raw[UPDATE_PKT_NEIGHBOR_ADDR_OFFSET], sizeof(neighbor_addr));
    return neighbor_addr;
}

uint8_t update_pkt_cells_count(const uint8_t *pkt_raw) {
    return pkt_raw[UPDATE_PKT_CELLS_COUNT_OFFSET];
}

uint8_t update_pkt_cell_link_options(const uint8_t *pkt_raw, uint8_t cell_number) {
    return pkt_raw[UPDATE_PKT_CELL_START(cell_number)];
}

uint16_t update_pkt_cell_timeslot(const uint8_t *pkt_raw, uint8_t cell_number) {
    uint16_t timeslot;
    memcpy(&timeslot, pkt_raw + UPDATE_PKT_CELL_START(cell_number) + sizeof(uint8_t), sizeof(timeslot));
    return timeslot;
    
}

uint16_t update_pkt_cell_channel(const uint8_t *pkt_raw, uint8_t cell_number) {
    uint16_t channel;
    memcpy(&channel, pkt_raw + UPDATE_PKT_CELL_START(cell_number) + sizeof(uint8_t) + sizeof(uint16_t), sizeof(channel));
    return channel;
}

void update_pkt_add_cells(const uint8_t *pkt, struct tsch_slotframe *sf) {
    uint8_t cell_count = update_pkt_cells_count(pkt);
    
    uint8_t i;
    for (i = 0; i < cell_count; i++) {
        uint8_t link_options = update_pkt_cell_link_options(pkt, i);
        uint16_t timeslot = update_pkt_cell_timeslot(pkt, i);
        uint16_t channel = update_pkt_cell_channel(pkt, i);
        linkaddr_t neighbor_addr = update_pkt_neighbor_addr(pkt);
        struct tsch_link *err = tsch_schedule_add_link(sf, link_options, LINK_TYPE_NORMAL, 
                &neighbor_addr, timeslot, channel, 1);
        if (err == NULL) {
            LOG_ERR("Error while adding a new link, tsch_schedule_add_link\n");
        }
    }
}

#define MAX_UPDATE_PACKET_SIZE 512
#define MAX_UPDATE_PACKETS_COUNT 10

static uint8_t update_packets[MAX_UPDATE_PACKET_SIZE * MAX_UPDATE_PACKETS_COUNT];
static uint8_t npacket = 0;
static uint16_t slotframe_handle = 1;
static struct tsch_slotframe* slotframe = NULL;

static void update_use_slotframe() {
    tsch_schedule_remove_slotframe(slotframe);
    slotframe = tsch_schedule_add_slotframe(slotframe_handle, TSCH_SCHEDULE_DEFAULT_LENGTH);
    if (slotframe == NULL) {
        LOG_ERR("The TSCH slotframe is NULL\n");
        return;
    }

    // inserting the new cells into the slotframe
    int i;
    for (i = 0; i < npacket; i++) {
        update_pkt_add_cells(update_packets + (i * MAX_UPDATE_PACKET_SIZE), slotframe);
    }
    npacket = 0;

    tsch_schedule_print();
}

void update_pkt_dispatch(const uint8_t *pkt, uint16_t packet_len) {
    update_pkt_log(pkt);
    static struct ctimer timer;
    switch (update_pkt_type(pkt)) {
        case schedule_updater_pkt_type_update:
            if (npacket >= MAX_UPDATE_PACKETS_COUNT) {
                LOG_ERR("Can't store more update schedule packet. Skipping this packet.\n");
                return;
            }
            if (packet_len > MAX_UPDATE_PACKET_SIZE) {
                LOG_ERR("The update packet size is too big to be stored. The packet len is %d, and the maximum packet size is %d\n", 
                        packet_len, MAX_UPDATE_PACKET_SIZE);
                return;
            }

            memcpy(update_packets + (npacket * MAX_UPDATE_PACKET_SIZE), pkt, packet_len);
            npacket++;
            break;
        case schedule_updater_pkt_type_update_complete:
            ctimer_set(&timer, 60 * 60 * CLOCK_SECOND, update_use_slotframe, NULL);
            break;
    }
}

void update_pkt_log(const uint8_t *pkt) {
    LOG_INFO("schedule_updater_pkt log:\n");
    LOG_INFO("  pkt->type_num = %d\n", update_pkt_type(pkt));
    switch (update_pkt_type(pkt)) {
        case schedule_updater_pkt_type_update_complete:
            LOG_INFO("  pkt->type = complete\n");
            break;
        case schedule_updater_pkt_type_update:
            LOG_INFO("  pkt->type = update\n");
            LOG_INFO("  pkt->neighbor_addr = ");
            linkaddr_t neighbor_addr = update_pkt_neighbor_addr(pkt);
            LOG_INFO_LLADDR(&neighbor_addr);
            LOG_PRINT("\n");
            LOG_INFO("  pkt->cell_count = %d\n", update_pkt_cells_count(pkt));
            LOG_INFO("  pkt->cell_ids =\n");
            uint8_t i;
            for (i = 0; i < update_pkt_cells_count(pkt); i++) {
                LOG_INFO("       (%d) link_options = %d\n", i, update_pkt_cell_link_options(pkt, i));
                LOG_INFO("       (%d) timeslot = %d\n", i, update_pkt_cell_timeslot(pkt, i));
                LOG_INFO("       (%d) channel = %d\n", i, update_pkt_cell_channel(pkt, i));
                printf("\n");
            }
            break;
    }
}

void update_pkt_log_type(const uint8_t *pkt) {
    switch (update_pkt_type(pkt)) {
        case schedule_updater_pkt_type_update_complete:
            LOG_WARN("pkt->type = complete\n");
            break;
        case schedule_updater_pkt_type_update:
            LOG_WARN("pkt->type = update\n");
            break;
    }
}
