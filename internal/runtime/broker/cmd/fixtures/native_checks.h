#include <windows.h>
#include <stdlib.h>
#include <crtdbg.h>

static volatile LONG native_flags;
static volatile LONG native_tls_seen;

static void native_inspect(void) {
    WIN32_FIND_DATAW entry;
    HANDLE search;
    HANDLE heap = GetProcessHeap();
    void *memory;
    if (IsDebuggerPresent()) InterlockedOr(&native_flags, 1);
    search = FindFirstFileW(L".gh-tree-start-*", &entry);
    if (search != INVALID_HANDLE_VALUE) {
        InterlockedOr(&native_flags, 2);
        FindClose(search);
    }
    memory = HeapAlloc(heap, HEAP_ZERO_MEMORY, 257);
    if (!memory || !HeapValidate(heap, 0, memory)) InterlockedOr(&native_flags, 4);
    if (memory && !HeapFree(heap, 0, memory)) InterlockedOr(&native_flags, 8);
}

static void NTAPI native_tls(PVOID module, DWORD reason, PVOID reserved) {
    (void)module;
    (void)reserved;
    if (reason == DLL_PROCESS_ATTACH) {
        InterlockedExchange(&native_tls_seen, 1);
        native_inspect();
    }
}

#ifdef _WIN64
#pragma comment(linker, "/include:_tls_used")
#pragma comment(linker, "/include:fixture_tls_slot")
#else
#pragma comment(linker, "/include:__tls_used")
#pragma comment(linker, "/include:_fixture_tls_slot")
#endif
#pragma section(".CRT$XLB", long, read)
__declspec(allocate(".CRT$XLB")) PIMAGE_TLS_CALLBACK fixture_tls_slot = native_tls;

static int native_check_result(void) {
    void *memory;
    native_inspect();
    _CrtSetDbgFlag(_CRTDBG_ALLOC_MEM_DF | _CRTDBG_CHECK_ALWAYS_DF);
    memory = malloc(513);
    if (!memory) InterlockedOr(&native_flags, 16);
    free(memory);
    if (!_CrtCheckMemory()) InterlockedOr(&native_flags, 32);
    if (!native_tls_seen) InterlockedOr(&native_flags, 64);
    return (int)native_flags;
}
