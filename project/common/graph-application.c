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
#include "packet.h"

#define LOG_MODULE "GraphApplication"
#define LOG_LEVEL LOG_LEVEL_INFO

PROCESS(graph_application, "Graph Application");

void graph_application_start() { process_start(&graph_application, NULL); }

static uint16_t encode_graph(uint8_t* packet_buffer);

static struct etimer timer;
PROCESS_THREAD(graph_application, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Graph application started\n");
    etimer_set(&timer, 185 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        send_server(encode_graph);
        etimer_reset(&timer);
    }
    PROCESS_END();
}

static uint16_t encode_graph_application_type(uint8_t* packet_buffer) {
#define GRAPH_APPLICATION_TYPE 0
    packet_buffer[0] = GRAPH_APPLICATION_TYPE;
    return 1;
}

static uint16_t encode_graph(uint8_t* packet_buffer) {
    LOG_INFO("Encoding the graph...\n");
    uip_sr_node_t* link;
    uip_ipaddr_t child_ipaddr;
    uip_ipaddr_t parent_ipaddr;
    link = uip_sr_node_head();
    uint16_t packet_size = encode_graph_application_type(packet_buffer);
    while (link != NULL) {
        if (link->parent == NULL) {
            link = uip_sr_node_next(link);
            continue;
        }

        NETSTACK_ROUTING.get_sr_node_ipaddr(&child_ipaddr, link);
        NETSTACK_ROUTING.get_sr_node_ipaddr(&parent_ipaddr, link->parent);
        memcpy(&packet_buffer[packet_size], &child_ipaddr, sizeof(child_ipaddr));
        packet_size += sizeof(child_ipaddr);
        memcpy(&packet_buffer[packet_size], &parent_ipaddr, sizeof(parent_ipaddr));
        packet_size += sizeof(parent_ipaddr);
        memcpy(&packet_buffer[packet_size], &link->lifetime, sizeof(link->lifetime));
        packet_size += sizeof(link->lifetime);

        link = uip_sr_node_next(link);
    }
    LOG_INFO("Graph size: %d\n", packet_size);
    return packet_size;
}