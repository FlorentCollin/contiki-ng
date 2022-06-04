#include "topology-application.h"

#include "contiki.h"
#include "etimer.h"
#include "net/ipv6/uip-ds6-nbr.h"
#include "net/ipv6/uip.h"
#include "net/routing/rpl-lite/rpl.h"
#include "sys/log.h"

#include "udpack-server.h"
#include "application-type.h"


#define LOG_MODULE "TopologyApplication"
#define LOG_LEVEL LOG_LEVEL_INFO

PROCESS(topology_application, "Topology Application");

void topology_application_start() { process_start(&topology_application, NULL); }

static uint16_t encode_topology(uint8_t* packet_buffer);
static uint16_t encode_neighbors(uint8_t* packet_buffer);

static struct etimer timer;
PROCESS_THREAD(topology_application, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Topology application started\n");
    etimer_set(&timer, 33 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        send_server(AppTypeTopology, encode_topology);
        etimer_set(&timer, 90 * CLOCK_SECOND);
    }
    PROCESS_END();
}

static uint16_t encode_topology(uint8_t* packet_buffer) {
    LOG_INFO("Encoding the topology\n");
    return encode_neighbors(packet_buffer);
}

/* This function sends all the neighbor in one packet but this might not work if this node has
   to many neighbors. This should be refactor to send multiples packets. 
   For testing and simulations it's ok though.
*/
static uint16_t encode_neighbors(uint8_t* packet_buffer) {
    uint16_t index = 0;
    // The first lladdr in the packet is the lladdr of the host.
    memcpy(packet_buffer + index, &uip_lladdr, sizeof(uip_lladdr));
    index += sizeof(uip_lladdr);

    // The rest is the neighbors lladdr.
    uip_ds6_nbr_t *nbr;
    linkaddr_t* nbr_lladdr;
    for(nbr = uip_ds6_nbr_head(); nbr != NULL; nbr = uip_ds6_nbr_next(nbr)) {
        nbr_lladdr = (linkaddr_t *) uip_ds6_nbr_lladdr_from_ipaddr(&nbr->ipaddr);
        memcpy(packet_buffer + index, nbr_lladdr, sizeof(*nbr_lladdr));
        index += sizeof(*nbr_lladdr);
    }
    LOG_INFO("TOPOLOGY size: %u\n", index);
    return index;
}
