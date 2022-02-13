#include "contiki.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/simple-udp.h"
#include "sys/log.h"
#include "sys/node-id.h"
#include "etimer.h"

#define LOG_MODULE "Proxy"
#define LOG_LEVEL LOG_LEVEL_INFO

#define UDP_CLIENT_PORT	8765
#define UDP_SERVER_PORT	3000

PROCESS(proxy_process, "Proxy Border");
AUTOSTART_PROCESSES(&proxy_process);

static struct simple_udp_connection udp_conn;
static struct etimer timer;

//  static void send_rpl_graph() {
//      static uint8_t data[38];
//      uip_sr_node_t *link;
//      uip_ipaddr_t child_ipaddr;
//      uip_ipaddr_t parent_ipaddr;
//      link = uip_sr_node_head();
//      while(link != NULL) {
//          if (link->parent == NULL) {
//              link = uip_sr_node_next(link);
//              continue;
//          }

//          NETSTACK_ROUTING.get_sr_node_ipaddr(&child_ipaddr, link);
//          NETSTACK_ROUTING.get_sr_node_ipaddr(&parent_ipaddr, link->parent);
//          uint16_t index = 0;
//          memcpy(&data[index], &child_ipaddr, sizeof(child_ipaddr));
//          index += sizeof(child_ipaddr);
//          memcpy(&data[index], &parent_ipaddr, sizeof(parent_ipaddr));
//          index += sizeof(parent_ipaddr);
//          memcpy(&data[index], &link->lifetime, sizeof(link->lifetime));
//          index += sizeof(link->lifetime);

//          simple_udp_send(&udp_conn, &data, index);
//          link = uip_sr_node_next(link);
//      }
//  }

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

PROCESS_THREAD(proxy_process, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Proxy border started\n");

    uip_ipaddr_t server_addr = {{0xfd, 0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}};
    simple_udp_register(&udp_conn, UDP_CLIENT_PORT, &server_addr,
                        UDP_SERVER_PORT, NULL);


    etimer_set(&timer, 180 * CLOCK_SECOND);

    if (node_id == 1) {
        NETSTACK_ROUTING.root_start();
    } else {
        LOG_ERR("The proxy border should be the node with id 1\n");
    }

    while(1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        // send_rpl_graph();
        etimer_reset(&timer);
    }
    PROCESS_END();
}
