package common

import (
	"time"
)

const EventCacheSize = 10

const ConfirmationsCount uint64 = 1
const ClusterTaskTimeout = 3 * time.Second

const DArcherIP = "127.0.0.1"
const DArcherPort = 1236

type LifecycleState string

const (
	CREATED  LifecycleState = "created"
	PENDING  LifecycleState = "pending"
	EXECUTED LifecycleState = "executed"
	// deprecated
	REVERTED LifecycleState = "reverted"
	// deprecated
	REEXECUTED LifecycleState = "re-executed"
	CONFIRMED  LifecycleState = "confirmed"
	DROPPED    LifecycleState = "dropped"
)
