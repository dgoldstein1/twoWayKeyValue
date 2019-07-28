package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var testingDir = "/tmp/twowaykv/temp"

func TestValidateEntry(t *testing.T) {
	t.Run("fails when both are noneType", func(t *testing.T) {
		err := ValidatEntry(Entry{})
		assert.NotNil(t, err)
	})
	t.Run("fails when value is <= 0", func(t *testing.T) {
		err := ValidatEntry(Entry{"test", -3})
		assert.NotNil(t, err)
	})
	t.Run("passes with valid entry", func(t *testing.T) {
		err := ValidatEntry(Entry{"test", 2534})
		assert.Nil(t, err)
	})
}

func TestRemoveDupliactes(t *testing.T) {
	t.Run("removes duplicates from array", func(t *testing.T) {
		passed := []Entry{Entry{"k", 1}, Entry{"k1", 2}, Entry{"k", 1}}
		expected := []Entry{Entry{"k", 1}, Entry{"k1", 2}}
		assert.Equal(t, expected, removeDuplicates(passed))
	})
	t.Run("returns normal array on no duplicates", func(t *testing.T) {
		passed := []Entry{Entry{"k", 1}, Entry{"k1", 2}, Entry{"k2", 3}}
		assert.Equal(t, passed, removeDuplicates(passed))
	})
}

