//go:build linux

package osutil

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>

static int nullErrorHandler(Display *d, XErrorEvent *e) {
    return 0;
}

static void installSafeErrorHandler() {
    XSetErrorHandler(nullErrorHandler);
}
*/
import "C"

// InstallSafeX11ErrorHandler replaces the default X11 error handler with a
// no-op so that non-fatal X errors (like BadAccess from XGrabKey) don't
// kill the process.
func InstallSafeX11ErrorHandler() {
	C.installSafeErrorHandler()
}
