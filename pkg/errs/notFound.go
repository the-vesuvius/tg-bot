package errs

type ErrNotFound struct {
	entity     string
	property   string
	propertyId string
}

func NewErrNotFound(entity, property, propertyId string) *ErrNotFound {
	return &ErrNotFound{
		entity:     entity,
		property:   property,
		propertyId: propertyId,
	}
}

func (e *ErrNotFound) Error() string {
	var msg string
	if e.property != "" && e.propertyId != "" {
		msg = e.entity + " with " + e.property + " " + e.propertyId + " not found"
	} else {
		msg = e.entity + " not found"
	}
	return msg
}

func (e *ErrNotFound) Is(target error) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}
