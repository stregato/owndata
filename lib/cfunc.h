#ifndef _CFUNC_H
#define _CFUNC_H

typedef struct Data {
    void* ptr;
    size_t len;
} Data;

typedef struct Result{
    void* ptr;
    size_t len;
    unsigned long long hnd;
	char* err;
} Result;


#endif