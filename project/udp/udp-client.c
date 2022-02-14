#include "contiki.h"
#include "udpack-server.h"
#include "etimer.h"
#include "net/ipv6/simple-udp.h"
#include "net/mac/tsch/tsch.h"
#include "net/netstack.h"
#include "net/routing/routing.h"
#include "schedule_updater.h"
#include "sys/log.h"

#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SEND_INTERVAL (20 * CLOCK_SECOND)

// static uint16_t highest_ack = 0;
// static void udp_rx_callback(struct simple_udp_connection *c,
//                             const uip_ipaddr_t *sender_addr,
//                             uint16_t sender_port,
//                             const uip_ipaddr_t *receiver_addr,
//                             uint16_t receiver_port,
//                             const uint8_t *data,
//                             uint16_t datalen) {
//     if (datalen < 2) {
//         LOG_ERR("Datalen is < 2 and therefore not packet number could be read from the data\n");
//         return;
//     }
//     uint16_t pkt_number = ((uint16_t)data[1] << 8) + data[0];
//     LOG_INFO("Received response with pkt_number %u \n", pkt_number);
//     if (highest_ack > pkt_number) {
//         LOG_INFO("ACK already sent, passing\n");
//         return;
//     }
//     LOG_INFO("Sending ack %u\n", pkt_number);
//     static uint8_t send_buffer[10] = {0};
//     send_buffer[0] = data[0];
//     send_buffer[1] = data[1];
//     simple_udp_sendto(c, send_buffer, 10, sender_addr);

//     if (pkt_number <= highest_ack) {
//         LOG_INFO("Packet already process, only sending ACK\n");
//         return;
//     }
//     highest_ack = pkt_number;

//     const uint8_t *pkt = data + 2;
//     update_pkt_log(pkt);
// }


#define HELLO_WORLD_APPLICATION 3
static uint16_t hello_world_encoder(uint8_t *buffer) {
    static uint8_t i = 0;
    LOG_INFO("Encoding hello world: %d\n", i);
    buffer[0] = HELLO_WORLD_APPLICATION;
    buffer[1] = i++;
    return 2;
}

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;
// static struct tsch_slotframe *sf_next;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
    tsch_schedule_add_link(
        sf_common,
        (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
        LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
    for (int i = 1; i < APP_SLOTFRAME_SIZE; i++) {
        for (int j = 0; j < 4; j++) {
            struct tsch_link *link = tsch_schedule_add_link(
                sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
                &tsch_broadcast_address, i, j, 1);

            if (link == NULL) {
                LOG_ERR("Couldn't had the link with (timeslot, channel): (%d, %d)\n", i, j);
            }
        }
    }
}

/*---------------------------------------------------------------------------*/
PROCESS(udp_client_process, "UDP client");
AUTOSTART_PROCESSES(&udp_client_process);
/*---------------------------------------------------------------------------*/
/*---------------------------------------------------------------------------*/
PROCESS_THREAD(udp_client_process, ev, data) {
    static struct etimer hello_timer;
    PROCESS_BEGIN();
    initialize_tsch_schedule();
    /* Initialize UDP connection */
    etimer_set(&hello_timer, 30 * CLOCK_SECOND);
    while (1) {
        PROCESS_YIELD();
        send_server(hello_world_encoder);
        etimer_reset(&hello_timer);
    }

    PROCESS_END();
}
/*---------------------------------------------------------------------------*/
