package errs

import "errors"

var ErrTaskHandlerNotRegistered = errors.New("handler has not been registered")
