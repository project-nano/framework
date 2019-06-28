package framework

import (
	"encoding/json"
	"fmt"
)

type JsonMessage struct {
	ID                MessageID             `json:"id"`
	Success           bool                  `json:"success,omitempty"`
	Sender            string                `json:"-"`
	From              SessionID             `json:"from,omitempty"`
	To                SessionID             `json:"to,omitempty"`
	Transaction       TransactionID         `json:"transaction,omitempty"`
	Error             string                `json:"error,omitempty"`
	BoolParams        map[ParamKey]bool     `json:"bool_params,omitempty"`
	StringParams      map[ParamKey]string   `json:"string_params,omitempty"`
	UIntParams        map[ParamKey]uint     `json:"uint_params,omitempty"`
	IntParams         map[ParamKey]int      `json:"int_params,omitempty"`
	FloatParams       map[ParamKey]float64  `json:"float_params,omitempty"`
	UIntArrayParams   map[ParamKey][]uint64 `json:"uint_array_params,omitempty"`
	StringArrayParams map[ParamKey][]string `json:"string_array_params,omitempty"`
}

func CreateJsonMessage(msg MessageID) (*JsonMessage, error) {
	return &JsonMessage{ID: msg}, nil
}

func CloneJsonMessage(origin Message) (clone *JsonMessage) {
	clone, _ = CreateJsonMessage(origin.GetID())
	clone.SetSuccess(origin.IsSuccess())
	clone.SetFromSession(origin.GetFromSession())
	clone.SetToSession(origin.GetToSession())
	clone.SetTransactionID(origin.GetTransactionID())
	if "" != origin.GetError(){
		clone.SetError(origin.GetError())
	}
	//clone params
	if 0 != len(origin.GetAllBoolean()){
		for key, value := range origin.GetAllBoolean(){
			clone.SetBoolean(key, value)
		}
	}
	if 0 != len(origin.GetAllString()){
		for key, value := range origin.GetAllString(){
			clone.SetString(key, value)
		}
	}
	if 0 != len(origin.GetAllUInt()){
		for key, value := range origin.GetAllUInt(){
			clone.SetUInt(key, value)
		}
	}
	if 0 != len(origin.GetAllInt()){
		for key, value := range origin.GetAllInt(){
			clone.SetInt(key, value)
		}
	}
	if 0 != len(origin.GetAllFloat()){
		for key, value := range origin.GetAllFloat(){
			clone.SetFloat(key, value)
		}
	}
	if 0 != len(origin.GetAllUIntArray()){
		for key, value := range origin.GetAllUIntArray(){
			clone.SetUIntArray(key, value)
		}
	}
	if 0 != len(origin.GetAllStringArray()){
		for key, value := range origin.GetAllStringArray(){
			clone.SetStringArray(key, value)
		}
	}
	return clone
}

func MessageFromJson(data []byte) (*JsonMessage, error){
	var msg JsonMessage
	var err = json.Unmarshal(data, &msg)
	return &msg, err
}

func (msg *JsonMessage) GetID() MessageID{
	return msg.ID
}

func (msg *JsonMessage) SetID(id MessageID){
	msg.ID = id
}

func (msg *JsonMessage)IsSuccess() bool{
	return msg.Success
}
func (msg *JsonMessage)SetSuccess(flag bool){
	msg.Success = flag
}
func (msg *JsonMessage)SetSender(value string){
	msg.Sender = value
}
func (msg *JsonMessage)GetSender() string{
	return msg.Sender
}

func (msg *JsonMessage)GetFromSession() SessionID{
	return msg.From
}
func (msg *JsonMessage)SetFromSession(session SessionID){
	msg.From = session
}
func (msg *JsonMessage)GetToSession() SessionID{
	return msg.To
}
func (msg *JsonMessage)SetToSession(session SessionID){
	msg.To = session
}

func (msg *JsonMessage)SetTransactionID(id TransactionID){
	msg.Transaction = id
}
func (msg *JsonMessage)GetTransactionID() TransactionID{
	return msg.Transaction
}

func (msg *JsonMessage)SetError(err string){
	msg.Error = err
}
func (msg *JsonMessage)GetError() string{
	return msg.Error
}


func (msg *JsonMessage)GetString(key ParamKey) (string, error){
	if msg.StringParams != nil{
		if value, exists := msg.StringParams[key]; exists{
			return value, nil
		}
	}
	return "", fmt.Errorf("no string param for key %d", key)
}

func (msg *JsonMessage)GetUInt(key ParamKey) (uint, error){
	if msg.UIntParams != nil{
		if value, exists := msg.UIntParams[key]; exists{
			return value, nil
		}
	}
	return 0, fmt.Errorf("no uint param for key %d", key)
}

func (msg *JsonMessage)GetInt(key ParamKey) (int, error){
	if msg.IntParams != nil{
		if value, exists := msg.IntParams[key]; exists{
			return value, nil
		}
	}
	return 0, fmt.Errorf("no int param for key %d", key)
}

