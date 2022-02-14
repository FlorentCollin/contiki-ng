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
    etimer_set(&timer, 180 * CLOCK_SECOND);
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
    uint16_t index = encode_graph_application_type(packet_buffer);
    while (link != NULL) {
        if (link->parent == NULL) {
            link = uip_sr_node_next(link);
            continue;
        }

        NETSTACK_ROUTING.get_sr_node_ipaddr(&child_ipaddr, link);
        NETSTACK_ROUTING.get_sr_node_ipaddr(&parent_ipaddr, link->parent);
        memcpy(&packet_buffer[index], &child_ipaddr, sizeof(child_ipaddr));
        index += sizeof(child_ipaddr);
        memcpy(&packet_buffer[index], &parent_ipaddr, sizeof(parent_ipaddr));
        index += sizeof(parent_ipaddr);
        memcpy(&packet_buffer[index], &link->lifetime, sizeof(link->lifetime));
        index += sizeof(link->lifetime);

        link = uip_sr_node_next(link);
    }
    return index;
}


 /*static void send_rpl_neighbors() {*/
     /*static uint8_t data[128];*/
     /*rpl_nbr_t *nbr = nbr_table_head(rpl_neighbors);*/
     /*uip_ipaddr_t *nbr_ip;*/
     /*uint16_t index = 0;*/
     /*while(nbr != NULL) {*/
         /*nbr_ip = rpl_neighbor_get_ipaddr(nbr);*/
         /*memcpy(&data[index], nbr_ip, sizeof(*nbr_ip));*/
         /*index += sizeof(*nbr_ip);*/
         /*if (index > 128) {*/
             /*LOG_ERR("Trying to write more neighbors than allowed by the size of the buffer\n");*/
             /*return;*/
         /*}*/
         /*nbr = nbr_table_next(rpl_neighbors, nbr);*/
     /*}*/
     /*simple_udp_send(&udp_conn, &data, index);*/

 /*}*/