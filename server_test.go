package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	badger "github.com/dgraph-io/badger"
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

func TestRemoveDupliactes(t *testing.T) {
	t.Run("removes duplicates from array", func(t *testing.T) {
		passed := []string{"k", "k1", "k"}
		expected := []string{"k", "k1"}
		assert.Equal(t, expected, removeDuplicates(passed))
	})
	t.Run("returns normal array on no duplicates", func(t *testing.T) {
		passed := []string{"k1", "k2", "k3"}
		assert.Equal(t, passed, removeDuplicates(passed))
	})
}

func TestRandomEntries(t *testing.T) {

	loadPath := "/tmp/twowaykv/randomEntries/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(loadPath)
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	router, s := SetupRouter("./api/*")

	// insert some randm stuff into db
	err = s.V2k.Update(func(txn *badger.Txn) error {
		for i := 0; i < 10; i++ {
			if e := txn.Set([]byte(strconv.Itoa(i+2)), []byte("TEST-KEY")); e != nil {
				return e
			}
		}
		return nil
	})
	require.Nil(t, err)

	type Test struct {
		Name                  string
		Path                  string
		Before                func()
		ExpectedCode          int
		ExpectedEntriesLength int
		ExpectedError         string
		Setup                 func()
		Method                string
	}
	// used for testing valid value lookup
	validTestValue := ""

	testTable := []Test{
		Test{
			Name:                  "gets random entry",
			Path:                  "/random",
			Before:                func() {},
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			Method:                "GET",
		},
		Test{
			Name:                  "invalid int",
			Path:                  "/random?n=XXXXX",
			Before:                func() {},
			ExpectedCode:          400,
			ExpectedError:         "Invalid int",
			ExpectedEntriesLength: 1,
			Method:                "GET",
		},
		Test{
			Name:                  "invalid int",
			Path:                  "/random?n=-34234",
			Before:                func() {},
			ExpectedCode:          400,
			ExpectedError:         "'n' must be positive and greater than 25",
			ExpectedEntriesLength: 1,
			Method:                "GET",
		},
		Test{
			Name: "returns error from db call",
			Path: "/random",
			Before: func() {
				// insert some randm stuff into db
				err = s.V2k.Update(func(txn *badger.Txn) error {
					for i := 0; i < 10; i++ {
						if e := txn.Delete([]byte(strconv.Itoa(i + 2))); e != nil {
							return e
						}
					}
					return nil
				})
				require.Nil(t, err)

			},
			ExpectedCode:          500,
			ExpectedError:         "max collisions reached finding random entries",
			ExpectedEntriesLength: 1,
			Method:                "GET",
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Before()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer([]byte("")))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)

			// fmt.Println(" ****> POST: " + string(test.Body))
			body := []byte(w.Body.String())
			if test.ExpectedCode == 200 {
				resp := []Entry{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				assert.Equal(t, test.ExpectedEntriesLength, len(resp))
				// set createdEntry on success
				if len(resp) > 0 {
					assert.NotEqual(t, 0, resp[0].Value)
					validTestValue = strconv.Itoa(resp[0].Value)
					assert.NotEqual(t, "0", validTestValue)
				}
			} else {
				resp := Error{}
				err := json.Unmarshal(body, &resp)
				require.Nil(t, err)
				assert.Equal(t, test.ExpectedError, resp.Error)
				assert.Equal(t, test.ExpectedCode, resp.Code)
			}
		})

	}

}

