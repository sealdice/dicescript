package dicescript

func (d *VMDictValue) V() *VMValue {
	return (*VMValue)(d)
}

func (d *VMDictValue) Store(key string, value *VMValue) {
	if dd, ok := d.V().ReadDictData(); ok {
		dd.Dict.Store(key, value)
	}
}

func (d *VMDictValue) Load(key string) (value *VMValue, ok bool) {
	if dd, ok := d.V().ReadDictData(); ok {
		return dd.Dict.Load(key)
	}
	return nil, false
}

func (d *VMDictValue) ToString() string {
	return d.V().ToString()
}
