package types

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

//Domain : Struct for a domain
type Domain struct {
	ID          int    `gorm:"primaryKey" example:"1"`
	OwnerID     int    `example:"2"`
	Fqdn        string `example:"example.org."`
	Description string `example:"My example website"`
	Serial      int    `example:"1"`
}

//GetDomain : get all domain infos from gorm database (by id)
func (d *Domain) GetDomain(db *gorm.DB) error {
	result := db.First(&d, d.ID)
	return result.Error
}

//GetOwner : get domain Owner_ID from gorm database (by id)
func (d *Domain) GetOwner(db *gorm.DB) error {
	result := db.Select("owner_id").First(&d, d.ID)
	return result.Error
}

//GetDomains : get all domains from gorm database (by id)
func GetDomains(db *gorm.DB, user User, count int, start int) ([]Domain, error) {
	domains := []Domain{}

	var rows *sql.Rows
	var err error

	if user.IsAdmin {
		rows, err = db.Limit(count).Offset(start).Model(&Domain{}).Rows()
	} else {
		rows, err = db.Limit(count).Offset(start).Where("owner_id = ?", user.ID).Model(&Domain{}).Rows()
	}
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var d Domain
		err = db.ScanRows(rows, &d)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

//GetDomainRecords : get all domains records in X domain from gorm database (by id)
func (d *Domain) GetDomainRecords(db *gorm.DB, count int, start int) ([]Record, error) {
	records := []Record{}

	searchInterface := Record{DomainID: d.ID}

	var rows *sql.Rows
	var err error

	rows, err = db.Limit(count).Offset(start).Where("domain_id = ?", searchInterface.DomainID).Model(&Record{}).Rows()
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var r Record
		err = db.ScanRows(rows, &r)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, nil
}

//CreateDomain : create domain in gorm database
func (d *Domain) CreateDomain(db *gorm.DB) error {
	result := db.Create(&d)
	return result.Error
}

//UpdateDomain : update domain from gorm database (by id)
func (d *Domain) UpdateDomain(db *gorm.DB) error {
	result := db.Save(&d)
	return result.Error
}

//DeleteDomain : delete domain from gorm database (by id)
func (d *Domain) DeleteDomain(db *gorm.DB) error {
	result := db.Delete(&d)
	return result.Error
}

//DeleteAllDomainRecords : delete all domain records from gorm database (by id)
func (d *Domain) DeleteAllDomainRecords(db *gorm.DB) error {
	result := db.Where("domain_id = ?", d.ID).Delete(Record{})
	return result.Error
}

//Exists : check if domain with the same FQDN already exists
func (d *Domain) Exists(db *gorm.DB) bool {
	result := db.Where("fqdn = ?", d.Fqdn).First(&d)
	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

//UpdateSOA : Generate SOA automaticly for domain
// Following Recommandations for DNS SOA Values
// https://www.ripe.net/publications/docs/ripe-203
func (d *Domain) UpdateSOA(db *gorm.DB, user User) {

	//Email to SOA email recommandations
	strings.ReplaceAll(user.Email, ".", "\\.")
	strings.ReplaceAll(user.Email, "@", ".")

	//Serial incrementation
	d.Serial = d.Serial + 1

	t := time.Now()
	serial := t.Format("20060102") //RFC 1912 2.2.
	soa := fmt.Sprintf("master.%s %s (%v%v 3600, 1800, 604800, 600)", d.Fqdn, user.Email, serial, d.Serial)
	record := Record{DomainID: d.ID, Fqdn: d.Fqdn, Type: 6, TTL: 3600, Content: soa}

	//Write new record
	db.Where("type = ? AND fqdn = ?", record.Type, record.Fqdn).Assign(record).FirstOrCreate(&record)

	//Write new serial
	d.UpdateDomain(db)
}
