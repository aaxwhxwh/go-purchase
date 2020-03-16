package utils

import (
	"encoding/json"
	"errors"
	"reflect"
)

func ToString(sli interface{}) string {
	s, _ := json.Marshal(sli)
	return string(s)
}

func StringToSliceString(str string) []string {
	var sli []string
	json.Unmarshal([]byte(str), &sli)
	return sli
}

func StringToSliceMap(str string) (sli []map[string]string) {
	json.Unmarshal([]byte(str), &sli)
	return
}

func StringToSliceInt(str string) []int {
	var sli []int
	json.Unmarshal([]byte(str), &sli)
	return sli
}

func StringToMap(str string) (m map[string]interface{}){
	json.Unmarshal([]byte(str), &m)
	return
}

// 判断元素是否在切片或数组内
func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}
	return false, errors.New("not in array")
}

// @Title 结构体相同字段值复制
func TransferToStruct(src, dst interface{}) (err error){
	d, err := json.Marshal(src)
	if err != nil {
		return
	}
	err = json.Unmarshal(d, dst)
	return
}

// 结构体copy
func CopyStruct(src, dst interface{}) {
	sVal := reflect.ValueOf(src).Elem()
	dVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < sVal.NumField(); i++ {
		value := sVal.Field(i)
		name := sVal.Type().Field(i).Name

		dValue := dVal.FieldByName(name)
		if dValue.IsValid() == false {
			continue
		}
		dValue.Set(value) //这里默认共同成员的类型一样，否则这个地方可能导致 panic，需要简单修改一下。
	}
}