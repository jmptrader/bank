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
	Msg     string `num:"0"    fmt:"fix"    en:"bcdl"  len:"4"`
	ProNum  string `num:"3"    fmt:"fix"    en:"bcdr"  len:"7"`
	SpcCode string `num:"25"   fmt:"fix"    en:"bcdl"  len:"2"`
	TermiId string `num:"41"   fmt:"llvar"  en:"bcdl" len:"35"  ll:"bcdr"`
	MerId   string `num:"42"   fmt:"llvar"  en:"ascii" len:"15"  ll:"bcdr"`
	OperId  string `num:"60"   fmt:"llvar"  en:"ascii" len:"5"   ll:"hex"`
	SafeArg []byte `num:"61"   fmt:"lllvar"  en:"bcdr" len:"99" lll:"bcdl"`
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
	//fmt.Printf("%v\n", isomap)
	for k, v := range isomap {
		fmt.Printf("%s:%x\n", k, v)
	}

	return nil, nil
}

func structFieldTags(filed reflect.StructField) (tag, error) {
	tags := make(map[string]string)

	val := filed.Tag.Get("num")
	if val != "" {
		tags["num"] = val
	}
	val = filed.Tag.Get("fmt")
	if val != "" {
		tags["fmt"] = val
	}
	val = filed.Tag.Get("en")
	if val != "" {
		tags["en"] = val
	}
	val = filed.Tag.Get("len")
	if val != "" {
		tags["len"] = val
	}
	val = filed.Tag.Get("ll")
	if val != "" {
		tags["ll"] = val
	}
	val = filed.Tag.Get("lll")
	if val != "" {
		tags["lll"] = val
	}
	return tags, nil
}
func buildDataByTag(data reflect.Value, tags tag) ([]byte, error) {
	switch data.Kind() {
	case reflect.String:
		s, err := buildSliceByTag([]byte(data.String()), tags)
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

func buildLlvar(s []byte, tags tag) ([]byte, error) {
	var (
		out      []byte
		err      error
		encoding string
	)

	//llvar of var
	if encoding, _ = tags["en"]; encoding == "" {
		return nil, BankErr{"-14", "没有指定编码属性"}
	}
	l := 0
	if l, err = strconv.Atoi(tags["len"]); err != nil {
		return nil, BankErr{"-15", "无效的长度标签"}
	}
	if len(s) > l {
		return nil, BankErr{"-16", "长度不符合标签定义"}
	}

	tags["len"] = strconv.Itoa(len(s))
	if out, err = buildFix(s, tags); err != nil {
		return nil, err
	}
	//llvar of ll
	ll := ""
	if ll, _ = tags["ll"]; ll == "" {
		return nil, BankErr{"-17", "ll-没有指定编码属性"}
	}
	out_last := make([]byte, 0)
	switch ll {
	case "ascii":
		out_last = append(out_last, fmt.Sprintf("%02d", len(s))...)
		out_last = append(out_last, out...)
	case "bcdl":
		olen := fmt.Sprintf("%d", len(s))
		if len(olen) == 1 {
			olen += "0"
		}
		ll_store, _ := buildBcdl([]byte(olen), tag{"len": "2"})
		out_last = append(out_last, ll_store...)
		out_last = append(out_last, out...)
	case "bcdr":
		olen := fmt.Sprintf("%02d", len(s))
		ll_store, _ := buildBcdr([]byte(olen), tag{"len": "2"})
		out_last = append(out_last, ll_store...)
		out_last = append(out_last, out...)
	case "hex":
		var ll_store [2]byte
		ll_store[0] = byte(len(s) / 256)
		ll_store[1] = byte(len(s) % 256)
		out_last = append(out_last, ll_store[0])
		out_last = append(out_last, ll_store[1])
		out_last = append(out_last, out...)
	default:
		return nil, BankErr{"-18", "ll-无效的编码属性"}
	}

	return out_last, err
}
func buildLllvar(s []byte, tags tag) ([]byte, error) {
	var (
		out      []byte
		err      error
		encoding string
	)

	//lllvar of var
	if encoding, _ = tags["en"]; encoding == "" {
		return nil, BankErr{"-20", "没有指定编码属性"}
	}
	l := 0
	if l, err = strconv.Atoi(tags["len"]); err != nil {
		return nil, BankErr{"-21", "无效的长度标签"}
	}
	if len(s) > l {
		return nil, BankErr{"-22", "长度不符合标签定义"}
	}

	tags["len"] = strconv.Itoa(len(s))
	if out, err = buildFix(s, tags); err != nil {
		return nil, err
	}
	//lllvar of ll
	lll := ""
	if lll, _ = tags["lll"]; lll == "" {
		return nil, BankErr{"-23", "lll-没有指定编码属性"}
	}
	out_last := make([]byte, 0)
	switch lll {
	case "ascii":
		out_last = append(out_last, fmt.Sprintf("%03d", len(s))...)
		out_last = append(out_last, out...)
	case "bcdl":
		olen := fmt.Sprintf("%d", len(s))
		if len(olen) == 1 {
			olen += "00"
		}
		if len(olen) == 2 {
			olen += "0"
		}
		lll_store, _ := buildBcdl([]byte(olen), tag{"len": "3"})
		out_last = append(out_last, lll_store...)
		out_last = append(out_last, out...)
	case "bcdr":
		olen := fmt.Sprintf("%03d", len(s))
		lll_store, _ := buildBcdr([]byte(olen), tag{"len": "3"})
		out_last = append(out_last, lll_store...)
		out_last = append(out_last, out...)
	default:
		return nil, BankErr{"-24", "lll-无效的编码属性"}
	}

	return out_last, err
}
func buildFix(s []byte, tags tag) ([]byte, error) {
	var (
		out []byte
		err error
	)
	encoding, _ := tags["en"]
	if encoding == "" {
		return nil, BankErr{"-4", "没有指定编码属性"}
	}
	switch encoding {
	case "ascii":
		out, err = buildAscii(s, tags)
	case "bcdl":
		out, err = buildBcdl(s, tags)
	case "bcdr":
		out, err = buildBcdr(s, tags)
	case "hex":
		buildHex := buildAscii
		out, err = buildHex(s, tags)
	default:
		return nil, BankErr{"-5", "无效的编码属性"}
	}
	return out, err
}

func buildAscii(s []byte, tags tag) ([]byte, error) {
	l, _ := tags["len"]
	length, err := strconv.Atoi(l)
	if err != nil {
		return nil, BankErr{"-8", "无效的长度标签"}
	}
	if length != len(s) {
		return nil, BankErr{"-9", "长度不符合标签定义"}
	}
	return s, nil
}
func buildBcdl(s []byte, tags tag) ([]byte, error) {
	l, _ := tags["len"]
	length, err := strconv.Atoi(l)
	if err != nil {
		return nil, BankErr{"-10", "无效的长度标签"}
	}
	if length != len(s) {
		return nil, BankErr{"-11", "长度不符合标签定义"}
	}
	tmp := make([]byte, 0)
	tmp = append(tmp, s...)
	if len(tmp)%2 != 0 {
		tmp = append(tmp, "\x00"...)
	}
	return toBcd(tmp), nil
}

func buildBcdr(s []byte, tags tag) ([]byte, error) {
	l, _ := tags["len"]
	length, err := strconv.Atoi(l)
	if err != nil {
		return nil, BankErr{"-12", "无效的长度标签"}
	}
	if length != len(s) {
		return nil, BankErr{"-13", "长度不符合标签定义"}
	}
	tmp := make([]byte, 0)
	if len(s)%2 != 0 {
		tmp = append(tmp, "\x00"...)
	}
	tmp = append(tmp, s...)
	return toBcd(tmp), nil
}
func toBcd(s []byte) []byte {
	out := make([]byte, len(s)/2)
	k := 0
	for i := 0; i < len(out); i++ {
		out[i] = ((s[k] & 0x0f) << 4) | (s[k+1] & 0x0f)
		k += 2
	}
	return out
}

func buildSliceByTag(s []byte, tags tag) ([]byte, error) {
	var (
		out []byte
		err error
		fmt string
	)
	if fmt, _ = tags["fmt"]; fmt == "" {
		return nil, BankErr{"-19", "没有指定fmt标签"}
	}
	switch fmt {
	case "fix":
		out, err = buildFix(s, tags)
	case "llvar":
		out, err = buildLlvar(s, tags)
	case "lllvar":
		out, err = buildLllvar(s, tags)
	default:
		return nil, BankErr{"-3", "无效的fmt标签值"}
	}
	return out, err
}
func main() {
	auth := auth_send{"0820",
		"1234567",
		"99",
		"1234567",
		"999999999911111",
		"\xA0\x01\x01",
		[]byte("1234567")}
	_, err := Marshal(auth)
	if err != nil {
		fmt.Println(err)
	}
}
