package youtube

import (
	"os"
	"reflect"
	"testing"
)

func Test_ExtractSearchResult(t *testing.T) {
	f, _ := os.Open("../testdata/search-results.html")

	want := []VideoInfo{{ID: "sAf5zzY2EH8", Title: "เธอเก่ง(Still) - Jetset'er [Official MV]", Duration: durationFromLengthText("6:51")}, {ID: "Rf4tE-H0tTE", Title: "เธอเก่ง (Still) - Jetsetter l Cover by ไอซ์", Duration: durationFromLengthText("5:57")}, {ID: "rgMJ0vUOPJ4", Title: "เธอเก่ง  - Jetset'er Cover by แก้ม วิชญาณี", Duration: durationFromLengthText("5:45")}, {ID: "ofWnWQpaITA", Title: "เธอเก่ง Jetset'er (audio)", Duration: durationFromLengthText("4:57")}, {ID: "Sh9YXIodkbY", Title: "PORZAX - เธอเก่งที่สุดแล้ว [OFFICIAL AUDIO]", Duration: durationFromLengthText("4:17")}, {ID: "BXgBcO_YOM0", Title: "เธอเก่ง - Jetset’er | กระแทกใจคนมูฟออนไม่ได้!! | Songtopia Livehouse", Duration: durationFromLengthText("5:45")}, {ID: "Z6secasl1Cc", Title: "เธอเก่ง - หน้ากากโสนน้อยเรือนงาม  | THE MASK วรรณคดีไทย", Duration: durationFromLengthText("6:34")}, {ID: "yachVFGlaMY", Title: "เธอเก่ง - Jetset'er", Duration: durationFromLengthText("4:57")}, {ID: "0lcvK-4jPpU", Title: "เธอเก่ง : Gam Concert My First Time", Duration: durationFromLengthText("8:03")}, {ID: "caaH5KRg2qo", Title: "เธอเก่ง (Still) - Jetset'er | Live Session", Duration: durationFromLengthText("7:00")}, {ID: "0SC41qwUp3E", Title: "เธอเก่ง แดน วรเวช  - Cover Night Plus : When A Man In Love", Duration: durationFromLengthText("5:43")}, {ID: "i7_HKbiPBOw", Title: "เธอเก่ง(Still) - Jetset'er + Rainy Mood", Duration: durationFromLengthText("5:01")}, {ID: "7iSia7rb1PY", Title: "F.HERO x Tilly Birds (Prod. By Billy & Ohm Cocktail) - จำเก่ง (Slipped Your Mind) [Official MV]", Duration: durationFromLengthText("3:44")}, {ID: "PY-jDenT1c0", Title: "[MAD] เธอเก่ง - Jetset'er (Cover) | Khaopoad (The Voice Thailand Season 4)", Duration: durationFromLengthText("6:27")}, {ID: "8l2C4oUNYx0", Title: "สอน เธอเก่ง jetseter - ตีคอร์ด+intro แบบง่ายไม่มีคอร์ดทาบ สำหรับมือใหม่ KeyG - น้าจร เชียงใหม่ cover", Duration: durationFromLengthText("6:12")}, {ID: "2ZvNA8r1vME", Title: "เธอเก่ง - Jetset'er | I Can See Your Voice -TH", Duration: durationFromLengthText("2:41")}, {ID: "3PsUP7hrelk", Title: "เธอเก่ง - Jetset'er | Cover | SCA STUDIO | หลิน SCA (แชมป์ มาสเตอร์คีย์เวทีแจ้งเกิด) feat.ศร SCA", Duration: durationFromLengthText("7:24")}, {ID: "9I_uO9AlVOo", Title: "เธอเก่ง (Still) - Jetsetter l Cover by เต้อ", Duration: durationFromLengthText("5:24")}, {ID: "ZysJE2heuA8", Title: "เธอเก่ง - Jetset'er (1 ก.ค.61)", Duration: durationFromLengthText("6:19")}}

	got, err := ExtractSearchResult(f)
	if err != nil {
		t.Error("extract error: ", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant: %+v\ngot %+v", want, got)
	}
}

func TestExtractSuggestedVideos(t *testing.T) {
	f, _ := os.Open("../testdata/watch-page.html")

	want := []VideoInfo{{ID: "dQ3QJoS5aaw", Title: "เก็บรัก - Ammy The Bottom Blues [Official MV]", Duration: durationFromLengthText("4:23")}, {ID: "xxHAADKqG1A", Title: "รถของเล่น/Toycar : เสือโคร่ง/ Tiger", Duration: durationFromLengthText("3:49")}, {ID: "h2nIHSgQ2Bs", Title: "เรื่องที่ขอ - LULA [Official MV]", Duration: durationFromLengthText("6:14")}, {ID: "5H4Lg-rl58U", Title: "ลมเปลี่ยนทิศ - BIG ASS「Official MV」", Duration: durationFromLengthText("4:33")}, {ID: "vIcyhCjWdOQ", Title: "เพลงเพราะ 🎧 เพลงใหม่ล่าสุด 2021 [ เพลงดัง 100 ล้านวิว ] ฟังต่อเนื่อง", Duration: durationFromLengthText("1:23:59")}, {ID: "zYntO183nuY", Title: "หัวใจทศกัณฐ์ [Devil's Heart] - เก่ง ธชย (TACHAYA) ft.ทศกัณฐ์ [Official Lyric Video]", Duration: durationFromLengthText("4:06")}, {ID: "TgbdwLQDV3s", Title: "ป๊อบ ปองกูล - ภาพจำ [Official MV]", Duration: durationFromLengthText("6:27")}, {ID: "twPfxOVczUs", Title: "รวมเพลง Cover Acoustic 2021 เศร้าๆ เพราะๆ เสียงคมชัด ไฟล์ Lossless จากห้องอัด ZaadOat Studio", Duration: durationFromLengthText("2:17:39")}, {ID: "nIVOFq3Xyzs", Title: "เจ็บที่ต้องรู้ - The Mousses「Official MV」", Duration: durationFromLengthText("6:22")}, {ID: "QyhrOruvT1c", Title: "อ้าว - Atom ชนกันต์ [Official MV]", Duration: durationFromLengthText("5:45")}, {ID: "HZV-ggoTQ7s", Title: "ไกลแค่ไหน คือ ใกล้ - getsunova (Official Audio)", Duration: durationFromLengthText("4:27")}, {ID: "lELqMu5HCY0", Title: "W\u200b\u200bANYAi แว่นใหญ่ - เจ็บจนพอ | Enough [Official MV]", Duration: durationFromLengthText("5:36")}, {ID: "pVLgtfpck_U", Title: "รวมเพลง ผู้ชายอกหักๆ BY 👉Ⓜ️🅔Ⓜ️🅘..❤️..", Duration: durationFromLengthText("1:50:32")}, {ID: "PTR-ad3pLQU", Title: "ไม่เดียงสา - BIG ASS「Official MV」", Duration: durationFromLengthText("6:53")}, {ID: "P_qVCZfG1d0", Title: "เหนื่อยไหมหัวใจ feat. ว่าน วันวาน - Retrospect「Lyric Video」", Duration: durationFromLengthText("4:24")}, {ID: "tm2N4gntigI", Title: "Undo - POP PONGKOOL X WONDERFRAME (JOOX 100x100 SEASON 2) 「Official MV」", Duration: durationFromLengthText("5:46")}, {ID: "wqJsZYibWcI", Title: "ซ่อนกลิ่น - PALMY「Official MV」", Duration: durationFromLengthText("5:15")}, {ID: "rt3Y3HEj08I", Title: "รวมเพลงเพราะๆ เพลงฟังตอนทำงาน เปิดในคาเฟ่", Duration: durationFromLengthText("2:10:17")}, {ID: "FlZzhi2usVI", Title: "ปล่อย (Miss) | Clockwork Motionless【Official MV】", Duration: durationFromLengthText("5:19")}}

	got, err := ExtractSuggestedVideos(f)
	if err != nil {
		t.Error("extract error: ", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant: %+v\ngot: %+v", want, got)
	}
}
