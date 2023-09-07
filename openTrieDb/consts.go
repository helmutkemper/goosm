package opentriedb

const (
	headerLength     = 0 //122
	byteSize         = 1
	uint64Size       = 8
	uint32Size       = 4
	uint64AddressInc = uint64Size * byteSize
	uint32AddressInc = uint32Size * byteSize

	pieceAddressInc = 4 * byteSize

	kLineLengthSize      = uint32Size
	kLineNextAddressSize = 8
	kLineAddressSize     = 8

	//kLineFormula = kLineLengthSize*8 + kLineNextAddressSize*8 + (kLineAddressSize*8+linePieceSize*8)*kDataLineLength*8

	kAddressByteSize = 8
	kRuneByteSize    = 4
	kLengthByteSize  = 4
)
