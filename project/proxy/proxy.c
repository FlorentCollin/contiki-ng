#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "sys/log.h"
#include "sys/node-id.h"

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

PROCESS_THREAD(proxy_process, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Proxy border started\n");

    if (node_id == 1) {
        NETSTACK_ROUTING.root_start();
    } else {
        LOG_ERR("The proxy border should be the node with id 1\n");
    }
    PROCESS_END();
}
