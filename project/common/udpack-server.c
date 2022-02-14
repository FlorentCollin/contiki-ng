#include "udpack-server.h"
#include "packet.h"
#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "etimer.h"
#include "sys/log.h"

#define LOG_MODULE "UDPAckServer"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SEND_BUFFER_SIZE 256
static uint8_t send_buffer[SEND_BUFFER_SIZE] = {0};
static int send_ready = 0;
static uint16_t send_buffer_len;
static uint8_t ack_sequence_number;

static struct simple_udp_connection udp_conn;

#define UDP_CLIENT_PORT 8765
#define UDP_SERVER_PORT 3000

/*---------------------------------------------------------------------------*/
PROCESS(udpack_process, "UDP Ack Process");
/*---------------------------------------------------------------------------*/

static void ack_middleware(struct simple_udp_connection *c,
                           const uip_ipaddr_t *sender_addr,
                           uint16_t sender_port,
                           const uip_ipaddr_t *receiver_addr,
                           uint16_t receiver_port,
                           const uint8_t *data,
                           uint16_t datalen) {
    // Test if the received packet is an ACK or not
    LOG_INFO("Received a packet from the server\n");
    Header header = 0;
    remove_header_from_packet(data, &header);
    enum PacketType packet_type = decode_packet_type(header);
    if (packet_type == PacketTypeAck) {
        LOG_INFO("Received an ACK polling udpack_process\n");
        ack_sequence_number = decode_sequence_number(header);
        process_poll(&udpack_process);
        return;
    }
    // TODO handle the data packet type...
    // udp_rx_callback(c, sender_addr, sender_port, receiver_addr, receiver_port, data, datalen);
}

PROCESS_THREAD(udpack_process, ev, data) {
    static uint8_t sequence_number = 1;
    static struct etimer resend_interval;
    static uint8_t i = 0;
    PROCESS_BEGIN();
    LOG_INFO("UDP Ack Process started\n");
    uip_ipaddr_t server_addr = {
        {0xfd, 0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}};

    simple_udp_register(&udp_conn, UDP_CLIENT_PORT, &server_addr, UDP_SERVER_PORT, ack_middleware);

    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(ev == PROCESS_EVENT_POLL);
        if (!send_ready) {
            // No packets was received just loop until we get a packet to send.
            continue;
        }
        LOG_INFO("Starting to send a new packet\n");
        if (sequence_number > SEQUENCE_NUMBER_MAX) {
            sequence_number = 0;
        }
        send_buffer_len += new_data_packet(send_buffer, sequence_number);
        LOG_INFO("Sending header: %d\n", send_buffer[0]);
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
            if (ack_sequence_number == sequence_number) {
                LOG_INFO("Correctly received the expected ACK\n");
                sequence_number++;
                break;
            }
            if (ack_sequence_number > sequence_number && sequence_number != 0) {
                LOG_INFO("The sequence number received is greather than the expected sequence number, therefore we do nothing and wait for the precedents packages to arrive\n");
                continue;
            } else {
                LOG_INFO("Ack received that was not the expected ack, resending the packet\n");
                LOG_INFO("Sending %d bytes to the server\n", send_buffer_len);
                simple_udp_send(&udp_conn, &send_buffer, send_buffer_len);

            }
        }

        if (ack_sequence_number != sequence_number - 1) {
            LOG_ERR("The packet sent was not ack after 4 retries, stopping...\n");
            // exit(256);
        }

        // Resetting buffer to null values
        memset(&send_buffer, 0, SEND_BUFFER_SIZE);
        send_ready = 0;
        send_buffer_len = 0;
        ack_sequence_number = 0;
    }
    PROCESS_END();
}

// send_server sends a packet to the server. The packet is created
// using the encoder function which must encode the packet to send into
// the buffer and return the number of bytes inserted into it.
void send_server(uint16_t encoder(uint8_t *buffer)) {
    static int first_call = 1;
    if (first_call != 0) {
        first_call = 0;
        process_start(&udpack_process, NULL);
    }
    send_buffer_len = encoder(send_buffer + sizeof(Header) /* Ack size */);
    send_ready = 1;
    process_poll(&udpack_process);
}