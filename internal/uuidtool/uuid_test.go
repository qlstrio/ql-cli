package uuidTool

import (
	"strings"
	"testing"
)

func TestIsUUID(t *testing.T) {
	valid := []string{
		"0d6ad000-bc92-4518-8d1c-044763bfc9b4",
		"0D6AD000-BC92-4518-8D1C-044763BFC9B4",
		"ffffffff-ffff-4fff-8fff-ffffffffffff",
		"00000000-0000-4000-8000-000000000000",
	}
	for _, u := range valid {
		if !isUUID(u) {
			t.Errorf("应该是合法 UUID: %s", u)
		}
	}

	invalid := []string{
		"",
		"not-a-uuid",
		"0d6ad000bc9245188d1c044763bfc9b4",
		"0d6ad000-bc92-3518-8d1c-044763bfc9b4",
		"0d6ad000-bc92-4518-0d1c-044763bfc9b4",
		"0d6ad000-bc92-4518-8d1c",
		"0d6ad000-bc92-4518-8d1c-044763bfc9b4extra",
	}
	for _, u := range invalid {
		if isUUID(u) {
			t.Errorf("不应该判定为合法 UUID: %q", u)
		}
	}
}

func TestHexCharValue(t *testing.T) {
	cases := map[byte]int{
		'0': 0, '1': 1, '9': 9,
		'a': 10, 'f': 15,
		'A': 10, 'F': 15,
	}
	for c, want := range cases {
		if got := hexCharValue(c); got != want {
			t.Errorf("hexCharValue(%q) = %d, want %d", c, got, want)
		}
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("非法十六进制字符应该 panic")
		}
	}()
	hexCharValue('g')
}

func TestHexByte(t *testing.T) {
	want := "0123456789abcdef"
	for i := 0; i < 16; i++ {
		if got := hexByte(i); got != want[i] {
			t.Errorf("hexByte(%d) = %c, want %c", i, got, want[i])
		}
	}
}

func TestCompress_RoundTrip_23(t *testing.T) {
	uuids := []string{
		"0d6ad000-bc92-4518-8d1c-044763bfc9b4",
		"ffffffff-ffff-4fff-bfff-ffffffffffff",
		"00000000-0000-4000-8000-000000000000",
	}

	for _, u := range uuids {
		short = false
		compressed := compress(u)
		if len(compressed) != 23 {
			t.Errorf("压缩后长度应为 23, 实际 %d (%q)", len(compressed), compressed)
		}

		got := captureDecompress(compressed)
		if got != strings.ToLower(u) {
			t.Errorf("Round-trip 不匹配:\n  原:  %s\n  解压: %s", u, got)
		}
	}
}

func TestCompress_RoundTrip_22(t *testing.T) {
	uuids := []string{
		"0d6ad000-bc92-4518-8d1c-044763bfc9b4",
		"ffffffff-ffff-4fff-bfff-ffffffffffff",
	}

	for _, u := range uuids {
		short = true
		compressed := compress(u)
		if len(compressed) != 22 {
			t.Errorf("压缩后长度应为 22, 实际 %d (%q)", len(compressed), compressed)
		}

		got := captureDecompress(compressed)
		if got != strings.ToLower(u) {
			t.Errorf("Round-trip 不匹配:\n  原:  %s\n  解压: %s", u, got)
		}
	}
	short = false
}

func TestCompress_Deterministic(t *testing.T) {
	u := "0d6ad000-bc92-4518-8d1c-044763bfc9b4"
	short = false
	a := compress(u)
	b := compress(u)
	if a != b {
		t.Errorf("相同输入应产出相同结果: %q vs %q", a, b)
	}
}

func captureDecompress(s string) string {
	if len(s) == 23 {
		var AsciiTo64 [128]int
		for i := range 64 {
			AsciiTo64[ase64KeyChars[i]] = i
		}
		hexChars := make([]byte, 0, 27)
		for i := 5; i < 23; i += 2 {
			lhs := AsciiTo64[s[i]]
			rhs := AsciiTo64[s[i+1]]
			hexChars = append(hexChars, hexByte(lhs>>2))
			hexChars = append(hexChars, hexByte(((lhs&3)<<2)|(rhs>>4)))
			hexChars = append(hexChars, hexByte(rhs&0xF))
		}
		s = s[0:5] + string(hexChars)
	} else if len(s) == 22 {
		var AsciiTo64 [128]int
		for i := range 64 {
			AsciiTo64[ase64KeyChars[i]] = i
		}
		hexChars := make([]byte, 0, 30)
		for i := 2; i < 22; i += 2 {
			lhs := AsciiTo64[s[i]]
			rhs := AsciiTo64[s[i+1]]
			hexChars = append(hexChars, hexByte(lhs>>2))
			hexChars = append(hexChars, hexByte(((lhs&3)<<2)|(rhs>>4)))
			hexChars = append(hexChars, hexByte(rhs&0xF))
		}
		s = s[0:2] + string(hexChars)
	}
	return strings.Join([]string{
		s[0:8], s[8:12], s[12:16], s[16:20], s[20:],
	}, "-")
}
