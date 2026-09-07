#include "native_checks.h"
#include <stdio.h>

__declspec(dllimport) int fixture_dll_result(void);

int main(void) {
    int executable = native_check_result();
    int library = fixture_dll_result();
    FILE *marker = NULL;
    if (fopen_s(&marker, "native-loader-ran", "wb") != 0) return 129;
    fputs("after DLL/TLS startup", marker);
    fclose(marker);
    printf("NATIVE_DLL_TLS_DEBUG_HEAP exe=%d dll=%d\n", executable, library);
    return executable || library;
}
