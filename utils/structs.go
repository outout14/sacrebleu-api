package utils

import "github.com/outout14/sacrebleu-api/api"

//Database : Struct for SQL Database configuration in the config.ini file
type Database struct {
	IP       string
	Port     string
	Username string
	Password string
	Db       string
	Type     string
}

//Conf : Struct for the whole config.ini file when it will be parsed by go-ini
type Conf struct {
	AppMode string
	App     api.App
	Database
}

//Record : Struct for a domain record
//Defined by it's ID, DomainID (parent domain), Fqdn (or name), Content (value of the record), Type (as Qtype/int), TTL (used only for the DNS response and not the Redis TTL)
type Record struct {
	ID       uint `gorm:"primaryKey"`
	DomainID int
	Fqdn     string
	Content  string
	Type     int
	Qtype    uint16 `gorm:"-"`
	TTL      int
}
