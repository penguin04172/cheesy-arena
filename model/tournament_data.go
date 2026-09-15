// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"encoding/json"

	"go.etcd.io/bbolt"
)

type TournamentDataClearResult struct {
	Matches         int
	MatchResults    int
	ScheduledBreaks int
	Rankings        int
	Alliances       int
}

// ClearTournamentData atomically removes match data for one stage and its derived records.
func (database *Database) ClearTournamentData(matchType MatchType) (TournamentDataClearResult, error) {
	result := TournamentDataClearResult{}
	err := database.bolt.Update(func(tx *bbolt.Tx) error {
		matchBucket, err := database.matchTable.getBucket(tx)
		if err != nil {
			return err
		}
		matchIds := make(map[int]bool)
		cursor := matchBucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var match Match
			if err = json.Unmarshal(value, &match); err != nil {
				return err
			}
			if match.Type == matchType {
				matchIds[match.Id] = true
				if err = cursor.Delete(); err != nil {
					return err
				}
				result.Matches++
			}
		}
		resultBucket, err := database.matchResultTable.getBucket(tx)
		if err != nil {
			return err
		}
		resultCursor := resultBucket.Cursor()
		for key, value := resultCursor.First(); key != nil; key, value = resultCursor.Next() {
			var matchResult MatchResult
			if err = json.Unmarshal(value, &matchResult); err != nil {
				return err
			}
			if matchIds[matchResult.MatchId] {
				if err = resultCursor.Delete(); err != nil {
					return err
				}
				result.MatchResults++
			}
		}
		breakBucket, err := database.scheduledBreakTable.getBucket(tx)
		if err != nil {
			return err
		}
		breakCursor := breakBucket.Cursor()
		for key, value := breakCursor.First(); key != nil; key, value = breakCursor.Next() {
			var scheduledBreak ScheduledBreak
			if err = json.Unmarshal(value, &scheduledBreak); err != nil {
				return err
			}
			if scheduledBreak.MatchType == matchType {
				if err = breakCursor.Delete(); err != nil {
					return err
				}
				result.ScheduledBreaks++
			}
		}
		if matchType == Qualification {
			result.Rankings, err = clearBucket(tx, database.rankingTable.bucketKey)
			if err != nil {
				return err
			}
		}
		if matchType == Playoff {
			result.Alliances, err = clearBucket(tx, database.allianceTable.bucketKey)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func clearBucket(tx *bbolt.Tx, key []byte) (int, error) {
	bucket := tx.Bucket(key)
	count := 0
	if bucket != nil {
		_ = bucket.ForEach(func(_, _ []byte) error { count++; return nil })
	}
	if err := tx.DeleteBucket(key); err != nil {
		return 0, err
	}
	_, err := tx.CreateBucket(key)
	return count, err
}
