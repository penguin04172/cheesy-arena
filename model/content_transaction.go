// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"encoding/json"
	"fmt"
	"sort"

	"go.etcd.io/bbolt"
)

// SaveAwardWithLowerThirds atomically creates or updates an award and its associated lower thirds.
func (database *Database) SaveAwardWithLowerThirds(award *Award, desired []LowerThird) error {
	originalId := award.Id
	err := database.bolt.Update(func(tx *bbolt.Tx) error {
		awardBucket, err := database.awardTable.getBucket(tx)
		if err != nil {
			return err
		}
		if award.Id == 0 {
			sequence, err := awardBucket.NextSequence()
			if err != nil {
				return err
			}
			award.Id = int(sequence)
		} else if awardBucket.Get(idToKey(award.Id)) == nil {
			return fmt.Errorf("can't update non-existent Award with ID %d", award.Id)
		}
		encodedAward, err := json.Marshal(award)
		if err != nil {
			return err
		}
		if err = awardBucket.Put(idToKey(award.Id), encodedAward); err != nil {
			return err
		}

		lowerThirdBucket, err := database.lowerThirdTable.getBucket(tx)
		if err != nil {
			return err
		}
		existing := make([]LowerThird, 0)
		maxOrder := 0
		if err = lowerThirdBucket.ForEach(func(_, value []byte) error {
			var lowerThird LowerThird
			if err := json.Unmarshal(value, &lowerThird); err != nil {
				return err
			}
			if lowerThird.DisplayOrder > maxOrder {
				maxOrder = lowerThird.DisplayOrder
			}
			if lowerThird.AwardId == award.Id {
				existing = append(existing, lowerThird)
			}
			return nil
		}); err != nil {
			return err
		}
		sort.Slice(existing, func(i, j int) bool { return existing[i].DisplayOrder < existing[j].DisplayOrder })
		for index := range desired {
			desired[index].AwardId = award.Id
			if index < len(existing) {
				desired[index].Id, desired[index].DisplayOrder = existing[index].Id, existing[index].DisplayOrder
			} else {
				sequence, err := lowerThirdBucket.NextSequence()
				if err != nil {
					return err
				}
				maxOrder++
				desired[index].Id, desired[index].DisplayOrder = int(sequence), maxOrder
			}
			encoded, err := json.Marshal(&desired[index])
			if err != nil {
				return err
			}
			if err = lowerThirdBucket.Put(idToKey(desired[index].Id), encoded); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		award.Id = originalId
	}
	return err
}

// DeleteAwardWithLowerThirds atomically removes an award and every associated lower third.
func (database *Database) DeleteAwardWithLowerThirds(awardId int) error {
	return database.bolt.Update(func(tx *bbolt.Tx) error {
		awardBucket, err := database.awardTable.getBucket(tx)
		if err != nil {
			return err
		}
		if awardBucket.Get(idToKey(awardId)) == nil {
			return fmt.Errorf("can't delete non-existent Award with ID %d", awardId)
		}
		if err = awardBucket.Delete(idToKey(awardId)); err != nil {
			return err
		}
		lowerThirdBucket, err := database.lowerThirdTable.getBucket(tx)
		if err != nil {
			return err
		}
		cursor := lowerThirdBucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var lowerThird LowerThird
			if err = json.Unmarshal(value, &lowerThird); err != nil {
				return err
			}
			if lowerThird.AwardId == awardId {
				if err = cursor.Delete(); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (database *Database) ReorderSponsorSlides(ids []int) error {
	return reorderContent(database.bolt, database.sponsorSlideTable, ids,
		func(value SponsorSlide) int { return value.Id }, func(value *SponsorSlide, order int) { value.DisplayOrder = order })
}

func (database *Database) ReorderLowerThirds(ids []int) error {
	return reorderContent(database.bolt, database.lowerThirdTable, ids,
		func(value LowerThird) int { return value.Id }, func(value *LowerThird, order int) { value.DisplayOrder = order })
}

func reorderContent[R any](bolt *bbolt.DB, table *table[R], ids []int, idOf func(R) int, setOrder func(*R, int)) error {
	return bolt.Update(func(tx *bbolt.Tx) error {
		bucket, err := table.getBucket(tx)
		if err != nil {
			return err
		}
		records := make(map[int]R)
		if err = bucket.ForEach(func(_, value []byte) error {
			var record R
			if err := json.Unmarshal(value, &record); err != nil {
				return err
			}
			records[idOf(record)] = record
			return nil
		}); err != nil {
			return err
		}
		if len(ids) != len(records) {
			return fmt.Errorf("order must contain every resource ID exactly once")
		}
		seen := make(map[int]bool, len(ids))
		for _, id := range ids {
			if _, exists := records[id]; !exists || seen[id] {
				return fmt.Errorf("order must contain every resource ID exactly once")
			}
			seen[id] = true
		}
		for index, id := range ids {
			record := records[id]
			setOrder(&record, index+1)
			encoded, err := json.Marshal(&record)
			if err != nil {
				return err
			}
			if err = bucket.Put(idToKey(id), encoded); err != nil {
				return err
			}
		}
		return nil
	})
}
