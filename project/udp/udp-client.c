#include "contiki.h"
#include "etimer.h"
#include "mutex.h"
#include "net/ipv6/simple-udp.h"
#include "net/mac/tsch/tsch.h"
#include "net/netstack.h"
#include "net/routing/routing.h"
#include "random.h"
#include "schedule_updater.h"
#include "sys/log.h"

#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define UDP_CLIENT_PORT 8765
#define UDP_SERVER_PORT 3000

#define SEND_INTERVAL (20 * CLOCK_SECOND)

static struct simple_udp_connection udp_conn;

#define SEQUENCE_NUMBER_MAX 0x7F
#define SEQUENCE_NUMBER_MASK 0x7F

// static uint16_t highest_ack = 0;
// static void udp_rx_callback(struct simple_udp_connection *c,
//                             const uip_ipaddr_t *sender_addr,
//                             uint16_t sender_port,
//                             const uip_ipaddr_t *receiver_addr,
//                             uint16_t receiver_port,
//                             const uint8_t *data,
//                             uint16_t datalen) {
//     if (datalen < 2) {
//         LOG_ERR("Datalen is < 2 and therefore not packet number could be read from the data\n");
//         return;
//     }
//     uint16_t pkt_number = ((uint16_t)data[1] << 8) + data[0];
//     LOG_INFO("Received response with pkt_number %u \n", pkt_number);
//     if (highest_ack > pkt_number) {
//         LOG_INFO("ACK already sent, passing\n");
//         return;
//     }
//     LOG_INFO("Sending ack %u\n", pkt_number);
//     static uint8_t send_buffer[10] = {0};
//     send_buffer[0] = data[0];
//     send_buffer[1] = data[1];
//     simple_udp_sendto(c, send_buffer, 10, sender_addr);

//     if (pkt_number <= highest_ack) {
//         LOG_INFO("Packet already process, only sending ACK\n");
//         return;
//     }
//     highest_ack = pkt_number;

//     const uint8_t *pkt = data + 2;
//     update_pkt_log(pkt);
// }

#define SEND_BUFFER_SIZE 256
static uint8_t send_buffer[SEND_BUFFER_SIZE] = {0};
static bool send_ready = false;
static uint16_t send_buffer_len;
static uint8_t ack_buffer[1] = {0};
static uint16_t ack_buffer_len;

/*---------------------------------------------------------------------------*/
PROCESS(udpack_process, "UDP Ack Process");
/*---------------------------------------------------------------------------*/

PROCESS_THREAD(udpack_process, ev, data) {
    static uint8_t sequence_number = 1;
    static struct etimer resend_interval;
    static uint8_t i = 0;
    PROCESS_BEGIN();
    LOG_INFO("UDP Ack Process started\n");
    while (true) {
        PROCESS_WAIT_EVENT_UNTIL(ev == PROCESS_EVENT_POLL);
        if (!send_ready) {
            // No packets was received just loop until we get a packet to send.
            continue;
        }
        LOG_INFO("Starting to send a new packet\n");
        if (sequence_number > SEQUENCE_NUMBER_MAX) {
            sequence_number = 0;
        }
        send_buffer[0] = sequence_number;
        send_buffer_len += 1;
        uint8_t packet_sequence_number = 0;
        // TODO: MAX SEQUENCE_NUMBER IS NOT CORRECTLY HANDLED.

        LOG_INFO("Sending %d bytes to the server\n", send_buffer_len);
        simple_udp_send(&udp_conn, &send_buffer, send_buffer_len);
        for (i = 0; i < 4; i++) {  // retries
            LOG_INFO("Value of i: %d\n", i);
            etimer_set(&resend_interval, 4 * CLOCK_SECOND);
            PROCESS_WAIT_EVENT_UNTIL(ev == PROCESS_EVENT_POLL || ev == PROCESS_EVENT_TIMER);
            if (ev == PROCESS_EVENT_TIMER) {
                LOG_INFO("Timer expired, retrying to send the packet\n", i);
                LOG_INFO("Sending %d bytes to the server\n", send_buffer_len);
                simple_udp_send(&udp_conn, &send_buffer, send_buffer_len);
                etimer_reset(&resend_interval);
                continue;
            }
            if (ev != PROCESS_EVENT_POLL) {
                LOG_ERR("Event should only be a PROCESS_EVENT_POLL\n");
                continue;
            }
            LOG_INFO("Processing received packet...\n");
            if (ack_buffer_len < 1) {
                LOG_ERR("Ack buffer is empty\n");
                break;
            }

            packet_sequence_number = ack_buffer[0] & SEQUENCE_NUMBER_MASK;
            LOG_INFO("Received ACK with sequence number: %d\n", packet_sequence_number);
            if (packet_sequence_number == sequence_number) {
                LOG_INFO("Correctly received the expected ACK\n");
                break;
            }
            if (packet_sequence_number > sequence_number && sequence_number != 0) {
                LOG_INFO("The sequence number received is greather than the expected sequence number, therefore we do nothing and wait for the precedents packages to arrive\n");
                continue;
            } else {
                LOG_INFO("Ack received that was not the expected ack, resending the packet\n");
                LOG_INFO("Sending %d bytes to the server\n", send_buffer_len);
                simple_udp_send(&udp_conn, &send_buffer, send_buffer_len);

            }
        }
        // Resetting buffer to null values
        memset(&send_buffer, 0, SEND_BUFFER_SIZE);
        send_ready = false;
        send_buffer_len = 0;
        memset(&ack_buffer, 0, 1);
        ack_buffer_len = 0;

        if (packet_sequence_number != sequence_number) {
            LOG_ERR("The packet sent was not ack after 4 retries, stopping...\n");
            // exit(256);
        } else {
            sequence_number++;
        }
    }
    PROCESS_END();
}

