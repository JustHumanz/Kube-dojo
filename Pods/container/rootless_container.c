#define _GNU_SOURCE
#include <sched.h>
#include <unistd.h>
#include <stdlib.h>
#include <sys/wait.h>
#include <signal.h>
#include <stdio.h>
#include <sys/mount.h>
#include <sys/syscall.h>
#include <sys/stat.h>
#include <stdarg.h>
#include <limits.h>
#include <string.h>
#include <fcntl.h>

static char child_stack[1024 * 1024];

static void update_map(char *mapping, char *map_file) {
    int fd, j;
    size_t map_len;

    map_len = strlen(mapping);
    for (j = 0; j < map_len; j++)
        if (mapping[j] == ',')
            mapping[j] = '\n';

    fd = open(map_file, O_RDWR);
    if (fd == -1) {
        fprintf(stderr, "open %s", map_file);
        exit(EXIT_FAILURE);
    }

    if (write(fd, mapping, map_len) != map_len) {
        fprintf(stderr, "write %s", map_file);
        exit(EXIT_FAILURE);
    }

    close(fd);
}

int child_main(void *arg) {
    mount("proc", "./rootfs/proc", "proc", 0, NULL);
    mount("sys", "./rootfs/sys", "sysfs", 0, NULL);
    mount("none", "/tmp", "tmpfs", 0, NULL);
    mount("none", "/dev", "tmpfs", MS_NOSUID | MS_STRICTATIME, NULL);

    umount2("/proc", MNT_DETACH);
    mount("./rootfs", "./rootfs", "bind", MS_BIND | MS_REC, "");
    mkdir("./rootfs/oldrootfs", 0755);
    syscall(SYS_pivot_root, "./rootfs", "./rootfs/oldrootfs");
    chdir("/");
    umount2("/oldrootfs", MNT_DETACH);
    rmdir("/oldrootfs");

    char * hostname = "humanz_c";
    sethostname(hostname,sizeof(hostname));

    putenv("PS1=[Kirov_Command~>]");

    char **argv = (char **)arg;
    execvp(argv[0], &argv[0]);
}

int main(int argc, char *argv[]) {
    int flags;
    char map_path[PATH_MAX];
    char * uid_map = "0 1000 1";
    char * gid_map = "0 1000 1";

    flags = CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWIPC | CLONE_NEWPID  | CLONE_NEWNET | CLONE_NEWUSER;
    int pid = clone(child_main, child_stack + sizeof(child_stack),
                    flags | SIGCHLD, argv + 1);

    snprintf(map_path, PATH_MAX, "/proc/%ld/uid_map",
            (long) pid);
    update_map(uid_map, map_path);

    snprintf(map_path, PATH_MAX, "/proc/%ld/setgroups",
            (long) pid);
    update_map("deny", map_path);

    snprintf(map_path, PATH_MAX, "/proc/%ld/gid_map",
            (long) pid);
    update_map(gid_map, map_path);

    printf("%s: PID of child created by clone() is %ld\n",
            argv[0], (long) pid);

    waitpid(pid, NULL, 0);
    printf("%s: terminating\n", argv[0]);
    exit(EXIT_SUCCESS);

}
