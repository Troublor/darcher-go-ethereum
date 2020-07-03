package common

import (
	"time"
)

const ConfirmationsCount uint64 = 1
const ClusterTaskTimeout = 3 * time.Second

const ServerPort = 1234     // server port for rpc with geth
const ServerPortJson = 1235 // server port for rpc with dArcher

const DArcherIP = "127.0.0.1"
const DArcherPort = 1236

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

const (
	ModeDirect   = "direct"
	ModeTraverse = "traverse"
	ModeTrivial  = "trivial"
)

const (
	DOER   Role = "DOER"
	TALKER Role = "TALKER"
)
