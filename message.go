package framework

type MessageID uint32
type SessionID uint32
type TransactionID uint32
type ParamKey uint32

type Message interface {
	GetID() MessageID
	SetID(MessageID)
	IsSuccess() bool
	SetSuccess(flag bool)
	SetSender(string)
	GetSender() string
	GetFromSession() SessionID
	SetFromSession(session SessionID)
	GetToSession() SessionID
	SetToSession(session SessionID)
	SetTransactionID(id TransactionID)
	GetTransactionID() TransactionID

	SetError(msg string)
	GetError() string

	GetString(key ParamKey) (string, error)
	GetUInt(key ParamKey) (uint, error)
	GetInt(key ParamKey) (int, error)
	GetFloat(key ParamKey) (float64, error)
	GetBoolean(key ParamKey) (bool, error)
	SetString(key ParamKey, value string)
	SetUInt(key ParamKey, value uint)
	SetInt(key ParamKey, value int)
	SetFloat(key ParamKey, value float64)
	SetBoolean(key ParamKey, value bool)

	SetUIntArray(key ParamKey, value []uint64)
	GetUIntArray(key ParamKey) ([]uint64, error)

	SetStringArray(key ParamKey, value []string)
	GetStringArray(key ParamKey) ([]string, error)

	Serialize() ([]byte, error)

	GetAllString() (map[ParamKey]string)
	GetAllUInt() (map[ParamKey]uint)
	GetAllInt() (map[ParamKey]int)
	GetAllFloat() (map[ParamKey]float64)
	GetAllBoolean() (map[ParamKey]bool)
	GetAllUIntArray() (map[ParamKey][]uint64)
	GetAllStringArray() (map[ParamKey][]string)
}




