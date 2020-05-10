package database

import "testing"

func TestDna(t *testing.T) {
	var testData = []struct{
		userName string
		dna int
	} {
		{"yang", 431},
		{"ycs01", 432},
		{"ycs02", 433},
		{"cs01", 311},
		{"cs02", 312},
		{"yangcs01cs02", 54},
		{"1028990481@qq.com", 177},
		//{"", -1},
	}
	for _, data := range testData {
		dna, err := Dna(data.userName)
		if err != nil {
			t.Errorf(err.Error())
		}
		if dna != data.dna {
			t.Errorf("aspect dna is %d, but get %d", data.dna, dna)
		}
	}
}

func TestDnaFromNumber(t *testing.T) {
	var testData = []struct{
		userName int
		dna int
	} {
		{18922311000, 0},
		{18922311001, 1},
		{18922311100, 100},
		{18922311101, 101},
		{18922313542, 542},
	}
	for _, data := range testData {
		dna, err := dnaFromNumber(data.userName)
		if err != nil {
			t.Errorf(err.Error())
		}
		if dna != data.dna {
			t.Errorf("aspect dna is %d, but get %d", data.dna, dna)
		}
	}
}