func TestCreateEntriesEntry(t *testing.T) {
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
		Setup                 func()
		Method                string
	}
	// used for testing valid value lookup
	validTestValue := ""

	testTable := []Test{
		Test{
			Name:                  "correctly retrieves valid entry",
			Path:                  "/entries",
			Body:                  []byte(`["testKey"]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "mutes key already exists",
			Path:                  "/entries?muteAlreadyExistsError=false",
			Body:                  []byte(`["testKey"]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{"Key testKey already exists in DB"},
			Method:                "POST",
		},
		Test{
			Name:                  "mutes key already exists",
			Path:                  "/entries?muteAlreadyExistsError=true",
			Body:                  []byte(`["testKey"]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		Test{
			Name:                  "bad json buffer",
			Path:                  "/entries",
			Body:                  []byte(`2085jf2 3j0d sdf}`),
			ExpectedCode:          400,
			ExpectedEntriesLength: 0,
			ExpectedErrors:        []string{"json: cannot unmarshal number into Go value of type []string"},
			Method:                "POST",
		},
		Test{
			Name:                  "Creates many entries succesfully",
			Path:                  "/entries",
			Body:                  []byte(`["/wiki/The_String_Cheese_Incident","/wiki/Korb%C3%A1%C4%8Diky","/wiki/Slovakia","/wiki/Cheese","/wiki/Mozzarella","/wiki/Milk","/wiki/Protein","/wiki/Slovakia","/wiki/Korb%C3%A1%C4%8Diky","/wiki/Sheep_milk","/wiki/Armenia","/wiki/Nigella_sativa","/wiki/Mahleb","/wiki/Syrian","/wiki/Georgia_(country)","/wiki/Sheep","/wiki/Cream","/wiki/Veal","/wiki/Processed_cheese","/wiki/Kerry_Group","/wiki/Bend_Me,_Shape_Me","/wiki/Disco","/wiki/Funfair","/wiki/Cheddar_cheese","/wiki/Mozzarella","/wiki/Red_leicester","/wiki/Bacon","/wiki/Pizza","/wiki/Gouda_cheese","/wiki/Charleville,_County_Cork","/wiki/Holland","/wiki/Emmental","/wiki/Tesco","/wiki/Dairylea_Cooperative_Inc.","/wiki/Dunnes_Stores","/wiki/Lunchables","/wiki/Tortilla_wrap","/wiki/Cracker_(food)","/wiki/Tomato_ketchup","/wiki/Spam_(food)","/wiki/Mexico","/wiki/Quesillo","/wiki/Queso_Oaxaca","/wiki/United_States","/wiki/Mozzarella","/wiki/Cheddar_cheese","/wiki/Bega_Cheese","/wiki/Armenian_cuisine","/wiki/List_of_cheeses","/wiki/List_of_stretch-cured_cheeses","/wiki/Pasta_filata","/wiki/The_Atlantic","/wiki/Atlantic_Media","/wiki/List_of_American_cheeses","/wiki/Swiss_cheese#Varieties","/wiki/Bergenost","/wiki/Brick_cheese","/wiki/Cheese_curd","/wiki/Colby_cheese","/wiki/Colby-Jack","/wiki/Cougar_Gold_cheese","/wiki/Cream_cheese","/wiki/Creole_cream_cheese","/wiki/Cuba_cheese","/wiki/D%27Isigny","/wiki/Farmer_cheese","/wiki/Hoop_cheese","/wiki/Humboldt_Fog","/wiki/Kunik_cheese","/wiki/Liederkranz_cheese","/wiki/Maytag_Blue_cheese","/wiki/Monterey_Jack","/wiki/Muenster_cheese","/wiki/Pinconning_cheese","/wiki/Red_Hawk_cheese","/wiki/Swiss_cheese","/wiki/Teleme_cheese","/wiki/Wisconsin_cheese","/wiki/String_cheese","/wiki/String_cheese","/wiki/Main_Page","/wiki/Main_Page","/wiki/String_cheese"]`),
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

// tests both "/entriesFromKeys" and "/entriesFromValues"
func TestGetEntries(t *testing.T) {
	loadPath := "/tmp/twowaykv/api/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(loadPath)
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	router, s := SetupRouter("./api/*")

	type Test struct {
		Name                  string
		Path                  string
		Body                  []byte
		ExpectedCode          int
		ExpectedEntriesLength int
		ExpectedErrorsLength  int
		Setup                 func()
		TearDown              func()
		Method                string
	}
	testTable := []Test{
		Test{
			Name:                  "gets correct entries from keys",
			Path:                  "/entriesFromKeys",
			Body:                  []byte(`["testKey"]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrorsLength:  0,
			Method:                "POST",
			Setup: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Set([]byte("testKey"), []byte("111")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)

			},
			TearDown: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Delete([]byte("testKey")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)
			},
		},
		Test{
			Name:                  "returns error for bad json",
			Path:                  "/entriesFromKeys",
			Body:                  []byte(`["testKeyad"'f']la;d;fla;df`),
			ExpectedCode:          400,
			ExpectedEntriesLength: 0,
			ExpectedErrorsLength:  1,
			Method:                "POST",
			Setup: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Set([]byte("testKey"), []byte("111")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)

			},
			TearDown: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Delete([]byte("testKey")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)
			},
		},
		Test{

			Name:                  "gets correct entries from values",
			Path:                  "/entriesFromValues",
			Body:                  []byte(`[115]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrorsLength:  0,
			Method:                "POST",
			Setup: func() {
				err := s.V2k.Update(func(txn *badger.Txn) error {
					if e := txn.Set([]byte("115"), []byte("testKey115")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)

			},
			TearDown: func() {
				err := s.V2k.Update(func(txn *badger.Txn) error {
					if e := txn.Delete([]byte("115")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)
			},
		},
		Test{

			Name:                  "returns error for bad json",
			Path:                  "/entriesFromValues",
			Body:                  []byte(`["1113"]`),
			ExpectedCode:          400,
			ExpectedEntriesLength: 0,
			ExpectedErrorsLength:  1,
			Method:                "POST",
			Setup: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Set([]byte("testKey"), []byte("111")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)

			},
			TearDown: func() {
				err := s.K2v.Update(func(txn *badger.Txn) error {
					if e := txn.Delete([]byte("testKey")); e != nil {
						return e
					}
					return nil
				})
				require.Nil(t, err)
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Setup()
			w := httptest.NewRecorder()
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
				if test.ExpectedErrorsLength != len(resp.Errors) && len(resp.Errors) != 0 {
					fmt.Println("------------------------------------------")
					fmt.Println(resp.Errors)
				}
				assert.Equal(t, test.ExpectedErrorsLength, len(resp.Errors))
			} else {
				resp := Error{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				assert.Equal(t, test.ExpectedCode, resp.Code)
				assert.NotEqual(t, "", resp.Error)
			}
			test.TearDown()
		})

	}
}

func TestSearch(t *testing.T) {

	loadPath := "/tmp/twowaykv/randomEntries/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(loadPath)
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	router, s := SetupRouter("./api/*")

	// insert some randm stuff into db
	err = s.K2v.Update(func(txn *badger.Txn) error {
		for i := 0; i < 10; i++ {
			if e := txn.Set([]byte("TEST-KEY-"+strconv.Itoa(i)), []byte("1")); e != nil {
				return e
			}
		}
		return nil
	})
	require.Nil(t, err)

	type Test struct {
		Name                  string
		Path                  string
		Before                func()
		ExpectedCode          int
		ExpectedEntriesLength int
		ExpectedErrorsLength  int
		Setup                 func()
		Method                string
	}
	testTable := []Test{
		Test{
			Name:                  "finds all keys starting with TEST-KEY-",
			Path:                  "/search?q=TES",
			Before:                func() {},
			ExpectedCode:          200,
			ExpectedEntriesLength: 10,
			Method:                "GET",
		},
		Test{
			Name:                 "returns error if no query is passed",
			Path:                 "/search",
			Before:               func() {},
			ExpectedCode:         400,
			ExpectedErrorsLength: 1,
			Method:               "GET",
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Before()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer([]byte("")))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)
			body := []byte(w.Body.String())
			if test.ExpectedCode == 200 {
				resp := RetrieveEntryResponse{}
				err := json.Unmarshal(body, &resp)
				assert.Nil(t, err)
				assert.Equal(t, test.ExpectedEntriesLength, len(resp.Entries))
				assert.Equal(t, test.ExpectedErrorsLength, len(resp.Errors))
			} else {
				resp := Error{}
				err := json.Unmarshal(body, &resp)
				require.Nil(t, err)
				assert.Equal(t, test.ExpectedCode, resp.Code)
				assert.Equal(t, test.ExpectedErrorsLength, 1)
			}
		})

	}

}
