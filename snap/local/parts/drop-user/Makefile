#
# apt-get install build-essential
# make clean
# make
# make install

CFLAGS += -g -O0 -Wall -Wstrict-prototypes

# snapcraft will copy anything from here
INSTALL_DIR := ${SNAPCRAFT_PART_INSTALL}

all: drop

drop:
	${CC} ${CFLAGS} ${LDFLAGS} drop.c -o $@ -ldl

install: drop
	mkdir -p ${INSTALL_DIR}
	cp -f drop ${INSTALL_DIR}/drop

clean:
	rm -f ./drop
