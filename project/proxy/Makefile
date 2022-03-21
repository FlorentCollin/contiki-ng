CONTIKI_PROJECT = client
all: $(CONTIKI_PROJECT)

PLATFORMS_ONLY = cooja

CONTIKI=../..

PROJECTDIRS += ../common
PROJECT_SOURCEFILES += schedule_updater.c udpack-server.c packet.c graph-application.c topology-application.c bandwidth-application.c

# force Security from command line
MAKE_WITH_SECURITY ?= 0

MAKE_MAC = MAKE_MAC_TSCH
MAKE_ROUTING = MAKE_ROUTING_RPL_LITE


include $(CONTIKI)/Makefile.dir-variables
MODULES += $(CONTIKI_NG_SERVICES_DIR)/rpl-border-router

ifeq ($(MAKE_WITH_SECURITY),1)
CFLAGS += -DWITH_SECURITY=1
endif

include $(CONTIKI)/Makefile.include
