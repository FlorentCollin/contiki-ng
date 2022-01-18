CONTIKI_PROJECT = client
all: $(CONTIKI_PROJECT)

PROJECT_SOURCEFILES += schedule_updater.c

PLATFORMS_ONLY = cooja

CONTIKI=../..

# force Security from command line
MAKE_WITH_SECURITY ?= 0

MAKE_MAC = MAKE_MAC_TSCH
MAKE_ROUTING = MAKE_ROUTING_RPL_LITE

ifeq ($(MAKE_WITH_SECURITY),1)
CFLAGS += -DWITH_SECURITY=1
endif

include $(CONTIKI)/Makefile.include
