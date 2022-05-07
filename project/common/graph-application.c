#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "net/ipv6/uip.h"
#include "net/ipv6/uip-sr.h"
#include "sys/node-id.h"
#include "sys/log.h"
#include "etimer.h"

#include "udpack-server.h"
#include "application-type.h"
#include "packet.h"

#define LOG_MODULE "GraphApplication"
#define LOG_LEVEL LOG_LEVEL_INFO

static uint16_t encode_graph(uint8_t* packet_buffer);

static uip_ipaddr_t* parent_ipaddr = NULL;

void rpl_parent_switch_callback(rpl_parent_t *old, rpl_parent_t *new) {
    parent_ipaddr = rpl_parent_get_ipaddr(new);
    if (parent_ipaddr != NULL) {
        LOG_INFO("Switching parent: ");
        LOG_INFO_6ADDR(parent_ipaddr);
        LOG_INFO("\n");
        send_server(AppTypeGraph, encode_graph);
    }
}


PROCESS(graph_application, "Graph Application");

void graph_application_start() { process_start(&graph_application, NULL); }

static struct etimer timer;
PROCESS_THREAD(graph_application, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Graph application started\n");
    etimer_set(&timer, 31 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        if (parent_ipaddr != NULL) {
            send_server(AppTypeGraph, encode_graph);
        }
        etimer_set(&timer, 30 * CLOCK_SECOND);
    }
    PROCESS_END();
}

static uint16_t encode_graph(uint8_t* packet_buffer) {
    LOG_INFO("Encoding the graph...\n");
    memcpy(packet_buffer, parent_ipaddr, sizeof(uip_ipaddr_t));
    return sizeof(uip_ipaddr_t);
}
