package repo

import (
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type NodePermit interface {
	ID() (*mytype.OID, error)
}

type AppleablePermit interface {
	AppledAt() time.Time
	ID() (*mytype.OID, error)
}

type EnrollablePermit interface {
	EnrolledAt() time.Time
	ID() (*mytype.OID, error)
}

type LabelablePermit interface {
	ID() (*mytype.OID, error)
	LabeledAt() time.Time
}

type TopicablePermit interface {
	ID() (*mytype.OID, error)
	TopicedAt() time.Time
}
