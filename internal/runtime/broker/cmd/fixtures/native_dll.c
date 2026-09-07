#include "native_checks.h"

static volatile LONG native_attached;

BOOL WINAPI DllMain(HINSTANCE module, DWORD reason, LPVOID reserved) {
    (void)module;
    (void)reserved;
    if (reason == DLL_PROCESS_ATTACH) {
        InterlockedExchange(&native_attached, 1);
        native_inspect();
    }
    return TRUE;
}

__declspec(dllexport) int fixture_dll_result(void) {
    int result = native_check_result();
    if (!native_attached) result |= 128;
    return result;
}
