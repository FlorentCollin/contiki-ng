CONTIKI_PROJECT = udp-client
all: $(CONTIKI_PROJECT)

ifdef UDP_SERVER_PORT
CFLAGS += -DUDP_SERVER_PORT=$(UDP_SERVER_PORT)
endif

ifdef UDP_SERVER_ADDR
CFLAGS += -DUDP_SERVER_ADDR=$(UDP_SERVER_ADDR)
endif

PROJECTDIRS += ../common
PROJECT_SOURCEFILES += schedule_updater.c udpack-server.c packet.c topology-application.c bandwidth-application.c graph-application.c stats-print.c

PLATFORMS_ONLY = cooja

CONTIKI=../..

# force Security from command line
MAKE_WITH_SECURITY ?= 0

MAKE_MAC = MAKE_MAC_TSCH
MAKE_ROUTING = MAKE_ROUTING_RPL_CLASSIC

# MODULES += $(CONTIKI_NG_SERVICES_DIR)/orchestra

ifeq ($(MAKE_WITH_SECURITY),1)
CFLAGS += -DWITH_SECURITY=1
endif

include $(CONTIKI)/Makefile.include
