package db

type InternalError struct {
	error
}

func NewInternalError(err error) *InternalError {
	return &InternalError{error: err}
}

func (ie *InternalError) Error() string {
	return ie.error.Error()
}

type UserError struct {
	error
}

func NewUserError(err error) *UserError {
	return &UserError{error: err}
}

func (ue *UserError) Error() string {
	return ue.error.Error()
}
