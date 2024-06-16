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

	case VMTypeNull:
		return json.Marshal(struct {
			TypeId VMValueType `json:"t"`
		}{v.TypeId})

	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		x := struct {
			TypeId VMValueType `json:"t"`
			Value  struct {
				Expr  string          `json:"expr"`
				Attrs json.RawMessage `json:"attrs,omitempty"`
			} `json:"v"`
		}{}
		x.TypeId = v.TypeId
		x.Value.Expr = cd.Expr
		if cd.Attrs != nil {
			attrJson, err := cd.Attrs.ToJSON()
			if err != nil {
				return nil, err
			}
			x.Value.Attrs = attrJson
		}
		return json.Marshal(x)

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

		lst2 := [][]byte{[]byte(`{"t":6,"v":{"list":[`)}
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

		dictJson, err := cd.Dict.ToJSON()
		if err != nil {
			return nil, err
		}

		lst2 := [][]byte{[]byte(`{"t":7,"v":{"dict":`)}
		lst2 = append(lst2, dictJson)
		lst2 = append(lst2, []byte("}}"))

		return bytes.Join(lst2, []byte("")), nil

	case VMTypeFunction:
		cd, _ := v.ReadFunctionData()
		return json.Marshal(struct {
			TypeId VMValueType `json:"t"`
			Value  struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			} `json:"v"`
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
			TypeId VMValueType `json:"t"`
			Value  struct {
				Name string `json:"name"`
			} `json:"v"`
		}{
			v.TypeId,
			struct {
				Name string `json:"name"`
			}{fd.Name},
		})
	case VMTypeNativeObject:
		fd, _ := v.ReadNativeObjectData()
		return json.Marshal(struct {
			TypeId VMValueType `json:"t"`
			Value  struct {
				Name string `json:"name"`
			} `json:"v"`
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
		TypeId VMValueType `json:"t"`
	}

	err := json.Unmarshal(input, &v0)
	if err != nil {
		return err
	}
	v.TypeId = v0.TypeId

	switch v0.TypeId {
	case VMTypeInt:
		var v1 struct {
			Value IntType `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			// 这里浪费了一点性能，但是流程的一致性会更好
			v.Value = NewIntVal(v1.Value).Value
		}
		return err
	case VMTypeFloat:
		var v1 struct {
			Value float64 `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = NewFloatVal(v1.Value).Value
		}
		return err
	case VMTypeString:
		var v1 struct {
			Value string `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = NewStrVal(v1.Value).Value
		}
		return err
	case VMTypeNull:
		return nil
	case VMTypeComputedValue:
		var v1 struct {
			Value struct {
				Expr  string          `json:"expr"`
				Attrs json.RawMessage `json:"attrs,omitempty"`
			} `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			cd := &ComputedData{
				Expr: v1.Value.Expr,
			}
			if v1.Value.Attrs != nil {
				cd.Attrs = &ValueMap{}
				if err := json.Unmarshal(v1.Value.Attrs, cd.Attrs); err != nil {
					return err
				}
			}
			v.Value = cd
		}
		return err
	case VMTypeArray:
		var v1 struct {
			Value struct {
				List []*VMValue `json:"list"`
			} `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			v.Value = NewArrayValRaw(v1.Value.List).Value
		}
		return err
	case VMTypeDict:
		var v1 struct {
			Value struct {
				Dict ValueMap `json:"dict"`
			} `json:"v"`
		}

		if err := json.Unmarshal(input, &v1); err != nil {
			return err
		}
		v.Value = NewDictVal(&v1.Value.Dict).Value
		return nil

	case VMTypeFunction:
		var v1 struct {
			Value struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			} `json:"v"`
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
			} `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			if val, ok := builtinValues[v1.Value.Name]; ok {
				v.Value = val.Value
			}
			return nil
		}
		return err
	case VMTypeNativeObject:
		var v1 struct {
			Value struct {
				Name string `json:"name"`
			} `json:"v"`
		}
		err := json.Unmarshal(input, &v1)
		if err == nil {
			od := &NativeObjectData{Name: v1.Value.Name}
			// 只能创建一个空壳，也许反序列化时跳过会更好
			v.Value = NewNativeObjectVal(od).Value
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
