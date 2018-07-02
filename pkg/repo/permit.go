package repo

import (
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Permit interface {
	ID() (*mytype.OID, error)
}

type AppleablePermit interface {
	AppledAt() (time.Time, error)
	ID() (*mytype.OID, error)
}

type EnrollablePermit interface {
	EnrolledAt() (time.Time, error)
	ID() (*mytype.OID, error)
}

type LabelablePermit interface {
	ID() (*mytype.OID, error)
	LabeledAt() (time.Time, error)
}

type TopicablePermit interface {
	ID() (*mytype.OID, error)
	TopicedAt() (time.Time, error)
}
