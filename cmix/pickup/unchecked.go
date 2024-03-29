////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package pickup

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/v4/cmix/pickup/store"
	"gitlab.com/elixxir/client/v4/stoppable"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
	"time"
)

// Constants for message retrieval backoff delays
// TODO: Make this a real backoff
const (
	tryZero  = 10 * time.Second
	tryOne   = 30 * time.Second
	tryTwo   = 5 * time.Minute
	tryThree = 30 * time.Minute
	tryFour  = 3 * time.Hour
	tryFive  = 12 * time.Hour
	trySix   = 24 * time.Hour
	// Amount of tries past which the backoff will not increase
	cappedTries = 7
)

var backOffTable = [cappedTries]time.Duration{
	tryZero, tryOne, tryTwo, tryThree, tryFour, tryFive, trySix}

// processUncheckedRounds will (periodically) check every checkInterval for
// rounds that failed message retrieval in processMessageRetrieval. Rounds will
// have a backoff duration in which they will be tried again. If a round is
// found to be due on a periodical check, the round is sent back to
// processMessageRetrieval.
// TODO: Make this system know which rounds are still in progress instead of
//
//	just assume by time
func (m *pickup) processUncheckedRounds(checkInterval time.Duration,
	backoffTable [cappedTries]time.Duration, stop *stoppable.Single) {
	ticker := time.NewTicker(checkInterval)
	uncheckedRoundStore := m.unchecked
	for {
		iterator := func(rid id.Round, rnd store.UncheckedRound) {
			jww.DEBUG.Printf(
				"Checking if round %d is due for a message lookup.", rid)
			// If this round is due for a round check, send the round over
			// to the retrieval thread. If not due, then check next round.
			if !isRoundCheckDue(rnd.NumChecks, rnd.LastCheck, backoffTable) {
				return
			}

			jww.INFO.Printf(
				"Round %d due for a message lookup, retrying...", rid)

			// Check if it needs to be processed by historical Rounds
			m.GetMessagesFromRound(rid, receptionID.EphemeralIdentity{
				EphId:  rnd.EpdId,
				Source: rnd.Source,
			})

			// Update the state of the round for next look-up (if needed)
			err := uncheckedRoundStore.IncrementCheck(rid, rnd.Source, rnd.EpdId)
			if err != nil {
				jww.ERROR.Printf("processUncheckedRounds error: Could not "+
					"increment check attempts for round %d: %v", rid, err)
			}
		}

		// Pull and iterate through uncheckedRound list
		m.unchecked.IterateOverList(iterator)
		select {
		case <-stop.Quit():
			ticker.Stop()
			stop.ToStopped()
			return
		case <-ticker.C:
		}
	}
}

// isRoundCheckDue determines whether this round is due for another check given
// the amount of tries and the timestamp the round was stored. Returns true if a
// new check is due
func isRoundCheckDue(tries uint64, ts time.Time,
	backoffTable [cappedTries]time.Duration) bool {
	now := netTime.Now()

	if tries >= uint64(len(backoffTable)) {
		tries = uint64(len(backoffTable)) - 1
	}
	roundCheckTime := ts.Add(backoffTable[tries])

	return now.After(roundCheckTime)
}
