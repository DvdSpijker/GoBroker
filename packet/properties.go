package packet

type (
  PropertyIdentifier byte
	PayloadFormatIndicator byte
  )


const (
	UnspecifiedBytes PayloadFormatIndicator = 0x00
	UtfCharacterData PayloadFormatIndicator = 0x01
)

const (
  WillDelayIntervalProperty PropertyIdentifier = 0x18
  PayloadFormatIndicatorProperty PropertyIdentifier = 0x01
  MessageExpiryIntervalProperty PropertyIdentifier = 0x02
  ContentTypeProperty PropertyIdentifier = 0x03
  ResponseTopicProperty PropertyIdentifier = 0x08
  CorrelationDataProperty PropertyIdentifier = 0x09
  UserPropertyProperty PropertyIdentifier = 0x26
)
