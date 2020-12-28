package types

import (
	"errors"

	"gorm.io/gorm"
)

//Record : Struct for a domain record
//Defined by it's ID, DomainID (parent domain), Fqdn (or name), Content (value of the record), Type (as Qtype/int), TTL
type Record struct {
	ID       int    `gorm:"primaryKey"`
	DomainID int    `example:"1"`
	Fqdn     string `example:"sub.example.org."`
	Content  string `example:"192.0.2.3"`
	Type     int    `example:"1"`
	Qtype    uint16 `gorm:"-"` //Not saved in the database
	TTL      int    `example:"3600"`
}

//GetRecord : get record from gorm database (by id)
func (r *Record) GetRecord(db *gorm.DB) error {
	result := db.First(&r, r.ID)
	return result.Error
}

//CreateRecord : create record in gorm database
func (r *Record) CreateRecord(db *gorm.DB) error {
	result := db.Create(&r)
	return result.Error
}

//UpdateRecord : update record from gorm database (by id)
func (r *Record) UpdateRecord(db *gorm.DB) error {
	result := db.Save(&r)
	return result.Error
}

//DeleteRecord : delete record from gorm database (by id)
func (r *Record) DeleteRecord(db *gorm.DB) error {
	result := db.Delete(&r)
	return result.Error
}

//Exists : check if record with the same FQDN already exists
func (r *Record) Exists(db *gorm.DB) bool {
	result := db.Where("fqdn = ?", r.Fqdn).First(&r)
	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}
