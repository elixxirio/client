package ud

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
)

// ConfirmFact confirms a fact first registered via AddFact. The
// confirmation ID comes from AddFact while the code will come over the
// associated communications system.
func (m *Manager) ConfirmFact(confirmationID, code string) error {
	jww.INFO.Printf("ud.ConfirmFact(%s, %s)", confirmationID, code)
	if err := m.confirmFact(confirmationID, code, m.comms); err != nil {
		return errors.WithMessage(err, "Failed to confirm fact")
	}
	return nil
}

// confirmFact is a helper function for ConfirmFact.
func (m *Manager) confirmFact(confirmationID, code string, comm confirmFactComm) error {
	// get UD host
	udHost, err := m.getOrAddUdHost()
	if err != nil {
		return err
	}

	msg := &pb.FactConfirmRequest{
		ConfirmationID: confirmationID,
		Code:           code,
	}
	_, err = comm.SendConfirmFact(udHost, msg)
	if err != nil {
		return err
	}

	err = m.store.ConfirmFact(confirmationID)
	if err != nil {
		return errors.WithMessagef(err,
			"Failed to confirm fact in storage with confirmation ID: %q",
			confirmationID)
	}

	return nil
}
