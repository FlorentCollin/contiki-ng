#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/simple-udp.h"
#include "sys/log.h"
#include "sys/node-id.h"
#include "etimer.h"

#include "graph-application.h"
#include "topology-application.h"
#include "bandwidth-application.h"

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

// #define APP_UNICAST_TIMESLOT 1
// static struct tsch_slotframe *sf_common;

// static void initialize_tsch_schedule() {
//     sf_common = tsch_schedule_get_slotframe_by_handle(0);
//     tsch_schedule_add_link(
//         sf_common,
//         (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
//         LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
//     if (sf_common == NULL) {
//         LOG_ERR("Couldn't create the initial slotframe\n");
//     }
//     for (int i = 1; i < APP_SLOTFRAME_SIZE; i++) {
//         for (int j = 0; j < 4; j++) {
//             struct tsch_link *link = tsch_schedule_add_link(
//                 sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
//                 &tsch_broadcast_address, i, j, 1);

//             if (link == NULL) {
//                 LOG_ERR("Couldn't had the link with (timeslot, channel): (%d, %d)\n", i, j);
//             }
//         }
//     }
// }


PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    static struct etimer timer;
    PROCESS_BEGIN();
    // initialize_tsch_schedule();
    LOG_INFO("Proxy border started\n");

    etimer_set(&timer, 10 * CLOCK_SECOND);
    NETSTACK_ROUTING.root_start();
    topology_application_start();
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
    etimer_reset(&timer);
    graph_application_start();
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
    etimer_reset(&timer);
    bandwidth_application_start(10);
    PROCESS_END();
}
