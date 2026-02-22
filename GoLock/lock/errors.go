package lock

import "errors"

var (
	ErrKeyEmpty     = errors.New("lock key is empty")
	ErrTTLInvalid   = errors.New("lock ttl must be positive")
	ErrNotAcquired  = errors.New("lock not acquired")
	ErrNotOwner     = errors.New("lock not owned by current holder")
	ErrLockerNil    = errors.New("locker is nil")
	ErrTokenMissing = errors.New("lock token is empty")
)
