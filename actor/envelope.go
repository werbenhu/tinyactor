package actor

type Envelope struct {
	Recipient int64
	Sender    int64
	Message   []byte
}
