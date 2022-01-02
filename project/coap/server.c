#include "coap-engine.h"
#include "contiki.h"
#include "lib/random.h"
#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "schedule_updater.h"
#include "sys/log.h"
#include "sys/node-id.h"

#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define UDP_PORT 8765
#define SEND_INTERVAL (60 * CLOCK_SECOND)

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(0, APP_SLOTFRAME_SIZE);
    /*tsch_schedule_add_link(*/
        /*sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),*/
        /*LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);*/
}

extern coap_resource_t res_hello;

PROCESS(server_process, "Coap Server Example");
AUTOSTART_PROCESSES(&server_process);

PROCESS_THREAD(server_process, ev, data) {
    PROCESS_BEGIN();

    PROCESS_PAUSE();
    LOG_INFO("Starting Coap Example Server\n");
    initialize_tsch_schedule();
    coap_activate_resource(&res_hello, "test/hello");

    if (node_id == 1) {
        NETSTACK_ROUTING.root_start();
    }

    PROCESS_END();
}
