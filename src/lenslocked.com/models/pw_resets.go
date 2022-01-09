package models

import (
	"github.com/eitah/lenslocked/src/lenslocked.com/hash"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/jinzhu/gorm"
)

// intentionally unexported so models can use it.
type pwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique undex"`
}

type pwResetService interface {
	pwResetDB
}

type pwResetValidator struct {
	pwResetDB
	hmac hash.HMAC
}

type pwResetDB interface {
	ByToken(token string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id uint) error
}

type pwResetGorm struct {
	db *gorm.DB
}

// func NewPwResetService(db *gorm.DB, hmacSecretKey string) *pwResetService {
// 	pwg := &pwResetGorm{
// 		db: db,
// 	}
// 	hmac := hash.NewHMAC(hmacSecretKey)
// 	pwv := NewPwResetValidator(pwg, hmac)
// 	return &pwResetService{
// 		pwResetDB: pwv,
// 	}
// }

func NewPwResetValidator(db pwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		pwResetDB: db,
		hmac:      hmac,
	}
}

func (pwrv *pwResetValidator) hmacToken(pwr *pwReset) error {
	if pwr.Token == "" {
		return nil
	}
	pwr.TokenHash = pwrv.hmac.Hash(pwr.Token)
	return nil
}

func (pwrv *pwResetValidator) setTokenIfUnset(pwr *pwReset) error {
	if pwr.Token != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	pwr.Token = token
	return nil
}

func (pwrv *pwResetValidator) ByToken(token string) (*pwReset, error) {
	pwr := pwReset{
		Token: token,
	}
	if err := runTokenValFns(&pwr, pwrv.hmacToken); err != nil {
		return nil, err
	}

	return pwrv.pwResetDB.ByToken(pwr.TokenHash)
}

func (pwrg *pwResetGorm) ByToken(tokenHash string) (*pwReset, error) {
	var pwr pwReset
	err := pwrg.db.First(pwrg.db.Where("token_hash = ?", tokenHash), &pwr).Error
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

func (pwrv *pwResetValidator) requireUserID(pwr *pwReset) error {
	if pwr.UserID == 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (pwrv *pwResetValidator) Create(pwr *pwReset) error {
	if err := runTokenValFns(pwr,
		pwrv.requireUserID,
		pwrv.setTokenIfUnset,
		pwrv.hmacToken); err != nil {
		return err
	}

	return pwrv.pwResetDB.Create(pwr)
}

func (pwrg *pwResetGorm) Create(pwr *pwReset) error {
	return pwrg.db.Save(pwr).Error
}

func (pwrv *pwResetValidator) Delete(id uint) error {
	pwr := pwReset{Model: gorm.Model{ID: id}}
	if err := runTokenValFns(&pwr, pwrv.requireUserID); err != nil {
		return err
	}

	return pwrv.pwResetDB.Create(&pwr)
}

func (pwrg *pwResetGorm) Delete(id uint) error {
	pwr := pwReset{Model: gorm.Model{ID: id}}
	return pwrg.db.Delete(pwr).Error
}

type pwResetFn func(*pwReset) error

func runTokenValFns(pwr *pwReset, fns ...pwResetFn) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}
