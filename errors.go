package gproc

import "errors"

var ErrClosed = errors.New("gproc: closed service cant request")
var ErrNotFoundRequesterKey = errors.New("gproc: not found requester key")
var ErrNotFoundNoTargetForwardHandle = errors.New("gproc: not found no target forward handle")
