// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.etcd.io/bbolt"
)

type databaseTableDescriptor struct {
	key        []byte
	recordType reflect.Type
}

// ValidateRestoreFile verifies that a file contains every registered table and decodable records.
func (database *Database) ValidateRestoreFile(filename string) error {
	source, err := bbolt.Open(filename, 0444, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return err
	}
	defer source.Close()
	return source.View(func(tx *bbolt.Tx) error { return validateBackupTables(tx, database.tableDescriptors()) })
}

// RestoreFrom validates a Bolt backup and atomically replaces every registered table.
func (database *Database) RestoreFrom(filename string) error {
	source, err := bbolt.Open(filename, 0444, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return err
	}
	defer source.Close()
	tables := database.tableDescriptors()
	if err = source.View(func(tx *bbolt.Tx) error { return validateBackupTables(tx, tables) }); err != nil {
		return err
	}
	return source.View(func(sourceTx *bbolt.Tx) error {
		return database.bolt.Update(func(destinationTx *bbolt.Tx) error {
			for _, table := range tables {
				sourceBucket := sourceTx.Bucket(table.key)
				if err := destinationTx.DeleteBucket(table.key); err != nil {
					return err
				}
				destinationBucket, err := destinationTx.CreateBucket(table.key)
				if err != nil {
					return err
				}
				if err = destinationBucket.SetSequence(sourceBucket.Sequence()); err != nil {
					return err
				}
				if err = sourceBucket.ForEach(func(key, value []byte) error {
					if value == nil {
						return fmt.Errorf("nested bucket in table %s", table.key)
					}
					return destinationBucket.Put(key, value)
				}); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (database *Database) tableDescriptors() []databaseTableDescriptor {
	return []databaseTableDescriptor{
		{database.allianceTable.bucketKey, database.allianceTable.recordType},
		{database.awardTable.bucketKey, database.awardTable.recordType},
		{database.eventSettingsTable.bucketKey, database.eventSettingsTable.recordType},
		{database.judgingSlotTable.bucketKey, database.judgingSlotTable.recordType},
		{database.lowerThirdTable.bucketKey, database.lowerThirdTable.recordType},
		{database.matchTable.bucketKey, database.matchTable.recordType},
		{database.matchResultTable.bucketKey, database.matchResultTable.recordType},
		{database.rankingTable.bucketKey, database.rankingTable.recordType},
		{database.scheduleBlockTable.bucketKey, database.scheduleBlockTable.recordType},
		{database.scheduledBreakTable.bucketKey, database.scheduledBreakTable.recordType},
		{database.sponsorSlideTable.bucketKey, database.sponsorSlideTable.recordType},
		{database.teamTable.bucketKey, database.teamTable.recordType},
		{database.userSessionTable.bucketKey, database.userSessionTable.recordType},
	}
}

func validateBackupTables(tx *bbolt.Tx, tables []databaseTableDescriptor) error {
	for _, table := range tables {
		bucket := tx.Bucket(table.key)
		if bucket == nil {
			return fmt.Errorf("missing table %s", table.key)
		}
		if err := bucket.ForEach(func(_, value []byte) error {
			if value == nil {
				return fmt.Errorf("nested bucket in table %s", table.key)
			}
			record := reflect.New(table.recordType).Interface()
			if err := json.Unmarshal(value, record); err != nil {
				return fmt.Errorf("invalid record in table %s: %w", table.key, err)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
