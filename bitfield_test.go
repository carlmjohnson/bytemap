package bytemap_test

import (
	"io"
	"strings"
	"testing"

	"github.com/carlmjohnson/bytemap"
	"golang.org/x/exp/maps"
)

func FuzzMakeBitField(f *testing.F) {
	f.Add("", "")
	f.Add("a", "a")
	f.Add("a", "b")
	f.Add("ab", "ab")
	f.Add("ab", "abc")
	for i := 0; i < 1_000_000; i = (i + 1) * 2 {
		for j := 0; j < 3; j++ {
			s := strings.Repeat("a", i)
			charset := strings.Repeat("a", j)
			f.Add(s, charset)
			f.Add(s+"b", charset)
		}
	}
	f.Fuzz(func(t *testing.T, s, charset string) {
		want := naiveContains(s, charset)
		t.Run("Make", func(t *testing.T) {
			m := bytemap.Make(charset).ToBitField()
			testContainment(t, m, s, charset, want)
		})
		t.Run("WriteString", func(t *testing.T) {
			m := &bytemap.BitField{}
			n, err := m.WriteString(charset)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(charset) {
				t.Fatal(len(charset))
			}
			testContainment(t, m, s, charset, want)
		})
		t.Run("Write", func(t *testing.T) {
			m := &bytemap.BitField{}
			n, err := m.Write([]byte(charset))
			if err != nil {
				t.Fatal(err)
			}
			if n != len(charset) {
				t.Fatal(len(charset))
			}
			testContainment(t, m, s, charset, want)
		})
		// Test io.Copy
		t.Run("Copy", func(t *testing.T) {
			m := &bytemap.BitField{}
			n64, err := io.Copy(m, strings.NewReader(charset))
			if err != nil {
				t.Fatal(err)
			}
			if n64 != int64(len(charset)) {
				t.Fatal(len(charset))
			}
			testContainment(t, m, s, charset, want)
		})
	})
}

func FuzzBitFieldToMap(f *testing.F) {
	f.Add("", "")
	f.Add("a", "b")
	f.Add(
		"the quick brown fox jumps over a lazy dog.",
		"abcdefghijklmnopqrstuvwxyz. ",
	)
	f.Fuzz(func(t *testing.T, a, b string) {
		aNaive := naiveMap(a)
		aMap := bytemap.Make(a).ToBitField()
		if !maps.Equal(aNaive, aMap.ToMap()) {
			t.Fatalf("input=%q want=%v got=%v",
				a, aNaive, aMap.ToMap())
		}
		testGet(t, aMap, aNaive)
		bNaive := naiveMap(b)
		bMap := bytemap.Make(b).ToBitField()
		if !maps.Equal(bNaive, bMap.ToMap()) {
			t.Fatal(b, bMap)
		}
		testGet(t, bMap, bNaive)
		if maps.Equal(aNaive, bNaive) != aMap.Equals(bMap) {
			t.Fatal(aMap, bMap)
		}
	})
}

func FuzzBitFieldSet(f *testing.F) {
	f.Add("", "", "")
	f.Add("a", "a", "a")
	f.Add("abc", "bcde", "b")
	f.Fuzz(func(t *testing.T, add, remove, restore string) {
		var bf bytemap.BitField
		m := make(map[byte]bool)
		for _, c := range []byte(add) {
			bf.Set(c, true)
			m[c] = true
		}
		for _, c := range []byte(remove) {
			bf.Set(c, false)
			m[c] = false
		}
		for _, c := range []byte(restore) {
			bf.Set(c, true)
			m[c] = true
		}
		// Fill in blanks
		for i := 0; i < bytemap.Len; i++ {
			m[byte(i)] = m[byte(i)]
		}
		if !maps.Equal(bf.ToMap(), m) {
			t.Fatal(bf)
		}
	})
}
