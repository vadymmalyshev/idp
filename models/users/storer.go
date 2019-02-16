package users

import (
	"context"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/pkg/errors"

	"github.com/sirupsen/logrus"
	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/otp/twofactor/sms2fa"
	"github.com/volatiletech/authboss/otp/twofactor/totp2fa"
)

var (
	db  = config.DB()
	log = config.Logger()

	assertUser   = &User{}
	assertStorer = &UserStorer{}

	_ authboss.User            = assertUser
	_ authboss.AuthableUser    = assertUser
	_ authboss.ConfirmableUser = assertUser
	_ authboss.LockableUser    = assertUser
	_ authboss.RecoverableUser = assertUser
	_ authboss.ArbitraryUser   = assertUser

	_ totp2fa.User = assertUser
	_ sms2fa.User  = assertUser

	_ authboss.CreatingServerStorer   = assertStorer
	_ authboss.ConfirmingServerStorer = assertStorer
	_ authboss.RecoveringServerStorer = assertStorer
)

type UserStorer struct {
}

func NewUserStorer() *UserStorer {
	return &UserStorer{}
}

// Load will look up the user based on the passed the PrimaryID
func (store UserStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	var user User

	notFoundByEmail := db.First(&user, "email = ?", key).RecordNotFound()

	if notFoundByEmail {
		return nil, errors.ErrUserNotFound
	}
	
	log.Info("user loaded by email", logrus.Fields{
		"email": user.Email,
	})

	return &user, nil
}

// Save persists the user in the database, this should never
// create a user and instead return ErrUserNotFound if the user
// does not exist.
func (store UserStorer) Save(ctx context.Context, user authboss.User) error {
	u := user.(*User)
	db.Save(&u)
	return nil
}

func (store UserStorer) New(ctx context.Context) authboss.User {
	return &User{}
}

func (store UserStorer) Create(ctx context.Context, user authboss.User) error {
	u := user.(*User)
	err := db.Create(u).Error

	if err != nil {
		return authboss.ErrUserFound
	}

	log.Infof("user created", logrus.Fields{
		"email": u.Email,
	})

	return nil
}

// LoadByConfirmSelector looks a user up by confirmation token
func (store UserStorer) LoadByConfirmSelector(ctx context.Context, selector string) (authboss.ConfirmableUser, error) {
	var user User

	err := db.Where(&User{ConfirmSelector: selector}).First(&user).Error
	return &user, err
}

// LoadByRecoverSelector looks a user up by confirmation selector
func (store UserStorer) LoadByRecoverSelector(ctx context.Context, selector string) (authboss.RecoverableUser, error) {
	var user User

	err := db.Where(&User{RecoverSelector: selector}).First(&user).Error
	return &user, err
}
