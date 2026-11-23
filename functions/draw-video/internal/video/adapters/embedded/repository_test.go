package embedded

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

const validID = "0001_61,66,18,51,32,37,03,06,50,53,79,56,74,30,70,44,16,45,43,59"
const validID2 = "0002_01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantIDs []string
		wantErr string
	}{
		{name: "one valid video", text: validID + "\n", wantIDs: []string{validID}},
		{name: "blank lines ignored", text: "\n" + validID + "\n\n" + validID2 + "\n\n", wantIDs: []string{validID, validID2}},
		{name: "empty catalog", text: "\n\n", wantErr: "empty"},
		{name: "missing underscore", text: "0001-01,02\n", wantErr: "line 1: id \"0001-01,02\" has no '_'"},
		{name: "non numeric ball", text: validID2 + "\n0003_01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,xx\n", wantErr: `line 2: id "0003_01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,xx" has ball "xx" that is not two digits`},
		{name: "sign in ball", text: "0001_+1,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: `ball "+1" that is not two digits`},
		{name: "single digit ball", text: "0001_1,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: `ball "1" that is not two digits`},
		{name: "three digit ball", text: "0001_001,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: `ball "001" that is not two digits`},
		{name: "empty prefix", text: "_01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: "empty or non-digit prefix"},
		{name: "non digit prefix", text: "00a1_01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: "empty or non-digit prefix"},
		{name: "trailing comma", text: validID + ",\n", wantErr: "has 21 balls, want 20"},
		{name: "empty field in the middle", text: "0001_01,,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: `ball "" that is not two digits`},
		{name: "CRLF line endings", text: validID + "\r\n" + validID2 + "\r\n", wantIDs: []string{validID, validID2}},
		{name: "surrounding spaces trimmed", text: "  " + validID + "  \n", wantIDs: []string{validID}},
		{name: "too few balls", text: "0001_01,02,03\n", wantErr: "has 3 balls, want 20"},
		{name: "too many balls", text: validID + ",80\n", wantErr: "has 21 balls, want 20"},
		{name: "ball zero", text: "0001_00,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: "ball 0 outside 1..80"},
		{name: "ball above 80", text: "0001_81,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: "ball 81 outside 1..80"},
		{name: "repeated ball", text: "0001_01,01,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20\n", wantErr: "repeats ball 1"},
		{name: "duplicate id", text: validID + "\n" + validID + "\n", wantErr: "line 2: duplicate id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videos, err := parse(tt.text)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parse() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse() unexpected error: %v", err)
			}
			var gotIDs []string
			for _, v := range videos {
				gotIDs = append(gotIDs, v.ID)
			}
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Fatalf("ids = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestParseReadsBallsInOrder(t *testing.T) {
	videos, err := parse(validID)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{61, 66, 18, 51, 32, 37, 3, 6, 50, 53, 79, 56, 74, 30, 70, 44, 16, 45, 43, 59}
	if !reflect.DeepEqual(videos[0].Balls, want) {
		t.Fatalf("balls = %v, want %v", videos[0].Balls, want)
	}
}

// Guarda contra desplegar un catalog.txt roto: si no carga, el test falla.
func TestEmbeddedCatalogLoads(t *testing.T) {
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}
	videos, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error: %v", err)
	}
	if len(videos) == 0 {
		t.Fatal("embedded catalog has no videos")
	}
}
