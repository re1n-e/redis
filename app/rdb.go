package main

import (
	"encoding/binary"
	"fmt"
)

type Header struct {
	Magic   [5]byte // "REDIS"
	Version [4]byte // e.g., "0011"
}

type Metadata struct {
	Start byte   // Always 0xFA
	Name  []byte // String-encoded metadata name
	Value []byte // String-encoded metadata value
}

type ExpireInfo struct {
	Type byte   // FC or FD
	Time []byte // 8 bytes if FC, 4 bytes if FD
}

type Entry struct {
	Expire *ExpireInfo // nil if no expiration
	Type   byte        // 0x00 = string
	Key    []byte      // string-encoded
	Value  []byte      // string-encoded
}

type Database struct {
	Start         byte           // 0xFE marker
	Index         int            // Parsed from size encoding
	HashTableInfo *HashTableInfo // nil if no 0xFB marker
	Entries       []Entry
}

type HashTableInfo struct {
	Marker        byte // 0xFB
	KeyValueSize  int  // Hash table size
	ExpireKeySize int  // Expire hash table size
}

type RDB struct {
	Header   Header
	Metadata []Metadata
	DBs      []Database
	Checksum [8]byte
}

func ParseRDB(data []byte) (*RDB, error) {
	rdb := &RDB{}
	offset := 0

	if len(data) < 9 {
		return nil, &ParseError{offset, "file too short for header"}
	}

	// Parse Header
	copy(rdb.Header.Magic[:], data[0:5])
	copy(rdb.Header.Version[:], data[5:9])
	offset = 9

	// Parse Metadata sections
	for offset < len(data) && data[offset] == 0xFA {
		metadata := Metadata{Start: data[offset]}
		offset++

		// Parse metadata name
		name, bytesRead, err := decodeString(data, offset)
		if err != nil {
			return nil, &ParseError{offset, fmt.Sprintf("error parsing metadata name: %v", err)}
		}
		metadata.Name = []byte(name)
		offset += bytesRead

		// Parse metadata value
		value, bytesRead, err := decodeString(data, offset)
		if err != nil {
			return nil, &ParseError{offset, fmt.Sprintf("error parsing metadata value: %v", err)}
		}
		metadata.Value = []byte(value)
		offset += bytesRead

		rdb.Metadata = append(rdb.Metadata, metadata)
	}

	// Parse Database sections
	for offset < len(data) && data[offset] == 0xFE {
		db := Database{Start: data[offset]}
		offset++

		// Parse database index
		dbIndex, bytesRead, err := decodeSizeEncoding(data, offset)
		if err != nil {
			return nil, &ParseError{offset, fmt.Sprintf("error parsing database index: %v", err)}
		}
		db.Index = dbIndex
		offset += bytesRead

		// Check for hash table info
		if offset < len(data) && data[offset] == 0xFB {
			hashInfo := &HashTableInfo{Marker: data[offset]}
			offset++

			// Parse key-value hash table size
			kvSize, bytesRead, err := decodeSizeEncoding(data, offset)
			if err != nil {
				return nil, &ParseError{offset, fmt.Sprintf("error parsing key-value size: %v", err)}
			}
			hashInfo.KeyValueSize = kvSize
			offset += bytesRead

			// Parse expire hash table size
			expireSize, bytesRead, err := decodeSizeEncoding(data, offset)
			if err != nil {
				return nil, &ParseError{offset, fmt.Sprintf("error parsing expire size: %v", err)}
			}
			hashInfo.ExpireKeySize = expireSize
			offset += bytesRead

			db.HashTableInfo = hashInfo
		}

		// Parse entries
		for offset < len(data) && data[offset] != 0xFE && data[offset] != 0xFF {
			entry := Entry{}

			// Check for expiration info
			if data[offset] == 0xFC || data[offset] == 0xFD {
				expireInfo := &ExpireInfo{Type: data[offset]}
				offset++

				timeSize := 8 // FC = 8 bytes
				if expireInfo.Type == 0xFD {
					timeSize = 4 // FD = 4 bytes
				}

				if offset+timeSize > len(data) {
					return nil, &ParseError{offset, "insufficient data for expire timestamp"}
				}

				expireInfo.Time = make([]byte, timeSize)
				copy(expireInfo.Time, data[offset:offset+timeSize])
				offset += timeSize

				entry.Expire = expireInfo
			}

			// Parse value type
			if offset >= len(data) {
				return nil, &ParseError{offset, "unexpected end of data when parsing value type"}
			}
			entry.Type = data[offset]
			offset++

			// Parse key
			key, bytesRead, err := decodeString(data, offset)
			if err != nil {
				return nil, &ParseError{offset, fmt.Sprintf("error parsing key: %v", err)}
			}
			entry.Key = []byte(key)
			offset += bytesRead

			// Parse value
			value, bytesRead, err := decodeString(data, offset)
			if err != nil {
				return nil, &ParseError{offset, fmt.Sprintf("error parsing value: %v", err)}
			}
			entry.Value = []byte(value)
			offset += bytesRead

			db.Entries = append(db.Entries, entry)
		}

		rdb.DBs = append(rdb.DBs, db)
	}

	// Parse end of file marker and checksum
	if offset < len(data) && data[offset] == 0xFF {
		offset++ // Skip 0xFF marker
		if offset+8 <= len(data) {
			copy(rdb.Checksum[:], data[offset:offset+8])
		}
	}

	return rdb, nil
}

