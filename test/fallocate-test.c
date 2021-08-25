#define _GNU_SOURCE
#include <fcntl.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
// #include <linux/falloc.h>

int main(void) {
    int fd = open("creserved", O_CREAT | O_RDWR , 0664);
    if (fd < 0) {
        printf("error opening file\n");
        return 1;
    } else {
        if (fallocate(fd, 0, 0, 100000000) == -1) {
            printf("fallocate returned error:\n%s\n", strerror(errno));
            return 1;
        }
    }
    printf("fallocate successful\n");
    return 0;
}