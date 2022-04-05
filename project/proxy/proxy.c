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

PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    static struct etimer timer;
    PROCESS_BEGIN();
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