func decodeSizeEncoding(data []byte, offset int) (int, int, error) {
	if offset >= len(data) {
		return 0, 0, fmt.Errorf("insufficient data for size encoding")
	}

	first := data[offset]
	switch first >> 6 { // First 2 bits
	case 0b00: // 6-bit size
		return int(first & 0x3F), 1, nil
	case 0b01: // 14-bit size
		if len(data) < offset+2 {
			return 0, 0, fmt.Errorf("insufficient data for 14-bit size")
		}
		size := int(first&0x3F)<<8 | int(data[offset+1])
		return size, 2, nil
	case 0b10: // 32-bit size
		if len(data) < offset+5 {
			return 0, 0, fmt.Errorf("insufficient data for 32-bit size")
		}
		size := binary.BigEndian.Uint32(data[offset+1 : offset+5])
		return int(size), 5, nil
	case 0b11: // Special string encoding
		return int(first & 0x3F), 1, nil // Return the encoding type
	}
	return 0, 0, fmt.Errorf("invalid size encoding")
}

func decodeString(data []byte, offset int) (string, int, error) {
	sizeVal, bytesRead, err := decodeSizeEncoding(data, offset)
	if err != nil {
		return "", 0, err
	}

	first := data[offset]
	if first>>6 == 0b11 { // Special encoding
		return decodeSpecialString(data, offset, sizeVal)
	}

	// Regular string
	if len(data) < offset+bytesRead+sizeVal {
		return "", 0, fmt.Errorf("insufficient data for string")
	}

	str := string(data[offset+bytesRead : offset+bytesRead+sizeVal])
	return str, bytesRead + sizeVal, nil
}

func decodeSpecialString(data []byte, offset int, encodingType int) (string, int, error) {
	switch encodingType {
	case 0: // 8-bit integer (0xC0)
		if len(data) < offset+2 {
			return "", 0, fmt.Errorf("insufficient data for 8-bit integer")
		}
		val := int8(data[offset+1])
		return fmt.Sprintf("%d", val), 2, nil
	case 1: // 16-bit integer (0xC1)
		if len(data) < offset+3 {
			return "", 0, fmt.Errorf("insufficient data for 16-bit integer")
		}
		val := binary.LittleEndian.Uint16(data[offset+1 : offset+3])
		return fmt.Sprintf("%d", val), 3, nil
	case 2: // 32-bit integer (0xC2)
		if len(data) < offset+5 {
			return "", 0, fmt.Errorf("insufficient data for 32-bit integer")
		}
		val := binary.LittleEndian.Uint32(data[offset+1 : offset+5])
		return fmt.Sprintf("%d", val), 5, nil
	case 3: // LZF compressed string (0xC3)
		return "", 0, fmt.Errorf("LZF compressed strings not supported")
	default:
		return "", 0, fmt.Errorf("unknown special string encoding: %d", encodingType)
	}
}

type ParseError struct {
	Offset int
	Msg    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at offset %d: %s", e.Offset, e.Msg)
}