func (msg *JsonMessage)GetFloat(key ParamKey) (float64, error){
	if msg.FloatParams != nil{
		if value, exists := msg.FloatParams[key]; exists{
			return value, nil
		}
	}
	return 0.0, fmt.Errorf("no float param for key %d", key)
}

func (msg *JsonMessage)GetBoolean(key ParamKey) (bool, error){
	if msg.BoolParams != nil{
		if value, exists := msg.BoolParams[key]; exists{
			return value, nil
		}
	}
	return false, fmt.Errorf("no bool param for key %d", key)
}


func (msg *JsonMessage)SetString(key ParamKey, value string){
	if msg.StringParams != nil{
		msg.StringParams[key] = value
	}else{
		msg.StringParams = map[ParamKey]string{key:value}
	}
}

func (msg *JsonMessage)SetUInt(key ParamKey, value uint){
	if msg.UIntParams != nil{
		msg.UIntParams[key] = value
	}else{
		msg.UIntParams = map[ParamKey]uint{key:value}
	}
}

func (msg *JsonMessage)SetInt(key ParamKey, value int){
	if msg.IntParams != nil{
		msg.IntParams[key] = value
	}else{
		msg.IntParams = map[ParamKey]int{key:value}
	}
}

func (msg *JsonMessage)SetFloat(key ParamKey, value float64){
	if msg.FloatParams != nil{
		msg.FloatParams[key] = value
	}else{
		msg.FloatParams = map[ParamKey]float64{key:value}
	}
}

func (msg *JsonMessage)SetBoolean(key ParamKey, value bool){
	if msg.BoolParams != nil{
		msg.BoolParams[key] = value
	}else{
		msg.BoolParams = map[ParamKey]bool{key:value}
	}
}

func (msg *JsonMessage)SetUIntArray(key ParamKey, value []uint64){
	if msg.UIntArrayParams != nil{
		msg.UIntArrayParams[key] = value
	}else{
		msg.UIntArrayParams = map[ParamKey][]uint64{key:value}
	}
}
func (msg *JsonMessage)GetUIntArray(key ParamKey) ([]uint64, error){
	if msg.UIntArrayParams != nil{
		if value, exists := msg.UIntArrayParams[key]; exists{
			return value, nil
		}
	}
	return nil, fmt.Errorf("no uint array for key %d", key)
}

func (msg *JsonMessage)SetStringArray(key ParamKey, value []string){
	if msg.StringArrayParams != nil{
		msg.StringArrayParams[key] = value
	}else{
		msg.StringArrayParams = map[ParamKey][]string{key:value}
	}
}
func (msg *JsonMessage)GetStringArray(key ParamKey) ([]string, error){
	if msg.StringArrayParams != nil{
		if value, exists := msg.StringArrayParams[key]; exists{
			return value, nil
		}
	}
	return nil, fmt.Errorf("no string array for key %d", key)
}

func (msg *JsonMessage)Serialize() ([]byte, error){
	return json.Marshal(msg)
}

var (
	emptyFloatMap        = map[ParamKey]float64{}
	emptyStringMap      = map[ParamKey]string{}
	emptyUIntMap        = map[ParamKey]uint{}
	emptyIntMap         = map[ParamKey]int{}
	emptyBooleanMap     = map[ParamKey]bool{}
	emptyStringArrayMap = map[ParamKey][]string{}
	emptyUIntArrayMap   = map[ParamKey][]uint64{}
)

func (msg *JsonMessage) GetAllString() (map[ParamKey]string){
	if msg.StringParams != nil{
		return msg.StringParams
	}
	return emptyStringMap
}
func (msg *JsonMessage) GetAllUInt() (map[ParamKey]uint){
	if msg.UIntParams != nil{
		return msg.UIntParams
	}
	return emptyUIntMap
}

func (msg *JsonMessage) GetAllInt() (map[ParamKey]int){
	if msg.IntParams != nil{
		return msg.IntParams
	}
	return emptyIntMap
}

func (msg *JsonMessage) GetAllFloat() (map[ParamKey]float64){
	if msg.FloatParams != nil{
		return msg.FloatParams
	}
	return emptyFloatMap
}

func (msg *JsonMessage) GetAllBoolean() (map[ParamKey]bool){
	if msg.BoolParams != nil{
		return msg.BoolParams
	}
	return emptyBooleanMap
}

func (msg *JsonMessage) GetAllUIntArray() (map[ParamKey][]uint64){
	if msg.UIntArrayParams != nil{
		return msg.UIntArrayParams
	}
	return emptyUIntArrayMap
}

func (msg *JsonMessage) GetAllStringArray() (map[ParamKey][]string)	{
	if msg.StringArrayParams != nil{
		return msg.StringArrayParams
	}
	return emptyStringArrayMap
}