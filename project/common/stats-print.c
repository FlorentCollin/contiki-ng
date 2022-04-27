#include "stats-print.h"

#include "contiki.h"
#include "os/net/link-stats.h"
#include "sys/log.h"
#include "ctimer.h"

#define LOG_MODULE "PrintStats"
#define LOG_LEVEL LOG_LEVEL_INFO

static struct ctimer periodic_timer;

static void print_link_stats() {
    struct link_stats *stats;

    for (stats = nbr_table_head(link_stats); stats != NULL; stats = nbr_table_next(link_stats, stats)) {
        LOG_INFO("{\"tx\": %u, \"ack\": %u, \"rx\": %u, \"neighbor\": \"", 
            stats->cnt_total.num_packets_tx + stats->cnt_current.num_packets_tx,
            stats->cnt_total.num_packets_acked + stats->cnt_current.num_packets_acked,
            stats->cnt_total.num_packets_rx + stats->cnt_current.num_packets_rx);
        LOG_INFO_LLADDR(link_stats_get_lladdr(stats));
        printf("\"}\n");
    }

    ctimer_reset(&periodic_timer);
}

void start_print_link_stats() {
    ctimer_set(&periodic_timer, 29 * CLOCK_SECOND, print_link_stats, NULL);
}