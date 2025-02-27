/*
 * Copyright (c) 2015, SICS Swedish ICT.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. Neither the name of the Institute nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE INSTITUTE AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE INSTITUTE OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 *
 *
 * Authors: Simon Duquennoy <simonduq@sics.se>
 */

#ifndef LINK_STATS_H_
#define LINK_STATS_H_

#include "net/linkaddr.h"
#include "os/net/nbr-table.h"

/* ETX fixed point divisor. 128 is the value used by RPL (RFC 6551 and RFC 6719) */
#ifdef LINK_STATS_CONF_ETX_DIVISOR
#define LINK_STATS_ETX_DIVISOR LINK_STATS_CONF_ETX_DIVISOR
#else /* LINK_STATS_CONF_ETX_DIVISOR */
#define LINK_STATS_ETX_DIVISOR                   128
#endif /* LINK_STATS_CONF_ETX_DIVISOR */

/* Option to infer the initial ETX from the RSSI of previously received packets. */
#ifdef LINK_STATS_CONF_INIT_ETX_FROM_RSSI
#define LINK_STATS_INIT_ETX_FROM_RSSI LINK_STATS_CONF_INIT_ETX_FROM_RSSI
#else /* LINK_STATS_CONF_INIT_ETX_FROM_RSSI */
#define LINK_STATS_INIT_ETX_FROM_RSSI              1
#endif /* LINK_STATS_CONF_INIT_ETX_FROM_RSSI */

/* Option to use packet and ACK count for ETX estimation, instead of EWMA */
#ifdef LINK_STATS_CONF_ETX_FROM_PACKET_COUNT
#define LINK_STATS_ETX_FROM_PACKET_COUNT LINK_STATS_CONF_ETX_FROM_PACKET_COUNT
#else /* LINK_STATS_CONF_ETX_FROM_PACKET_COUNT */
#define LINK_STATS_ETX_FROM_PACKET_COUNT           0
#endif /* LINK_STATS_ETX_FROM_PACKET_COUNT */

/* Store and periodically print packet counters? */
#ifdef LINK_STATS_CONF_PACKET_COUNTERS
#define LINK_STATS_PACKET_COUNTERS LINK_STATS_CONF_PACKET_COUNTERS
#else /* LINK_STATS_CONF_PACKET_COUNTERS */
#define LINK_STATS_PACKET_COUNTERS           0
#endif /* LINK_STATS_PACKET_COUNTERS */

NBR_TABLE_DECLARE(link_stats);

typedef uint16_t link_packet_stat_t;

struct link_packet_counter {
  /* total attempts to transmit unicast packets */
  link_packet_stat_t num_packets_tx;
  /* total ACKs for unicast packets */
  link_packet_stat_t num_packets_acked;
  /* total number of unicast and broadcast packets received */
  link_packet_stat_t num_packets_rx;
  /* total number of packets dropped before transmission due to insufficient memory */
  link_packet_stat_t num_queue_drops;
};


/* All statistics of a given link */
struct link_stats {
  clock_time_t last_tx_time;  /* Last Tx timestamp */
  uint16_t etx;               /* ETX using ETX_DIVISOR as fixed point divisor */
  int16_t rssi;               /* RSSI (received signal strength) */
  uint8_t freshness;          /* Freshness of the statistics */
#if LINK_STATS_ETX_FROM_PACKET_COUNT
  uint8_t tx_count;           /* Tx count, used for ETX calculation */
  uint8_t ack_count;          /* ACK count, used for ETX calculation */
#endif /* LINK_STATS_ETX_FROM_PACKET_COUNT */

#if LINK_STATS_PACKET_COUNTERS
  struct link_packet_counter cnt_current; /* packets in the current period */
  struct link_packet_counter cnt_total;   /* packets in total */
#endif
};

/* Returns the neighbor's link statistics */
const struct link_stats *link_stats_from_lladdr(const linkaddr_t *lladdr);
/* Returns the address of the neighbor */
const linkaddr_t *link_stats_get_lladdr(const struct link_stats *);
/* Are the statistics fresh? */
int link_stats_is_fresh(const struct link_stats *stats);
/* Resets link-stats module */
void link_stats_reset(void);
/* Initializes link-stats module */
void link_stats_init(void);
/* Packet sent callback. Updates statistics for transmissions on a given link */
void link_stats_packet_sent(const linkaddr_t *lladdr, int status, int numtx);
/* Packet input callback. Updates statistics for receptions on a given link */
void link_stats_input_callback(const linkaddr_t *lladdr);

#endif /* LINK_STATS_H_ */
