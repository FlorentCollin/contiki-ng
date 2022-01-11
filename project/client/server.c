#include "coap-engine.h"
#include "contiki.h"
#include "dev/button-hal.h"
#include "lib/random.h"
#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "sys/log.h"
#include "sys/node-id.h"


#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

extern coap_resource_t res_event;

PROCESS(server_process, "Coap Server Example");
AUTOSTART_PROCESSES(&server_process);

PROCESS_THREAD(server_process, ev, data) {
    PROCESS_BEGIN();

    PROCESS_PAUSE();
    LOG_INFO("Starting Coap Example Server\n");
    coap_activate_resource(&res_event, "schedule");

    if (node_id == 1) {
        NETSTACK_ROUTING.root_start();
    }

    while (1) {
        PROCESS_YIELD();
        if (ev == button_hal_release_event) {
            LOG_INFO("Triggered\n");
            res_event.trigger();
            /*res_event.resume();*/
        }
    }

    PROCESS_END();
}
