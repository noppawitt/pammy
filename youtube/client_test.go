package youtube

import (
	"os"
	"reflect"
	"testing"
)

func Test_ExtractSearchResult(t *testing.T) {
	f, _ := os.Open("../testdata/search-results.html")

	want := []VideoInfo{{ID: "sAf5zzY2EH8", Title: "เธอเก่ง(Still) - Jetset'er [Official MV]", LengthText: "6:51"}, {ID: "Rf4tE-H0tTE", Title: "เธอเก่ง (Still) - Jetsetter l Cover by ไอซ์", LengthText: "5:57"}, {ID: "rgMJ0vUOPJ4", Title: "เธอเก่ง  - Jetset'er Cover by แก้ม วิชญาณี", LengthText: "5:45"}, {ID: "ofWnWQpaITA", Title: "เธอเก่ง Jetset'er (audio)", LengthText: "4:57"}, {ID: "Sh9YXIodkbY", Title: "PORZAX - เธอเก่งที่สุดแล้ว [OFFICIAL AUDIO]", LengthText: "4:17"}, {ID: "BXgBcO_YOM0", Title: "เธอเก่ง - Jetset’er | กระแทกใจคนมูฟออนไม่ได้!! | Songtopia Livehouse", LengthText: "5:45"}, {ID: "Z6secasl1Cc", Title: "เธอเก่ง - หน้ากากโสนน้อยเรือนงาม  | THE MASK วรรณคดีไทย", LengthText: "6:34"}, {ID: "yachVFGlaMY", Title: "เธอเก่ง - Jetset'er", LengthText: "4:57"}, {ID: "0lcvK-4jPpU", Title: "เธอเก่ง : Gam Concert My First Time", LengthText: "8:03"}, {ID: "caaH5KRg2qo", Title: "เธอเก่ง (Still) - Jetset'er | Live Session", LengthText: "7:00"}, {ID: "0SC41qwUp3E", Title: "เธอเก่ง แดน วรเวช  - Cover Night Plus : When A Man In Love", LengthText: "5:43"}, {ID: "i7_HKbiPBOw", Title: "เธอเก่ง(Still) - Jetset'er + Rainy Mood", LengthText: "5:01"}, {ID: "7iSia7rb1PY", Title: "F.HERO x Tilly Birds (Prod. By Billy & Ohm Cocktail) - จำเก่ง (Slipped Your Mind) [Official MV]", LengthText: "3:44"}, {ID: "PY-jDenT1c0", Title: "[MAD] เธอเก่ง - Jetset'er (Cover) | Khaopoad (The Voice Thailand Season 4)", LengthText: "6:27"}, {ID: "8l2C4oUNYx0", Title: "สอน เธอเก่ง jetseter - ตีคอร์ด+intro แบบง่ายไม่มีคอร์ดทาบ สำหรับมือใหม่ KeyG - น้าจร เชียงใหม่ cover", LengthText: "6:12"}, {ID: "2ZvNA8r1vME", Title: "เธอเก่ง - Jetset'er | I Can See Your Voice -TH", LengthText: "2:41"}, {ID: "3PsUP7hrelk", Title: "เธอเก่ง - Jetset'er | Cover | SCA STUDIO | หลิน SCA (แชมป์ มาสเตอร์คีย์เวทีแจ้งเกิด) feat.ศร SCA", LengthText: "7:24"}, {ID: "9I_uO9AlVOo", Title: "เธอเก่ง (Still) - Jetsetter l Cover by เต้อ", LengthText: "5:24"}, {ID: "ZysJE2heuA8", Title: "เธอเก่ง - Jetset'er (1 ก.ค.61)", LengthText: "6:19"}}

	got, err := ExtractSearchResult(f)
	if err != nil {
		t.Error("extract error: ", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant: %+v\ngot %+v", want, got)
	}
}
