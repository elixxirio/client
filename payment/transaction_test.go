package payment

import (
	"github.com/golang/protobuf/proto"
	"gitlab.com/privategrity/crypto/coin"
	"testing"
	"time"
	"gitlab.com/privategrity/client/cmixproto"
	"gitlab.com/privategrity/crypto/id"
)

// TODO are there any error cases for formatting the invoice that we should
// generate errors on?
// Smoke test for formatting the invoice. Makes sure that fields that are put in
// come out.
func TestTransaction_FormatInvoice(t *testing.T) {
	// Generate serialized invoice message
	value := uint64(42)

	sleeve, err := coin.NewSleeve(value)
	if err != nil {
		t.Error(err.Error())
	}
	transaction := Transaction{
		Create:    sleeve,
		Destroy:   nil,
		Change:    NilSleeve,
		Sender:    id.NewUserIDFromUint(2, t),
		Recipient: id.NewUserIDFromUint(5, t),
		Memo:      "Just a test",
		Timestamp: time.Now(),
		Value:     value,
	}

	formattedInvoice := transaction.FormatPaymentInvoice()
	if err != nil {
		t.Error(err.Error())
	}

	// Unpack the serialized invoice message and verify all fields
	var invoice cmixproto.PaymentInvoice
	err = proto.Unmarshal(formattedInvoice.Body, &invoice)
	if err != nil {
		t.Error(err.Error())
	}
	if invoice.Time != transaction.Timestamp.Unix() {
		t.Errorf("Timestamp didn't match. Got %v, expected %v", invoice.Time,
			transaction.Timestamp.Unix())
	}
	if invoice.Memo != transaction.Memo {
		t.Errorf("Memo didn't match. Got %v, expected %v", invoice.Memo,
			transaction.Memo)
	}
	var createdCompound coin.Compound
	copy(createdCompound[:], invoice.CreatedCoin)
	if createdCompound != *sleeve.Compound() {
		t.Errorf("Created compound didn't match. Got %q, expected %q",
			createdCompound, sleeve.Compound())
	}
}
