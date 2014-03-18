package main

import (
	"fmt"
	"reflect"
	"strconv"
)

type BankErr struct {
	key string
	dsp string
}

func (err BankErr) Error() string {
	return fmt.Sprintf("%s:%s", err.key, err.dsp)
}

type auth_send struct {
	Msg     string `num:"0"  fmt:"fix"    en:"bcdl"  len:"4"`
	ProNum  string `num:"3"  fmt:"fix"    en:"bcdl"  len:"6"`
	SpcCode string `num:"25" fmt:"fix"    en:"bcdl"  len:"2"`
	TermiId string `num:"41" fmt:"fix"    en:"ascii" len:"8"`
	MerId   string `num:"42" fmt:"fix"    en:"ascii" len:"15"`
	OperId  string `num:"60" fmt:"lllvar" en:"hex"   len:"3"`
	SafeArg []byte `num:"61" fmt:"lllvar" en:"hex"   len:"999" lll:"bcdr"`
}

type tag map[string]string

func Marshal(o interface{}) ([]byte, error) {
	var (
	//bitmap [8]byte
	)
	isomap := make(map[string][]byte)
	typ := reflect.TypeOf(o)
	val := reflect.ValueOf(o)

	for i := 0; i < typ.NumField(); i++ {
		tags, err := structFieldTags(typ.Field(i))
		if err != nil {
			return nil, err
		}
		data := val.Field(i)
		fmt.Println("tag:", tags, "value:", data)
		isomap[tags["num"]], err = buildDataByTag(data, tags)
		if err != nil {
			return nil, BankErr{typ.Field(i).Name, err.Error()}
		}
	}
	fmt.Printf("%v\n", isomap)

	return nil, nil
}

func structFieldTags(filed reflect.StructField) (tag, error) {
	tags := make(map[string]string)

	tags["num"] = filed.Tag.Get("num")
	tags["fmt"] = filed.Tag.Get("fmt")
	tags["en"] = filed.Tag.Get("en")
	tags["len"] = filed.Tag.Get("len")
	tags["ll"] = filed.Tag.Get("ll")
	tags["lll"] = filed.Tag.Get("lll")
	return tags, nil
}
func buildDataByTag(data reflect.Value, tags tag) ([]byte, error) {
	switch data.Kind() {
	case reflect.String:
		s, err := buildStringByTag(data.String(), tags)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case reflect.Slice:
		s, err := buildSliceByTag(data.Bytes(), tags)
		if err != nil {
			return nil, err
		}
		return s, nil
	default:
		s := fmt.Sprintf("[%s]-无效的数据格式字段", data.Kind().String())
		return nil, BankErr{"1", s}
	}
}

func buildStringByTag(s string, tags tag) ([]byte, error) {
	var (
		out []byte
		err error
		fmt string
		ok  bool
	)

	if fmt, ok = tags["fmt"]; !ok {
		return nil, BankErr{"2", "没有指定fmt标签"}
	}
	switch fmt {
	case "fix":
		out, err = buildFix([]byte(s), tags)
	case "llvar":
	case "lllvar":
	default:
		return nil, BankErr{"3", "无效的fmt标签值"}
	}
	return out, err
}

func buildFix(s []byte, tags tag) ([]byte, error) {
	var (
		out      []byte
		err      error
		encoding string
		ok       bool
	)
	if encoding, ok = tags["en"]; !ok {
		return nil, BankErr{"4", "没有指定编码属性"}
	}
	switch encoding {
	case "ascii":
		out, err = buildAscii(s, tags)
	case "bcdl":
	case "bcdr":
	case "hex":
	default:
		return nil, BankErr{"5", "无效的编码属性"}
	}
	return out, err
}

func buildAscii(s []byte, tags tag) ([]byte, error) {
	l, ok := tags["len"]
	if !ok {
		return nil, BankErr{"6", "没有指定长度标签"}
	}
	if l == "" {
		return nil, BankErr{"7", "无效的长度标签"}
	}
	length, err := strconv.Atoi(l)
	if err != nil {
		fmt.Println(err)
		return nil, BankErr{"8", "无效的长度标签"}
	}
	if length != len(s) {
		return nil, BankErr{"9", "长度不符合标签定义"}
	}
	return s, nil
}
func buildSliceByTag(s []byte, tags tag) ([]byte, error) {
	return s, nil
}
func main() {
	auth := auth_send{"ddd", "343", "454", "12345678", "999999999911111", "tgt", []byte("gggg")}
	_, err := Marshal(auth)
	if err != nil {
		fmt.Println(err)
	}
}
