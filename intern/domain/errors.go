package domain

import "errors"

var ErrLoginAlreadyExists error = errors.New("login already exists")
var ErrUserNotFound error = errors.New("user not found")
var ErrInvalidPassword error = errors.New("invalid password")
var ErrFailedToValidatePhone error = errors.New("couldnt validate the phone number")
var ErrWeakPass error = errors.New("write a stronger pass")
var ErrSamePass error = errors.New("the new pass is the same to the old one")
var ErrNoExpiredCards error = errors.New("no expired cards")
var ErrNoActiveCards error = errors.New("no cards available")
var ErrUserAlreadyExists error = errors.New("User already exists")
