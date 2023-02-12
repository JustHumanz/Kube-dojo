#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>

#define BLOCK_SIZE 1024*1024*265 //256Mb

int main(){
    char * str_buf;
    
    str_buf = malloc(BLOCK_SIZE);
    if (str_buf != NULL){
        memset(str_buf,atoi("Kano"),BLOCK_SIZE);
    }

    pause();
    return 0;
}