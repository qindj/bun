package pgdialect

import (
	"database/sql"
	"encoding/json"
	"net"
	"reflect"

	"github.com/uptrace/bun/dialect/sqltype"
	"github.com/uptrace/bun/schema"
)

const (
	// Date / Time
	pgTypeTimestampTz = "TIMESTAMPTZ"         // Timestamp with a time zone
	pgTypeDate        = "DATE"                // Date
	pgTypeTime        = "TIME"                // Time without a time zone
	pgTypeTimeTz      = "TIME WITH TIME ZONE" // Time with a time zone
	pgTypeInterval    = "INTERVAL"            // Time Interval

	// Network Addresses
	pgTypeInet    = "INET"    // IPv4 or IPv6 hosts and networks
	pgTypeCidr    = "CIDR"    // IPv4 or IPv6 networks
	pgTypeMacaddr = "MACADDR" // MAC addresses

	// Serial Types
	pgTypeSmallSerial = "SMALLSERIAL" // 2 byte autoincrementing integer
	pgTypeSerial      = "SERIAL"      // 4 byte autoincrementing integer
	pgTypeBigSerial   = "BIGSERIAL"   // 8 byte autoincrementing integer

	// Binary Data Types
	pgTypeBytea = "BYTEA" // binary string
)

var (
	ipType             = reflect.TypeOf((*net.IP)(nil)).Elem()
	ipNetType          = reflect.TypeOf((*net.IPNet)(nil)).Elem()
	jsonRawMessageType = reflect.TypeOf((*json.RawMessage)(nil)).Elem()
	nullStringType     = reflect.TypeOf((*sql.NullString)(nil)).Elem()
)

func (d *Dialect) DefaultVarcharLen() int {
	return 0
}

func fieldSQLType(field *schema.Field) string {
	if field.UserSQLType != "" {
		return field.UserSQLType
	}

	if v, ok := field.Tag.Option("composite"); ok {
		return v
	}
	if field.Tag.HasOption("hstore") {
		return sqltype.HSTORE
	}

	if field.Tag.HasOption("array") {
		switch field.IndirectType.Kind() {
		case reflect.Slice, reflect.Array:
			sqlType := sqlType(field.IndirectType.Elem())
			return sqlType + "[]"
		}
	}

	if field.DiscoveredSQLType == sqltype.Blob {
		return pgTypeBytea
	}

	return sqlType(field.IndirectType)
}

func sqlType(typ reflect.Type) string {
	switch typ {
	case nullStringType: // typ.Kind() == reflect.Struct, test for exact match
		return sqltype.VarChar
	case ipType:
		return pgTypeInet
	case ipNetType:
		return pgTypeCidr
	case jsonRawMessageType:
		return sqltype.JSONB
	}

	sqlType := schema.DiscoverSQLType(typ)
	switch sqlType {
	case sqltype.Timestamp:
		sqlType = pgTypeTimestampTz
	}

	switch typ.Kind() {
	case reflect.Map, reflect.Struct: // except typ == nullStringType, see above
		if sqlType == sqltype.VarChar {
			return sqltype.JSONB
		}
		return sqlType
	case reflect.Array, reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return pgTypeBytea
		}
		return sqltype.JSONB
	}

	return sqlType
}