#define PACKET_TYPE_MASK 0x80
// refactor by specifying the packet type using an enum
static bool packet_is_ack(const uint8_t *packet, uint16_t packet_len) {
    if (packet_len < 1) {
        return false;
    }
    return (packet[0] & PACKET_TYPE_MASK) != 0;
}

static void ack_middleware(struct simple_udp_connection *c,
                           const uip_ipaddr_t *sender_addr,
                           uint16_t sender_port,
                           const uip_ipaddr_t *receiver_addr,
                           uint16_t receiver_port,
                           const uint8_t *data,
                           uint16_t datalen) {
    // Test if the received packet is an ACK or not
    bool is_ack = packet_is_ack(data, datalen);
    if (is_ack) {
        LOG_INFO("Received an ACK polling udpack_process\n");
        ack_buffer[0] = data[0];
        ack_buffer_len = 1;
        process_poll(&udpack_process);
        return;
    }

    // udp_rx_callback(c, sender_addr, sender_port, receiver_addr, receiver_port, data, datalen);
}

// send_server sends a packet to the server. The packet is created
// using the encoder function which must encode the packet to send into
// the buffer and return the number of bytes inserted into it.
static void send_server(uint16_t encoder(uint8_t *buffer)) {
    send_buffer_len = encoder(send_buffer + 1 /* Ack size */);
    send_ready = true;
    if (send_buffer[1] != 3) {
        LOG_ERR("Something wrong happend...\n");
    }
    process_poll(&udpack_process);
}

#define HELLO_WORLD_APPLICATION 3
static uint16_t hello_world_encoder(uint8_t *buffer) {
    static uint8_t i = 0;
    LOG_INFO("Encoding hello world: %d\n", i);
    buffer[0] = HELLO_WORLD_APPLICATION;
    buffer[1] = i++;
    return 2;
}

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;
// static struct tsch_slotframe *sf_next;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
    tsch_schedule_add_link(
        sf_common,
        (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
        LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
    for (int i = 1; i < APP_SLOTFRAME_SIZE; i++) {
        for (int j = 0; j < 4; j++) {
            struct tsch_link *link = tsch_schedule_add_link(
                sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
                &tsch_broadcast_address, i, j, 1);

            if (link == NULL) {
                LOG_ERR("Couldn't had the link with (timeslot, channel): (%d, %d)\n", i, j);
            }
        }
    }
}

/*---------------------------------------------------------------------------*/
PROCESS(udp_client_process, "UDP client");
AUTOSTART_PROCESSES(&udp_client_process, &udpack_process);
/*---------------------------------------------------------------------------*/
/*---------------------------------------------------------------------------*/
PROCESS_THREAD(udp_client_process, ev, data) {
    uip_ipaddr_t server_addr = {
        {0xfd, 0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}};

    static struct etimer hello_timer;
    PROCESS_BEGIN();
    initialize_tsch_schedule();
    /* Initialize UDP connection */
    simple_udp_register(&udp_conn, UDP_CLIENT_PORT, &server_addr, UDP_SERVER_PORT, ack_middleware);

    etimer_set(&hello_timer, 30 * CLOCK_SECOND);
    while (1) {
        PROCESS_YIELD();
        send_server(hello_world_encoder);
        etimer_reset(&hello_timer);
    }

    PROCESS_END();
}
/*---------------------------------------------------------------------------*/
