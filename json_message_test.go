package framework

import "testing"
import (
	"math/rand"
	"time"
	"encoding/base64"
	"fmt"
	"errors"
	"strings"
)

const (
	testRepeat = 10
)

var generator = rand.New(rand.NewSource(time.Now().UnixNano()))

var checkConsistency = func(t *testing.T, f func(msg *JsonMessage)) {
	for i := 0; i < testRepeat; i++{
		origin, err := generateMessage()
		if err != nil{
			t.Fatalf("generate message fail: %s", err.Error())
		}
		f(origin)
		data, err := origin.Serialize()
		if err != nil{
			t.Fatalf("serialize fail: %s", err.Error())
		}
		target, err := MessageFromJson(data)
		if err != nil{
			t.Fatalf("parse message fail: %s", err.Error())
		}
		result, err := isIdentical(origin, target, t)
		if err != nil{
			t.Fatalf("identify fail: %s", err.Error())
		}
		if !result{
			t.Fatal("message corrupted")
		}
		t.Logf("%dth try success", i + 1)
	}
}

func Test_BaseMember(t *testing.T){
	var empty = func(msg *JsonMessage) {}
	checkConsistency(t, empty)
}

func Test_BoolParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateBoolParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_StringParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateStringParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_UIntParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateUIntParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_IntParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateIntParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_FloatParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateFloatParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_UIntArrayParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateUIntArrayParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_StringArrayParam(t *testing.T){
	const paramCount = 5
	var prepare = func(msg *JsonMessage) {
		generateStringArrayParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_MixedParam(t *testing.T){
	const paramCount = 3
	var prepare = func(msg *JsonMessage) {
		generateBoolParam(msg, paramCount)
		generateUIntParam(msg, paramCount)
		generateIntParam(msg, paramCount)
		generateStringParam(msg, paramCount)
		generateFloatParam(msg, paramCount)
		generateUIntArrayParam(msg, paramCount)
		generateStringArrayParam(msg, paramCount)
	}
	checkConsistency(t, prepare)
}

func Test_CloneMessage(t *testing.T) {
	const (
		paramCount = 3
	)
	for i := 0; i < testRepeat; i++{
		origin, err := generateMessage()
		if err != nil{
			t.Fatalf("generate message fail: %s", err.Error())
		}
		generateBoolParam(origin, paramCount)
		generateUIntParam(origin, paramCount)
		generateIntParam(origin, paramCount)
		generateStringParam(origin, paramCount)
		generateFloatParam(origin, paramCount)
		generateUIntArrayParam(origin, paramCount)
		generateStringArrayParam(origin, paramCount)

		var clone = CloneJsonMessage(origin)
		if err != nil{
			t.Fatalf("clone message fail: %s", err.Error())
		}
		identical, err := isIdentical(origin, clone, t)
		if err != nil{
			t.Fatalf("compare %d clone fail: %s", i, err.Error())
		}
		if !identical {
			t.Fatalf("%dth clone is not identical", i)
		}
		t.Logf("%dth clone is identical", i)
	}

}

func generateMessage() (msg *JsonMessage, err error){
	const (
		ErrorLength = 32
	)
	msg, err = CreateJsonMessage(MessageID(generator.Uint32()))
	msg.SetFromSession(SessionID(generator.Uint32()))
	msg.SetToSession(SessionID(generator.Uint32()))
	msg.SetTransactionID(TransactionID(generator.Uint32()))
	if generator.Intn(2) > 0 {
		msg.SetSuccess(true)
	}else{
		var buf = make([]byte, ErrorLength)
		_, err = generator.Read(buf)
		if err != nil{
			return
		}
		msg.SetError(base64.StdEncoding.EncodeToString(buf))
		msg.SetSuccess(false)
	}
	return msg, nil
}

func generateBoolParam(msg Message, count int){
	for i := 0; i < count; i++{
		var key = ParamKey(generator.Uint32())
		if generator.Intn(2) > 0 {
			msg.SetBoolean(key, true)
		}else{
			msg.SetBoolean(key, false)
		}
	}
}


func generateStringParam(msg Message, count int){
	const (
		MinStringLength = 6
		MaxStringLength = 20
	)
	var bufSize = MinStringLength + generator.Intn(MaxStringLength - MinStringLength)
	var buf = make([]byte, bufSize)
	for i := 0; i < count; i++{
		generator.Read(buf)
		var key = ParamKey(generator.Uint32())
		msg.SetString(key, base64.StdEncoding.EncodeToString(buf))
	}
}


func generateUIntParam(msg Message, count int){
	for i := 0; i < count; i++{
		var key = ParamKey(generator.Uint32())
		var value = uint(generator.Uint64())
		msg.SetUInt(key, value)
	}
}


func generateIntParam(msg Message, count int){
	for i := 0; i < count; i++{
		var key = ParamKey(generator.Uint32())
		var value = generator.Int()
		msg.SetInt(key, value)
	}
}


func generateFloatParam(msg Message, count int){
	for i := 0; i < count; i++{
		var key = ParamKey(generator.Uint32())
		var value = generator.Float64()
		msg.SetFloat(key, value)
	}
}

func generateStringArrayParam(msg Message, count int){
	const (
		MinArrayLength = 1
		MaxArrayLength = 10
		MinStringLength = 6
		MaxStringLength = 20
	)
	var arrayLength = MinArrayLength + generator.Intn(MaxArrayLength - MinArrayLength)
	var bufSize = MinStringLength + generator.Intn(MaxStringLength - MinStringLength)
	var buf = make([]byte, bufSize)
	for i := 0; i < count; i++{
		var value = make([]string, arrayLength)
		for j := 0; j < arrayLength; j++{
			generator.Read(buf)
			value[j] = base64.StdEncoding.EncodeToString(buf)
		}
		var key = ParamKey(generator.Uint32())
		msg.SetStringArray(key, value)
	}
}


func generateUIntArrayParam(msg Message, count int){
	const (
		MinArrayLength = 1
		MaxArrayLength = 10
	)
	var size = MinArrayLength + generator.Intn(MaxArrayLength - MinArrayLength)
	for i := 0; i < count; i++{
		var value = make([]uint64, size)
		for j := 0; j < size; j++{
			value[j] = generator.Uint64()
		}
		var key = ParamKey(generator.Uint32())
		msg.SetUIntArray(key, value)
	}
}



func isIdentical(source, target *JsonMessage, t *testing.T) (bool, error){
	if source.ID != target.ID{
		return false, fmt.Errorf("different ID %d / %d", source.ID, target.ID)
	}

	if source.Success != target.Success{
		return false, fmt.Errorf("different success flag %t / %t", source.Success, target.Success)
	}
	if source.From != target.From{
		return false, fmt.Errorf("different from property %d / %d", source.From, target.From)
	}
	if source.To != target.To{
		return false, fmt.Errorf("different to property %d / %d", source.To, target.To)
	}
	if source.Transaction != target.Transaction{
		return false, fmt.Errorf("different transaction property %d / %d", source.Transaction, target.Transaction)
	}
	if source.Error != target.Error{
		return false, fmt.Errorf("different errror message %s / %s", source.Error, target.Error)
	}
	if t != nil{
		t.Logf("msg %08X from %08X to %08X, status %t, trans %08X",
			target.GetID(), target.GetFromSession(), target.GetToSession(), target.IsSuccess(), target.GetTransactionID())
	}
	{
		//bool
		if source.BoolParams == nil && target.BoolParams != nil{
			return false, errors.New("unexpected bool param")
		} else if source.BoolParams != nil{
			if target.BoolParams == nil{
				return false, errors.New("bool param lost")
			}
			for key, sourceValue := range source.BoolParams{
				targetValue, exists := target.BoolParams[key]
				if !exists{
					return false, fmt.Errorf("no bool param available for key %d", key)
				}
				if targetValue != sourceValue{
					return false, fmt.Errorf("different bool param of key %d, %t => %t", key, sourceValue, targetValue)
				}
				if t != nil{
					t.Logf("bool %d: %t", key, targetValue)
				}
			}
		}
	}
	{
		//string
		if source.StringParams == nil && target.StringParams != nil{
			return false, errors.New("unexpected string param")
		} else if source.StringParams != nil{
			if target.StringParams == nil{
				return false, errors.New("string param lost")
			}
			for key, sourceValue := range source.StringParams{
				targetValue, exists := target.StringParams[key]
				if !exists{
					return false, fmt.Errorf("no string param available for key %d", key)
				}
				if targetValue != sourceValue{
					return false, fmt.Errorf("different string param of key %d, %s => %s", key, sourceValue, targetValue)
				}
				if t != nil{
					t.Logf("string %d: %s", key, targetValue)
				}

			}
		}
	}
	{
		//uint
		if source.UIntParams == nil && target.UIntParams != nil{
			return false, errors.New("unexpected uint param")
		} else if source.UIntParams != nil{
			if target.UIntParams == nil{
				return false, errors.New("uint param lost")
			}
			for key, sourceValue := range source.UIntParams{
				targetValue, exists := target.UIntParams[key]
				if !exists{
					return false, fmt.Errorf("no uint param available for key %d", key)
				}
				if targetValue != sourceValue{
					return false, fmt.Errorf("different uint param of key %d, %d => %d", key, sourceValue, targetValue)
				}
				if t != nil{
					t.Logf("uint %d: %d", key, targetValue)
				}

			}
		}
	}
	{
		//int
		if source.IntParams == nil && target.IntParams != nil{
			return false, errors.New("unexpected int param")
		} else if source.IntParams != nil{
			if target.IntParams == nil{
				return false, errors.New("int param lost")
			}
			for key, sourceValue := range source.IntParams{
				targetValue, exists := target.IntParams[key]
				if !exists{
					return false, fmt.Errorf("no int param available for key %d", key)
				}
				if targetValue != sourceValue{
					return false, fmt.Errorf("different int param of key %d, %d => %d", key, sourceValue, targetValue)
				}
				if t != nil{
					t.Logf("int %d: %d", key, targetValue)
				}

			}
		}
	}
	{
		//float
		if source.FloatParams == nil && target.FloatParams != nil{
			return false, errors.New("unexpected float param")
		} else if source.FloatParams != nil{
			if target.FloatParams == nil{
				return false, errors.New("float param lost")
			}
			for key, sourceValue := range source.FloatParams{
				targetValue, exists := target.FloatParams[key]
				if !exists{
					return false, fmt.Errorf("no float param available for key %d", key)
				}
				if targetValue != sourceValue{
					return false, fmt.Errorf("different float param of key %d, %f => %f", key, sourceValue, targetValue)
				}
				if t != nil{
					t.Logf("float %d: %f", key, targetValue)
				}

			}
		}
	}
	{
		//uint array
		if source.UIntArrayParams == nil && target.UIntArrayParams != nil{
			return false, errors.New("unexpected uint array param")
		} else if source.UIntArrayParams != nil{
			if target.UIntArrayParams == nil{
				return false, errors.New("uint array param lost")
			}
			for key, sourceArray := range source.UIntArrayParams{
				targetArray, exists := target.UIntArrayParams[key]
				if !exists{
					return false, fmt.Errorf("no uint array param available for key %d", key)
				}
				var arrayCount = len(sourceArray)
				if arrayCount != len(targetArray){
					return false, fmt.Errorf("different uint array length for key %d, %d => %d", key, arrayCount, len(targetArray))
				}
				for i := 0; i < arrayCount; i++{
					if sourceArray[i] != targetArray[i]{
						return false, fmt.Errorf("%dth element in uint array with key %d unmatched, %d => %d", i, key, sourceArray[i], targetArray[i])
					}
				}
				if t != nil{
					var array = make([]string, 0)
					for _, value := range targetArray{
						array = append(array, fmt.Sprintf("%d", value))
					}
					t.Logf("uint array %d: [%s]", key, strings.Join(array, ","))
				}

			}
		}
	}
	{
		//string array
		if source.StringArrayParams == nil && target.StringArrayParams != nil{
			return false, errors.New("unexpected string array param")
		} else if source.StringArrayParams != nil{
			if target.StringArrayParams == nil{
				return false, errors.New("string array param lost")
			}
			for key, sourceArray := range source.StringArrayParams{
				targetArray, exists := target.StringArrayParams[key]
				if !exists{
					return false, fmt.Errorf("no string array param available for key %d", key)
				}
				var arrayCount = len(sourceArray)
				if arrayCount != len(targetArray){
					return false, fmt.Errorf("different string array length for key %d, %d => %d", key, arrayCount, len(targetArray))
				}
				for i := 0; i < arrayCount; i++{
					if sourceArray[i] != targetArray[i]{
						return false, fmt.Errorf("%dth element in string array with key %d unmatched, %s => %s", i, key, sourceArray[i], targetArray[i])
					}
				}
				if t != nil{
					t.Logf("string array %d: [%s]", key, strings.Join(targetArray, ","))
				}
			}
		}
	}

	return true, nil
}

