package e2e

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/e2e/ratchet/partner/session"
)

// fpGenerator is a wrapper that allows the network manager's fingerprint
// interface to be passed into ratchet without exposing ratchet to the business
// logic.
type fpGenerator struct {
	m *manager
}

// AddKey adds a fingerprint to be tracked for the given cypher.
func (fpg *fpGenerator) AddKey(cy session.Cypher) {
	err := fpg.m.net.AddFingerprint(
		fpg.m.myID, cy.Fingerprint(), &processor{cy, fpg.m})
	if err != nil {
		jww.ERROR.Printf(
			"Could not add fingerprint %s: %+v", cy.Fingerprint(), err)
	}
}

// DeleteKey deletes the fingerprint for the given cypher.
func (fpg *fpGenerator) DeleteKey(cy session.Cypher) {
	fpg.m.net.DeleteFingerprint(fpg.m.myID, cy.Fingerprint())
}
