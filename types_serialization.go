package dicescript

import (
	"encoding/json"
	"errors"
)

func (v *VMValue) ToJSON() ([]byte, error) {
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

		//case VMTypeArray:
		//	s := "["
		//	arr, _ := v.ReadArray()
		//	for index, i := range arr.List {
		//		if i.TypeId == VMTypeArray {
		//			s += "[...]"
		//		} else {
		//			s += i.ToString()
		//		}
		//		if index != len(arr.List)-1 {
		//			s += ", "
		//		}
		//	}
		//	s += "]"
		//	return s
		//case VMTypeNativeFunction:
		//	cd, _ := v.ReadNativeFunctionData()
		//	return "nfunction " + cd.Name
	}
	return nil, nil
}

func ValueFromJSON(data []byte) (*VMValue, error) {
	var v0 struct {
		TypeId VMValueType `json:"typeId"`
	}

	err := json.Unmarshal(data, &v0)
	if err != nil {
		return nil, err
	}

	switch v0.TypeId {
	case VMTypeInt:
		var v1 struct {
			Value int64 `json:"value"`
		}
		err := json.Unmarshal(data, &v1)
		if err == nil {
			return VMValueNewInt(v1.Value), nil
		}
		return nil, err
	case VMTypeFloat:
		var v1 struct {
			Value float64 `json:"value"`
		}
		err := json.Unmarshal(data, &v1)
		if err == nil {
			return VMValueNewFloat(v1.Value), nil
		}
		return nil, err
	case VMTypeString:
		var v1 struct {
			Value string `json:"value"`
		}
		err := json.Unmarshal(data, &v1)
		if err == nil {
			return VMValueNewStr(v1.Value), nil
		}
		return nil, err
	case VMTypeUndefined:
		return VMValueNewUndefined(), nil
	case VMTypeNull:
		return VMValueNewNull(), nil
	case VMTypeComputedValue:
		var v1 struct {
			Value struct {
				Expr string `json:"expr"`
			} `json:"value"`
		}
		err := json.Unmarshal(data, &v1)
		if err == nil {
			return VMValueNewComputed(v1.Value.Expr), nil
		}
		return nil, err
	case VMTypeFunction:
		var v1 struct {
			Value struct {
				Expr   string   `json:"expr"`
				Name   string   `json:"name"`
				Params []string `json:"params"`
			} `json:"value"`
		}
		err := json.Unmarshal(data, &v1)
		if err == nil {
			fd := &FunctionData{Expr: v1.Value.Expr, Name: v1.Value.Name, Params: v1.Value.Params}
			return VMValueNewFunctionRaw(fd), nil
		}
		return nil, err
	}
	return nil, nil
}
