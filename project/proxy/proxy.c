#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-classic/rpl.h"
#include "net/routing/rpl-classic/rpl-dag-root.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/simple-udp.h"
#include "sys/log.h"
#include "sys/node-id.h"
#include "etimer.h"
#include "net/mac/tsch/tsch-schedule.h"
#include "node-id.h"

#include "topology-application.h"
#include "bandwidth-application.h"
#include "stats-print.h"

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    tsch_schedule_create_minimal();
    sf_common = tsch_schedule_get_slotframe_by_handle(0);
    if (sf_common == NULL) {
        LOG_ERR("Couldn't create the initial slotframe\n");
    }
}


/* Border router process */
PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    static struct etimer timer;
    PROCESS_BEGIN();
    NETSTACK_ROUTING.root_start();
    initialize_tsch_schedule();
    LOG_INFO("Proxy border started\n");

    etimer_set(&timer, 10 * CLOCK_SECOND);
    start_print_link_stats();
    topology_application_start();
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
    etimer_reset(&timer);
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
    etimer_reset(&timer);
    bandwidth_application_start(10);
    PROCESS_END();
}