func TestRetrieveEntry(t *testing.T) {
	loadPath := "/tmp/twowaykv/retrieveEntry/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(loadPath)
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	router, _ := SetupRouter("./api/*")

	type Test struct {
		Name                  string
		Path                  string
		Body                  []byte
		ExpectedCode          int
		ExpectedEntriesLength int
		ExpectedErrors        []string
		Method                string
	}
	// used for testing valid value lookup
	validTestValue := ""

	testTable := []Test{
		Test{
			Name:                  "correctly retrieves valid entry",
			Path:                  "/entries",
			Body:                  []byte(`[{"key":"testKey","value":92238547725307}]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "correctly retrieves valid entry (key only)",
			Path:                  "/entries",
			Body:                  []byte(`[{"key":"testKey"}]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "correctly retrieves valid entry (value only)",
			Path:                  "/entries",
			Body:                  []byte(""), // set at execution time
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "adds new key if doesnt exist",
			Path:                  "/entries",
			Body:                  []byte(`[{"key":"testKey2432"}]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "validates bad int type",
			Path:                  "/entries",
			Body:                  []byte(`[{"value":"0"}]`),
			ExpectedCode:          400,
			ExpectedEntriesLength: 0,
			ExpectedErrors:        []string{"json: cannot unmarshal string into Go struct field Entry.value of type int"},
			Method:                "POST",
		},
		Test{
			Name:                  "bad json buffer",
			Path:                  "/entries",
			Body:                  []byte(`2085jf2 3j0d sdf}`),
			ExpectedCode:          400,
			ExpectedEntriesLength: 0,
			ExpectedErrors:        []string{"json: cannot unmarshal number into Go value of type []main.Entry"},
			Method:                "POST",
		},
		Test{
			Name:                  "Creates many entries succesfully",
			Path:                  "/entries",
			Body:                  []byte(`[{"key":"/wiki/The_String_Cheese_Incident","value":0},{"key":"/wiki/Korb%C3%A1%C4%8Diky","value":0},{"key":"/wiki/Slovakia","value":0},{"key":"/wiki/Cheese","value":0},{"key":"/wiki/Mozzarella","value":0},{"key":"/wiki/Milk","value":0},{"key":"/wiki/Protein","value":0},{"key":"/wiki/Slovakia","value":0},{"key":"/wiki/Korb%C3%A1%C4%8Diky","value":0},{"key":"/wiki/Sheep_milk","value":0},{"key":"/wiki/Armenia","value":0},{"key":"/wiki/Nigella_sativa","value":0},{"key":"/wiki/Mahleb","value":0},{"key":"/wiki/Syrian","value":0},{"key":"/wiki/Georgia_(country)","value":0},{"key":"/wiki/Sheep","value":0},{"key":"/wiki/Cream","value":0},{"key":"/wiki/Veal","value":0},{"key":"/wiki/Processed_cheese","value":0},{"key":"/wiki/Kerry_Group","value":0},{"key":"/wiki/Bend_Me,_Shape_Me","value":0},{"key":"/wiki/Disco","value":0},{"key":"/wiki/Funfair","value":0},{"key":"/wiki/Cheddar_cheese","value":0},{"key":"/wiki/Mozzarella","value":0},{"key":"/wiki/Red_leicester","value":0},{"key":"/wiki/Bacon","value":0},{"key":"/wiki/Pizza","value":0},{"key":"/wiki/Gouda_cheese","value":0},{"key":"/wiki/Charleville,_County_Cork","value":0},{"key":"/wiki/Holland","value":0},{"key":"/wiki/Emmental","value":0},{"key":"/wiki/Tesco","value":0},{"key":"/wiki/Dairylea_Cooperative_Inc.","value":0},{"key":"/wiki/Dunnes_Stores","value":0},{"key":"/wiki/Lunchables","value":0},{"key":"/wiki/Tortilla_wrap","value":0},{"key":"/wiki/Cracker_(food)","value":0},{"key":"/wiki/Tomato_ketchup","value":0},{"key":"/wiki/Spam_(food)","value":0},{"key":"/wiki/Mexico","value":0},{"key":"/wiki/Quesillo","value":0},{"key":"/wiki/Queso_Oaxaca","value":0},{"key":"/wiki/United_States","value":0},{"key":"/wiki/Mozzarella","value":0},{"key":"/wiki/Cheddar_cheese","value":0},{"key":"/wiki/Bega_Cheese","value":0},{"key":"/wiki/Armenian_cuisine","value":0},{"key":"/wiki/List_of_cheeses","value":0},{"key":"/wiki/List_of_stretch-cured_cheeses","value":0},{"key":"/wiki/Pasta_filata","value":0},{"key":"/wiki/The_Atlantic","value":0},{"key":"/wiki/Atlantic_Media","value":0},{"key":"/wiki/List_of_American_cheeses","value":0},{"key":"/wiki/Swiss_cheese#Varieties","value":0},{"key":"/wiki/Bergenost","value":0},{"key":"/wiki/Brick_cheese","value":0},{"key":"/wiki/Cheese_curd","value":0},{"key":"/wiki/Colby_cheese","value":0},{"key":"/wiki/Colby-Jack","value":0},{"key":"/wiki/Cougar_Gold_cheese","value":0},{"key":"/wiki/Cream_cheese","value":0},{"key":"/wiki/Creole_cream_cheese","value":0},{"key":"/wiki/Cuba_cheese","value":0},{"key":"/wiki/D%27Isigny","value":0},{"key":"/wiki/Farmer_cheese","value":0},{"key":"/wiki/Hoop_cheese","value":0},{"key":"/wiki/Humboldt_Fog","value":0},{"key":"/wiki/Kunik_cheese","value":0},{"key":"/wiki/Liederkranz_cheese","value":0},{"key":"/wiki/Maytag_Blue_cheese","value":0},{"key":"/wiki/Monterey_Jack","value":0},{"key":"/wiki/Muenster_cheese","value":0},{"key":"/wiki/Pinconning_cheese","value":0},{"key":"/wiki/Red_Hawk_cheese","value":0},{"key":"/wiki/Swiss_cheese","value":0},{"key":"/wiki/Teleme_cheese","value":0},{"key":"/wiki/Wisconsin_cheese","value":0},{"key":"/wiki/String_cheese","value":0},{"key":"/wiki/String_cheese","value":0},{"key":"/wiki/Main_Page","value":0},{"key":"/wiki/Main_Page","value":0},{"key":"/wiki/String_cheese","value":0}]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 75,
			Method:                "POST",
			ExpectedErrors:        []string{},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			if test.Name == "correctly retrieves valid entry (value only)" {
				test.Body = []byte(`[{"value" : ` + string(validTestValue) + `}]`)
			}
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)

			// fmt.Println(" ****> POST: " + string(test.Body))
			body := []byte(w.Body.String())
			if test.ExpectedCode == 200 {
				resp := RetrieveEntryResponse{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				assert.Equal(t, test.ExpectedEntriesLength, len(resp.Entries))
				assert.Equal(t, test.ExpectedErrors, resp.Errors)
				// set createdEntry on success
				if len(resp.Entries) > 0 {
					assert.NotEqual(t, 0, resp.Entries[0].Value)
					validTestValue = strconv.Itoa(resp.Entries[0].Value)
					assert.NotEqual(t, "0", validTestValue)
				}
			} else {
				resp := Error{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				assert.Equal(t, test.ExpectedErrors[0], resp.Error)
				assert.Equal(t, test.ExpectedCode, resp.Code)
			}
		})

	}
}

func TestExportDb(t *testing.T) {
	loadPath := "/tmp/twowaykv/iotest/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(loadPath)
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	router, _ := SetupRouter("./api/*")

	type Test struct {
		Name           string
		Path           string
		Body           []byte
		ExpectedCode   int
		ExpectedErrors []string
		Method         string
	}
	// used for testing valid value lookup

	testTable := []Test{
		Test{
			Name:           "returns no nil stream on success",
			Path:           "/export",
			Body:           []byte(""),
			ExpectedCode:   200,
			ExpectedErrors: []string{},
			Method:         "GET",
		},
		Test{
			Name:           "returns error if file could not be found",
			Path:           "/export",
			Body:           []byte(""),
			ExpectedCode:   500,
			ExpectedErrors: []string{"\tzip warning: name not matched: /temp/randomDir/doesntexist/k2v\n\tzip warning: name not matched: /temp/randomDir/doesntexist/v2k\n\nzip error: Nothing to do! (try: zip -r /temp/randomDir/doesntexist/twowaykv_export.zip . -i /temp/randomDir/doesntexist/k2v /temp/randomDir/doesntexist/v2k)\n"},
			Method:         "GET",
		},
	}

	for _, test := range testTable {
		if test.Name == "returns error if file could not be found" {
			os.Setenv("GRAPH_DB_STORE_DIR", "/temp/randomDir/doesntexist")
			defer os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
		}
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)

			body := []byte(w.Body.String())
			if test.ExpectedCode == 200 {
				assert.NotNil(t, body)
			} else {
				resp := Error{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				if err != nil {
					t.Errorf("Could not unmarshal resp %s", string(body))
				}
				assert.Equal(t, test.ExpectedErrors[0], resp.Error)
				assert.Equal(t, test.ExpectedCode, resp.Code)
			}
		})

	}
}
