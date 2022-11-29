package dicescript

import (
	"bytes"
	"encoding/json"
	"errors"
)

func (v *VMValue) ToJSONRaw(save map[*VMValue]bool) ([]byte, error) {
	if v == nil {
		return nil, errors.New("nil pointer")
	}
	switch v.TypeId {
	case VMTypeInt:
		fallthrough
	case VMTypeFloat:
		fallthrough
	case VMTypeString:
		return json.Marshal(v)

	case VMTypeUndefined:
		fallthrough
	case VMTypeNull:
		return json.Marshal(struct {
			TypeId VMValueType `json:"typeId"`
		}{v.TypeId})

	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		return json.Marshal(struct {
			TypeId VMValueType `json:"typeId"`
			Value  struct {
				Expr string `json:"expr"`
			} `json:"value"`
		}{
			v.TypeId,
			struct {
				Expr string `json:"expr"`
			}{cd.Expr},
		})

	case VMTypeArray:
		if save == nil {
			save = map[*VMValue]bool{}
		}
		if _, exists := save[v]; exists {
			return nil, errors.New("值错误: 序列化时检测到循环引用")
		}
		save[v] = true
		ad, _ := v.ReadArray()
		lst := [][]byte{}
		for _, i := range ad.List {
			json_data, err := i.ToJSONRaw(save)
			if err != nil {
				return nil, err
			}
			lst = append(lst, json_data)
		}

		lst2 := [][]byte{[]byte(`{"typeId":6,"value":{"list":[`)}
		lst2 = append(lst2, bytes.Join(lst, []byte(",")))
		lst2 = append(lst2, []byte("]}}"))

		return bytes.Join(lst2, []byte("")), nil

	case VMTypeDict:
		if save == nil {
			save = map[*VMValue]bool{}
		}
		if _, exists := save[v]; exists {
			return nil, errors.New("值错误: 序列化时检测到循环引用")
		}
		save[v] = true
		cd := v.MustReadDictData()

		lst := [][]byte{}
		var err error
		cd.Dict.Range(func(key string, value *VMValue) bool {
			var jsonKey []byte
			var jsonData []byte
			jsonData, err = value.ToJSONRaw(save)
			if err != nil {
				return true
			}
			jsonKey, err = json.Marshal(key)
			if err != nil {
				return true
			}

			b := append(jsonKey, []byte(":")...)
			b = append(b, jsonData...)

			lst = append(lst, b)
			return false
		})
		if err != nil {
			return nil, err
		}

		lst2 := [][]byte{[]byte(`{"typeId":7,"value":{"dict":{`)}
		lst2 = append(lst2, bytes.Join(lst, []byte(",")))
		lst2 = append(lst2, []byte("}}}"))

		return bytes.Join(lst2, []byte("")), nil

	case VMTypeFunction:
		cd, _ := v.ReadFunctionData()
		return json.Marshal(struct {
			TypeId VMValueType `json:"typeId"`
			Value  struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			} `json:"value"`
		}{
			v.TypeId,
			struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			}{cd.Expr, cd.Name, cd.Params},
		})

	case VMTypeNativeFunction:
		fd, _ := v.ReadNativeFunctionData()
		return json.Marshal(struct {
			TypeId VMValueType `json:"typeId"`
			Value  struct {
				Name string `json:"name"`
			} `json:"value"`
		}{
			v.TypeId,
			struct {
				Name string `json:"name"`
			}{fd.Name},
		})
	}
	return nil, nil
}

func (v *VMValue) ToJSON() ([]byte, error) {
	return v.ToJSONRaw(nil)
}

func (v *VMValue) UnmarshalJSON(input []byte) error {
	var v0 struct {
		TypeId VMValueType `json:"typeId"`
	}

	err := json.Unmarshal(input, &v0)
	if err != nil {
		return err
	}
	v.TypeId = v0.TypeId

	switch v0.TypeId {
	case VMTypeInt:
		var v1 struct {
			Value int64 `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			// 这里浪费了一点性能，但是流程的一致性会更好
			v.Value = VMValueNewInt(v1.Value).Value
		}
		return err
	case VMTypeFloat:
		var v1 struct {
			Value float64 `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = VMValueNewFloat(v1.Value).Value
		}
		return err
	case VMTypeString:
		var v1 struct {
			Value string `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = VMValueNewStr(v1.Value).Value
		}
		return err
	case VMTypeUndefined:
		return nil
	case VMTypeNull:
		return nil
	case VMTypeComputedValue:
		var v1 struct {
			Value struct {
				Expr string `json:"expr"`
			} `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = VMValueNewComputed(v1.Value.Expr).Value
		}
		return err
	case VMTypeArray:
		var v1 struct {
			Value struct {
				List []*VMValue `json:"list"`
			} `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = VMValueNewArrayRaw(v1.Value.List).Value
		}
		return err
	case VMTypeDict:
		var v1 struct {
			Value struct {
				Dict map[string]*VMValue `json:"dict"`
			} `json:"value"`
		}

		err := json.Unmarshal(input, &v1)
		if err == nil {
			m := &ValueMap{}
			for k, v := range v1.Value.Dict {
				m.Store(k, v)
			}
			v.Value = VMValueNewDict(m).Value
		}
		return err

	case VMTypeFunction:
		var v1 struct {
			Value struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			} `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			fd := &FunctionData{Expr: v1.Value.Expr, Name: v1.Value.Name, Params: v1.Value.Params}
			v.Value = fd
			return nil
		}
		return err

	case VMTypeNativeFunction:
		var v1 struct {
			Value struct {
				Name string `json:"name"`
			} `json:"value"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			if val, ok := builtinValues[v1.Value.Name]; ok {
				v.Value = val.Value
			}
			return nil
		}
		return err
	}
	return nil
}

func VMValueFromJSON(data []byte) (*VMValue, error) {
	var v VMValue
	err := json.Unmarshal(data, &v)
	return &v, err
}
