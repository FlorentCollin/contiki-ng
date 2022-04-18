#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/simple-udp.h"
#include "sys/log.h"
#include "sys/node-id.h"
#include "etimer.h"
#include "net/mac/tsch/tsch-schedule.h"

#include "graph-application.h"
#include "topology-application.h"
#include "bandwidth-application.h"

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    tsch_schedule_create_minimal();
    sf_common = tsch_schedule_get_slotframe_by_handle(0);
    if (sf_common == NULL) {
        LOG_ERR("Couldn't create the initial slotframe\n");
    }
    tsch_schedule_add_link(sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
    &tsch_broadcast_address, 1, 0, 1);
    tsch_schedule_add_link(sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
    &tsch_broadcast_address, 2, 0, 1);
}


PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    static struct etimer timer;
    PROCESS_BEGIN();
    initialize_tsch_schedule();
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
    while (1) {
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        // tsch_schedule_print();
        etimer_set(&timer, CLOCK_SECOND * 30);
    }
    PROCESS_END();
}
