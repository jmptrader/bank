package main

import (
	"fmt"
	"reflect"
)

type BankErr struct {
	code int
	dsp  string
}

func (err BankErr) Error() string {
	return fmt.Sprintf("%d:%s", err.code, err.dsp)
}

type auth_send struct {
	Msg     string `num:"0"  fmt:"fix"    en:"bcdl"  len:"n      4"`
	ProNum  string `num:"3"  fmt:"fix"    en:"bcdl"  len:"n      6"`
	SpcCode string `num:"25" fmt:"fix"    en:"bcdl"  len:"n      2"`
	TermiId string `num:"41" fmt:"fix"    en:"ascii" len:"n      8"`
	MerId   string `num:"42" fmt:"fix"    en:"ascii" len:"n     15"`
	OperId  string `num:"60" fmt:"lllvar" en:"hex"   len:"h      3"`
	SafeArg []byte `num:"61" fmt:"lllvar" en:"hex"   len:"h .. 999" lll:"bcdr"`
}

func Marshal(o interface{}) ([]byte, error) {
	var (
		//bitmap [8]byte
		isomap map[string][]byte
	)
	typ := reflect.TypeOf(o)
	val := reflect.ValueOf(o)

	for i := 0; i < typ.NumField(); i++ {
		tags, err := structFieldTags(typ.Field(i))
		data := val.Field(i)
		if err != nil {
			return nil, err
		}
		fmt.Println("tag:", tags, "value:", data)
		isomap[tags["num"]], err = buildDataByTag(data, tags)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func structFieldTags(filed reflect.StructField) (map[string]string, error) {
	tags := make(map[string]string)

	tags["num"] = filed.Tag.Get("num")
	tags["fmt"] = filed.Tag.Get("fmt")
	tags["en"] = filed.Tag.Get("en")
	tags["len"] = filed.Tag.Get("len")
	tags["ll"] = filed.Tag.Get("ll")
	tags["lll"] = filed.Tag.Get("lll")
	return tags, nil
}
func buildDataByTag(data reflect.Value, tags map[string]string) ([]byte, error) {
	switch data.Kind() {
	case reflect.String:
	case reflect.Slice:
	default:
		s := fmt.Sprintf("[%s]-无效的数据格式字段", data.Kind().String())
		return nil, BankErr{1, s}
	}
}

func main() {
	auth := auth_send{"ddd", "343", "454", "rfr", "rgrg", "tgt", []byte("gggg")}
	Marshal(auth)
}
