#include "topology-application.h"
#include "udpack-server.h"

#include "contiki.h"
#include "etimer.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/uip.h"
#include "net/routing/rpl-lite/rpl.h"
#include "sys/log.h"


#define LOG_MODULE "TopologyApplication"
#define LOG_LEVEL LOG_LEVEL_INFO

PROCESS(topology_application, "Topology Application");

void topology_application_start() { process_start(&topology_application, NULL); }

static uint16_t encode_topology(uint8_t* packet_buffer);
static uint16_t encode_topology_application_type(uint8_t* packet_buffer);
static uint16_t encode_rpl_neighbors(uint8_t* packet_buffer);

static struct etimer timer;
PROCESS_THREAD(topology_application, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Topology application started\n");
    etimer_set(&timer, 185 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        send_server(encode_topology);
        etimer_reset(&timer);
    }
    PROCESS_END();
}

static uint16_t encode_topology_application_type(uint8_t* packet_buffer) {
#define TOPOLOGY_APPLICATION_TYPE 1
    packet_buffer[0] = TOPOLOGY_APPLICATION_TYPE;
    return 1; /* size of application type */
}

static uint16_t encode_topology(uint8_t* packet_buffer) {
    LOG_INFO("Encoding the topology\n");
    uint16_t packet_size = encode_topology_application_type(packet_buffer);
    packet_size += encode_rpl_neighbors(packet_buffer + packet_size);
    return packet_size;
}

/* TODO: This function sends all the neighbor in one packet but this might not work if this node has
   to many neighbors. This should be refactor to send multiples packets. For testing it's ok
*/
static uint16_t encode_rpl_neighbors(uint8_t* packet_buffer) {
    rpl_nbr_t* nbr = nbr_table_head(rpl_neighbors);
    const linkaddr_t* nbr_ip;
    uint16_t index = 0;
    // The first lladdr in the packet is the lladdr of the host.
    memcpy(packet_buffer + index, &uip_lladdr, sizeof(uip_lladdr));
    index += sizeof(uip_lladdr);
    
    while (nbr != NULL) {
        nbr_ip = rpl_neighbor_get_lladdr(nbr);
        memcpy(packet_buffer + index, nbr_ip, sizeof(*nbr_ip));
        index += sizeof(*nbr_ip);
        nbr = nbr_table_next(rpl_neighbors, nbr);
    }
    return index;
}
