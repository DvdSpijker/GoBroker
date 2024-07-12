package packet


type ReasonCode byte

const (
	Success                     ReasonCode = 0x00
	UnspecifiedError            ReasonCode = 0x80
	MalformedPacket             ReasonCode = 0x81
	ProtocolError               ReasonCode = 0x82
	ImplementationSpecificError ReasonCode = 0x83
	UnsupportedProtocolVersion  ReasonCode = 0x84
	ClientIdentifierNotValid    ReasonCode = 0x85
	BadUserNameOrPassword       ReasonCode = 0x86
	NotAuthenterized            ReasonCode = 0x87
	ServerUnavailable           ReasonCode = 0x88
	ServerBusy                  ReasonCode = 0x89
	Banned                      ReasonCode = 0x8A
	BadAuthenticationMethod     ReasonCode = 0x8C
	TopicNameInvalid            ReasonCode = 0x90
	PacketTooLarge              ReasonCode = 0x95 // (That's what she said)
	QuotaExceeded               ReasonCode = 0x97
	PayloadFormatInvalid        ReasonCode = 0x99
	RetainNotSupported          ReasonCode = 0x9A
	QosNotSupported             ReasonCode = 0x9B
	UseAnotherServer            ReasonCode = 0x9C
	ServerMoved                 ReasonCode = 0x9D
	ConnectionRateExceeded      ReasonCode = 0x9F
)
