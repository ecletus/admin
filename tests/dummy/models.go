package dummy

import (
	"time"

	"github.com/moisespsena-go/aorm"
	"github.com/aghape/media/oss"
)

type CreditCard struct {
	aorm.Model
	Number string
	Issuer string
}

type Company struct {
	aorm.Model
	Name string
}

type Address struct {
	aorm.Model
	UserID   uint
	Address1 string
	Address2 string
}

type Language struct {
	aorm.Model
	Name string
}

type User struct {
	aorm.Model
	Name         string
	Age          uint
	Role         string
	Active       bool
	RegisteredAt *time.Time
	Avatar       oss.OSS
	Profile      Profile // has one
	CreditCardID uint
	CreditCard   CreditCard // belongs to
	Addresses    []Address  // has many
	CompanyID    uint
	Company      *Company   // belongs to
	Languages    []Language `gorm:"many2many:user_languages;"` // many 2 many
}

type Profile struct {
	aorm.Model
	UserID uint
	Name   string
	Sex    string

	Phone Phone
}

type Phone struct {
	aorm.Model

	ProfileID uint64
	Num       string
}
