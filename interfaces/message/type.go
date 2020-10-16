package message

const TypeLen = 4

type Type uint32

const (
	// Used as a wildcard for listeners to listen to all existing types.
	// Think of it as "No type in particular"
	NoType Type = 0

	// A message with no message structure
	// this is a reserved type, messages sent via SendCmix automatically gain
	// this type. Sent messages with this type will be rejected and received
	// non Cmix messages will be ignored
	Raw Type = 1

	//General text message, contains human readable text
	Text Type = 2

	// None of the UDB message types are proto bufs because I haven't had time
	// to migrate UDB fully to the new systems yet.

	// I was considering migrating these types to proto bufs to make them more
	// compact for transmission, but you would have to compress them to even
	// have a chance of fitting the whole key in one Cmix message. In any case,
	// I don't think the benefit is there for the time investment.

	// The prefixes of the UDB response messages are made redundant by the
	// message types in this very enumeration, so at some point we can remove
	// them from the UDB code that generates the responses.

	// The push key message includes two string fields, separated by a space.

	// First field is the key fingerprint, which the UDB uses as an key into
	// the map of, uhh, the keys. This can be any string that doesn't have a
	// space in it.

	// Second field is the key data itself. This should be 2048 bits long
	// (according to the message length that our prime allows) and is
	// base64-encoded.
	UdbPushKey = 10
	// The push key response message is a string. If the key push was a
	// success, the UDB should respond with a message that starts with "PUSHKEY
	// COMPLETE", followed by the fingerprint of the key that was pushed.
	// If the response doesn't begin with "PUSHKEY COMPLETE", the message is
	// an error message and should be shown to the user.
	UdbPushKeyResponse = 11
	// The get key message includes a single string field with the key
	// fingerprint of the key that needs gettin'. This is the same fingerprint
	// you would have pushed.
	UdbGetKey = 12
	// The get key response message is a string. The first space-separated
	// field should always be "GETKEY". The second field is the fingerprint of
	// the key. The third field is "NOTFOUND" if the UDB didn't find the key,
	// or the key itself, encoded in base64, otherwise.
	UdbGetKeyResponse = 13
	// The register message is unchanged from the OG UDB code, except that
	// the REGISTER command in front has been replaced with the type string
	// corresponding to this entry in the enumeration.

	// To wit: The first argument in the list of space-separated fields is
	// the type of the registration. Currently the only allowed type is
	// "EMAIL". The second argument is the value of the type you're registering
	// with. In all currently acceptable registration types, this would be an
	// email address. If you could register with your phone, it would be your
	// phone number, and so on. Then, the key fingerprint of the user's key is
	// the third argument. To register successfully, you must have already
	// pushed the key with that fingerprint.
	UdbRegister = 14
	// The registration response is just a string. It will be either an error
	// message to show to the user, or the message "REGISTRATION COMPLETE" if
	// registration was successful.
	UdbRegisterResponse = 15
	// The search message is just another space separated list. The first field
	// will contain the type of registered user you're searching for, namely
	// "EMAIL". The second field with contain the value of that type that
	// you're searching for.
	UdbSearch = 16
	// The search response is a list of fields. The first is always "SEARCH".
	// The second is always the value that the user searched for. The third is
	// "FOUND" or "NOTFOUND" depending on whether the UDB found the user. If
	// the user was FOUND, the last field will contain their key fingerprint,
	// which you can use with GET_KEY to get the keys you need to talk with
	// that user. Otherwise, this fourth field won't exist.
	UdbSearchResponse = 17

	// The client sends payment transaction messages to the payment bot to
	// fund compound coins with seed coins. In the current implementation,
	// there's one compound that gets funded that's from the payee. This comes
	// across in a PAYMENT_INVOICE. And there's a second compound that contains
	// the change from the seeds that the payer is using to fund the invoice.
	// The rest are the seeds that are the source of the payment.

	// All of the seeds and compounds are in an ordered list, and they get
	// categorized and processed on the payment bot.

	// End to End Rekey message types
	// Trigger a rekey, this message is used locally in client only
	KeyExchangeTrigger = 30
	// Rekey confirmation message. Sent by partner to confirm completion of a rekey
	KeyExchangeConfirm = 31
)
