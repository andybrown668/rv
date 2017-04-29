#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sched.h>
#include <unistd.h>

// Used to gain maximum performance from device during
// receiving bunch of data from sensors like DHTxx.
static int set_max_priority(void) {
    struct sched_param sched;
    memset(&sched, 0, sizeof(sched));
    // Use FIFO scheduler with highest priority
    // for the lowest chance of the kernel context switching.
    sched.sched_priority = sched_get_priority_max(SCHED_FIFO);
    if (-1 == sched_setscheduler(0, SCHED_FIFO, &sched)) {
        fprintf(stderr, "Unable to set SCHED_FIFO priority to the thread\n");
        return -1;
    }
    return 0;
}

// Get back normal thread priority.
static int set_default_priority(void) {
    struct sched_param sched;
    memset(&sched, 0, sizeof(sched));
    // Go back to regular schedule priority.
    sched.sched_priority = 0;
    if (-1 == sched_setscheduler(0, SCHED_OTHER, &sched)) {
        fprintf(stderr, "Unable to set SCHED_OTHER priority to the thread\n");
        return -1;
    }
    return 0;
}

int ready(int32_t fd) {
    fd_set rfds;
    struct timeval tv;
    int retval;

    /* Watch stdin (fd 0) to see when it has input. */
    FD_ZERO(&rfds);
    FD_SET(fd, &rfds);
    /* Wait up to 1 seconds. */
    tv.tv_sec = 1;
    tv.tv_usec = 0;
    retval = select(1, NULL, NULL, &rfds, &tv);
    /* Don’t rely on the value of tv now! */

    if (retval == -1)
        perror("select()");
    else if (retval)
        printf("Data is available now.\n");
        /* FD_ISSET(0, &rfds) will be true. */
    else
        printf("No data within five seconds.\n");
    return retval;
}

int readOne(int32_t fd) {
	char v;
//	lseek(fd, 0, 0);
    if (-1 == pread(fd, &v, sizeof(v), 0)) {
        return -1;
    } else {
    	return v;
    }
}

// read high/low state of fd, returning array of state/duration
int sample(int32_t fd, int states) {

}