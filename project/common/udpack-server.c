#include "udpack-server.h"
#include "packet.h"
#include "net/ipv6/simple-udp.h"
#include "net/ipv6/uip.h"
#include "etimer.h"
#include "sys/log.h"
#include "schedule_updater.h"

#define LOG_MODULE "UDPAckServer"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SEND_BUFFER_SIZE 512
static uint8_t send_buffer[SEND_BUFFER_SIZE] = {0};
static int send_ready = 0;
static uint16_t send_buffer_len;
static uint8_t ack_sequence_number;

static struct simple_udp_connection udp_conn;

#define UDP_CLIENT_PORT 8765
#ifndef UDP_SERVER_PORT
#define UDP_SERVER_PORT 3000
#endif

#ifndef UDP_SERVER_ADDR
#define UDP_SERVER_ADDR {{0xfd, 0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}}
#endif

/*---------------------------------------------------------------------------*/
PROCESS(udpack_process, "UDP Ack Process");
/*---------------------------------------------------------------------------*/

static uint16_t highest_ack = 0;
static void udp_rx_callback(struct simple_udp_connection *c,
                            const uip_ipaddr_t *sender_addr,
                            uint16_t sender_port,
                            const uip_ipaddr_t *receiver_addr,
                            uint16_t receiver_port,
                            const uint8_t *data,
                            uint16_t datalen) {
    if (datalen < 2) {
        LOG_ERR("Datalen is < 2 and therefore no packet number could be read from the data\n");
        return;
    }
    Header header;
    remove_header_from_packet(data, &header);
    uint8_t sequence_number = decode_sequence_number(header);
    LOG_INFO("Received response with sequence_number %u \n", sequence_number);
    if (highest_ack > sequence_number) {
        LOG_INFO("ACK already sent, passing\n");
        return;
    }
    LOG_INFO("Sending ack %u\n", sequence_number);
    static uint8_t send_buffer[10] = {0};

    uint16_t len = new_ack_packet(send_buffer, sequence_number);
    // encode the confirmation message for now we accept every new schedule sent to us
#define CONFIRMATION 1
    send_buffer[len++] = CONFIRMATION;
    simple_udp_sendto(c, send_buffer, 10, sender_addr);

    if (sequence_number <= highest_ack) {
        LOG_INFO("Packet already process, only sending ACK\n");
        return;
    }
    highest_ack = sequence_number;

    const uint8_t *pkt = data + 1;
    update_pkt_dispatch(pkt);
}

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
    LOG_INFO("Call udp_rx_callback\n");
    udp_rx_callback(c, sender_addr, sender_port, receiver_addr, receiver_port, data, datalen);
}

typedef uint16_t encoder_fn(uint8_t* packet_buffer);

PROCESS_THREAD(udpack_process, ev, encoder) {
    static uint8_t sequence_number = 1;
    static struct etimer resend_interval;
    static uint8_t i = 0;
    PROCESS_BEGIN();
    LOG_INFO("UDP Ack Process started\n");
    uip_ipaddr_t server_addr = UDP_SERVER_ADDR;
    LOG_ERR_6ADDR(&server_addr);

    simple_udp_register(&udp_conn, UDP_CLIENT_PORT, &server_addr, UDP_SERVER_PORT, ack_middleware);

    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(ev == PROCESS_EVENT_CONTINUE);
        send_buffer_len = ((encoder_fn*) encoder)(send_buffer + sizeof(Header));

        if (sequence_number > SEQUENCE_NUMBER_MAX) {
            sequence_number = 0;
        }
        send_buffer_len += new_data_packet(send_buffer, sequence_number);
        LOG_INFO("Sending %d bytes to the server\n", send_buffer_len);
        simple_udp_send(&udp_conn, &send_buffer, send_buffer_len);
        for (i = 0; i < 4; i++) {  // retries
            etimer_set(&resend_interval, 10 * CLOCK_SECOND);
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
                LOG_INFO("Resending %d bytes to the server\n", send_buffer_len);
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


void init_udpack_server() {
    static int first_call = 1;
    if (first_call != 0) {
        first_call = 0;
        process_start(&udpack_process, NULL);
    }
}

void send_server_ack(uint16_t encoder(uint8_t *buffer)) {
    init_udpack_server();
    process_post(&udpack_process, PROCESS_EVENT_CONTINUE, encoder);
}

// send_server sends a packet to the server without waiting for an ACK.
// The packet is created // using the encoder function which must encode 
// the packet to send into the buffer and return the number of bytes 
// inserted into it.
void send_server(uint16_t encoder(uint8_t *buffer)) {
    init_udpack_server();
    Header header = new_header(PacketTypeDataNoACK);
    send_buffer[0] = header;
    uint16_t len = sizeof(header);
    len += encoder(send_buffer + len);
    simple_udp_send(&udp_conn, &send_buffer, len);
}