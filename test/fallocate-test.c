#include <fcntl.h>
#include <stdio.h>

int main(void) {
    int fd = open("creserved", O_CREAT, 0664);
    if (fd) {
        if (fallocate(fd, 0664, 0, 1000000) == -1) {
            printf("fallocate returned error\n");
            return 1;
        }
    } else {
        printf("error opening file\n");
        return 1;
    }
    printf("fallocate successful\n");
    return 0;
}