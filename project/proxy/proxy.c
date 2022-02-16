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

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

#define UDP_CLIENT_PORT	8765
#define UDP_SERVER_PORT	3000

PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Proxy border started\n");

    if (node_id == 1) {
        NETSTACK_ROUTING.root_start();
        graph_application_start();
        topology_application_start();
    } else {
        LOG_ERR("The proxy border should be the node with id 1\n");
    }

    PROCESS_END();
}
