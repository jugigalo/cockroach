// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Kenji Kaneda (kenji.kaneda@gmail.com)

package storage_test

import (
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/storage"
	"github.com/cockroachdb/cockroach/util"
	"github.com/cockroachdb/cockroach/util/leaktest"
)

// TestRangeGCQueueDropReplica verifies that the range GC queue
// removes a range from a store that no longer should have a replica.
func TestRangeGCQueueDropReplica(t *testing.T) {
	defer leaktest.AfterTest(t)

	mtc := startMultiTestContext(t, 3)
	defer mtc.Stop()
	raftID := int64(1)
	mtc.replicateRange(raftID, 0, 1, 2)

	mtc.unreplicateRange(raftID, 0, 1)

	// Increment the clock's timestamp to expire the leader lease.
	mtc.manualClock.Increment(int64(storage.DefaultLeaderLeaseDuration) + 1)

	// Make sure the range is not yet removed from the store.
	numTrials := 3
	for i := 0; i < numTrials; i++ {
		store := mtc.stores[1]
		store.ForceRangeGCScan(t)
		if _, err := store.GetRange(raftID); err != nil {
			t.Error("unexpected range removal")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Increment the clock's timestamp to make the range GC queue process the range.
	mtc.manualClock.Increment(int64(storage.RangeGCQueueUnleasedDuration))

	// Make sure the range is removed from the store.
	util.SucceedsWithin(t, time.Second, func() error {
		store := mtc.stores[1]
		store.ForceRangeGCScan(t)
		if _, err := store.GetRange(raftID); err == nil {
			return util.Error("expected range removal")
		}
		return nil
	})

	// Restart the store to tear down the test cleanly.
	mtc.stopStore(1)
	mtc.restartStore(1)
}
