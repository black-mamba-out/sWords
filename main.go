package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type allEvents []event

var events = allEvents{
	{
		ID:          "1",
		Title:       "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	},
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Anfield Road")
}

func main() {
	initDb()
	defer db.Close()
	router := mux.NewRouter().StrictSlash(true)
	//router.HandleFunc("/", homeLink)
	router.HandleFunc("/getBasicInfoOfTheWord", getBasicInfoOfTheWord).Methods("GET")
	router.HandleFunc("/indexHandler", indexHandler).Methods("GET")
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	router.HandleFunc("/events/{id}", updateEvent).Methods("PATCH")
	router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
	fmt.Println("LFC Forever!")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newEvent)
	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)
}
func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}
func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}
func updateEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]
	var updatedEvent event

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(reqBody, &updatedEvent)

	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			singleEvent.Title = updatedEvent.Title
			singleEvent.Description = updatedEvent.Description
			events = append(events[:i], singleEvent)
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}
func deleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			events = append(events[:i], events[i+1:]...)
			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", eventID)
		}
	}
}

type responseTemplate []struct {
	Meta struct {
		ID      string `json:"id"`
		UUID    string `json:"uuid"`
		Src     string `json:"src"`
		Section string `json:"section"`
		Target  struct {
			Tuuid string `json:"tuuid"`
			Tsrc  string `json:"tsrc"`
		} `json:"target"`
		Stems     []string   `json:"stems"`
		Syns      [][]string `json:"syns"`
		Ants      [][]string `json:"ants"`
		Offensive bool       `json:"offensive"`
	} `json:"meta"`
	Hwi struct {
		Hw string `json:"hw"`
	} `json:"hwi"`
	Fl  string `json:"fl"`
	Def []struct {
		Sseq [][][]interface{} `json:"sseq"`
	} `json:"def"`
	Shortdef []string `json:"shortdef"`
}
type requestTemplate struct {
	Key  string `json:"Key"`
	Word string `json:"Word"`
}
type basicInfoOfTheWord struct {
	Word     string `json:"Word"`
	ShortDef string `json:"ShortDef"`
	WordType string `json:"WordType"`
}
type wordTemplate struct {
	UUID             string   `json:"UUID"`
	Name             string   `json:"Name"`
	Language         string   `json:"Language"`
	IsOffensive      bool     `json:"IsOffensive"`
	Type             string   `json:"Type"`
	Synonyms         []string `json:"Synonyms"`
	Antonyms         []string `json:"Antonyms"`
	ShortDefinitions []string `json:"ShortDefinition"`
}

func getBasicInfoOfTheWord(w http.ResponseWriter, r *http.Request) {
	var newRequest requestTemplate
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(requestBody, &newRequest)

	// var word wordTemplate
	// word.Name = newRequest.Word
	existence, err := wordExistenceControlByName(newRequest.Word)
	if err != nil {
		fmt.Fprintf(w, "Come'n bro, couldn't check basic existence :( ")
	}
	var result wordTemplate
	if !existence {
		newWordResponse, err := goToMerriamWebsterServer(newRequest.Word, newRequest.Key)
		if err != nil {
			fmt.Fprintf(w, "Come'n bro, couldn't get anything :( ")
		}
		responseHandler(newWordResponse)
		result.Type = newWordResponse[0].Fl
		result.Name = newWordResponse[0].Meta.ID
		result.ShortDefinitions = newWordResponse[0].Shortdef
		result.Synonyms = newWordResponse[0].Meta.Syns[0]
		result.Antonyms = newWordResponse[0].Meta.Ants[0]
		result.UUID = newWordResponse[0].Meta.UUID
		result.IsOffensive = newWordResponse[0].Meta.Offensive
	} else {
		result.Name = newRequest.Word
		err = selectWordDefinitionByName(&result)
		if err != nil {
			fmt.Fprintf(w, "Come'n bro, couldn't get it from database :( ")
		}
	}
	json.NewEncoder(w).Encode(result)
}

func initDb() {
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	var word wordTemplate
	word.Name = "YNWA"

	err := selectWordDefinitionByName(&word)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(word)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func selectWordDefinitionByName(word *wordTemplate) error {
	query := fmt.Sprintf(`SELECT * FROM "DEF"."WORDS" AS "A" INNER JOIN "DEF"."SHORT_DEFINITIONS" AS "B" ON "A"."ID" = "B"."WORD_ID" WHERE "A"."NAME" = '%v' limit 1`, word.Name)
	row, err := db.Query(query)
	if err != nil {
		return err
	}
	defer row.Close()
	for row.Next() {
		err = row.Scan(
			&word.Name,
			&word.Type,
			&word.UUID,
			&word.ShortDefinitions[0],
			&word.IsOffensive,
		)

		if err != nil {
			return err
		}
	}
	err = row.Err()
	if err != nil {
		return err
	}
	return nil
}

func insertWord(word basicInfoOfTheWord) error {
	sqlStatement := `INSERT INTO public."WORD"("LAST_UPDATE_DATE", "RECORD_STATUS", "TYPE", "REGISTER_DATE", "NAME", "MEANING", "LANGUAGE")
						VALUES (CURRENT_TIMESTAMP, 'A', $1, CURRENT_TIMESTAMP, $2, $3, 'ENGLISH')
							RETURNING "ID"`
	id := 0
	err := db.QueryRow(sqlStatement, word.WordType, word.Word, word.ShortDef).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Println("New record ID is:", id)
	return err
}

func wordExistenceControlByName(wordName string) (bool, error) {
	query := fmt.Sprintf(`select 1 as EXISTENCE FROM "DEF"."WORDS" WHERE "NAME" = '%v' limit 1;`, wordName)
	row, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer row.Close()
	existence := 0
	for row.Next() {
		err = row.Scan(
			&existence,
		)
		if err != nil {
			return false, err
		}
	}
	err = row.Err()
	if err != nil {
		return false, err
	}
	if existence == 1 {
		return true, err
	}
	return false, err
}
func wordDefinitionExistenceControlByName(wordName string, wordDefinition string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 as EXISTENCE FROM "DEF"."WORDS" AS "A" INNER JOIN "DEF"."SHORT_DEFINITIONS" AS "B" ON "A"."ID" = "B"."WORD_ID" 
	                            WHERE "A"."NAME" = '%s' AND "B"."MEANING" = '%s' limit 1`, strings.Replace(wordName, "'", "''", -1), strings.Replace(wordDefinition, "'", "''", -1))
	row, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer row.Close()
	existence := 0
	for row.Next() {
		err = row.Scan(
			&existence,
		)
		if err != nil {
			return false, err
		}
	}
	err = row.Err()
	if err != nil {
		return false, err
	}
	if existence == 1 {
		return true, err
	}
	return false, err
}

func goToMerriamWebsterServer(word string, key string) (responseTemplate, error) {
	var comingWord responseTemplate
	url := fmt.Sprintf("https://www.dictionaryapi.com/api/v3/references/thesaurus/json/%v?key=%v", word, key)
	resp, err := http.Get(url)
	if err != nil {
		return comingWord, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return comingWord, err
	}
	json.Unmarshal(body, &comingWord)
	return comingWord, err
}
func responseHandler(response responseTemplate) error {
	var err error
	for i, item := range response {
		var word wordTemplate
		word.UUID = item.Meta.UUID
		word.Type = item.Fl
		word.Name = item.Meta.ID
		word.Language = "ENGLISH"
		word.IsOffensive = item.Meta.Offensive
		word.Synonyms = item.Meta.Syns[0]
		// word.Antonyms = item.Meta.Ants[0]
		word.ShortDefinitions = item.Shortdef
		err = insertWordToDatabase(word)
		if err != nil {
			return err
		}
		for j, syn := range word.Synonyms {
			existence, err := wordExistenceControlByName(syn)
			if err != nil {
				return err
			}
			var wordSyn wordTemplate
			if !existence {
				newWord, err := goToMerriamWebsterServer(syn, "dd2d81f6-6955-48a3-a50d-64675396750f")
				wordSyn.UUID = newWord[0].Meta.UUID
				wordSyn.Type = newWord[0].Fl
				wordSyn.Name = newWord[0].Meta.ID
				wordSyn.Language = "ENGLISH"
				wordSyn.IsOffensive = item.Meta.Offensive
				wordSyn.ShortDefinitions = newWord[0].Shortdef
				err = insertWordToDatabase(wordSyn)
				for _, synShortDef := range wordSyn.ShortDefinitions {
					existence, err := wordDefinitionExistenceControlByName(wordSyn.Name, synShortDef)
					if err != nil {
						return err
					}
					if !existence {
						err = insertShortDefinitionToDatabase(wordSyn, synShortDef)
						if err != nil {
							return err
						}
					}
				}
				if err != nil {
					return err
				}
				err = insertSynonymRealtionToDatabase(word, wordSyn)
				if err != nil {
					return err
				}
			}
			fmt.Println(j, wordSyn)
		}
		// for k, ant := range word.Antonyms {
		// 	existence, err := wordExistenceControlByName(ant)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	var wordAnt wordTemplate
		// 	if !existence {
		// 		newWord, err := goToMerriamWebsterServer(ant, "dd2d81f6-6955-48a3-a50d-64675396750f")
		// 		wordAnt.UUID = newWord[0].Meta.UUID
		// 		wordAnt.Type = newWord[0].Fl
		// 		wordAnt.Name = newWord[0].Meta.ID
		// 		wordAnt.Language = "ENGLISH"
		// 		wordAnt.IsOffensive = item.Meta.Offensive
		// 		err = insertWordToDatabase(wordAnt)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		insertAntonymRealtionToDatabase(word, wordAnt)
		// 	}
		// 	fmt.Println(k, wordAnt)
		// }
		for _, def := range word.ShortDefinitions {
			existence, err := wordDefinitionExistenceControlByName(word.Name, def)
			if err != nil {
				return err
			}
			if !existence {
				err = insertShortDefinitionToDatabase(word, def)
				if err != nil {
					return err
				}
			}
		}
		fmt.Println(i, item)
		return err
	}
	return err
}
func insertWordToDatabase(word wordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."WORDS"("RECORD_STATUS", "NAME", "LANGUAGE", "IS_OFFENSIVE", "REGISTER_DATE", "LAST_UPDATE_DATE", "UUID", "TYPE")
						VALUES (B'1', $1, $2, B'1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $3, $4)
							RETURNING "ID"`
	ID := 0
	err := db.QueryRow(sqlStatement, word.Name, word.Language, word.UUID, word.Type).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New record ID is:", ID)
	return err
}
func insertSynonymRealtionToDatabase(word wordTemplate, wordSyn wordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."SYNONYMS"( "WORD_ID", "SYNONYM_WORD_ID", "RECORD_STATUS", "LAST_UPDATE_DATE")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), (SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $2), B'1', CURRENT_TIMESTAMP)
							RETURNING "WORD_ID"`
	ID := 0
	err := db.QueryRow(sqlStatement, word.UUID, wordSyn.UUID).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New synonym relation was added to:", ID)
	return err
}
func insertAntonymRealtionToDatabase(word wordTemplate, wordAnt wordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."ANTONYMS"( "WORD_ID", "ANTONYM_WORD_ID", "RECORD_STATUS", "LAST_UPDATE_DATE")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), (SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $2), B'1', CURRENT_TIMESTAMP)
							RETURNING "WORD_ID"`
	ID := 0
	err := db.QueryRow(sqlStatement, word.UUID, wordAnt.UUID).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New antonym relation was added to:", ID)
	return err
}
func insertShortDefinitionToDatabase(word wordTemplate, wordDefinition string) error {
	sqlStatement := `INSERT INTO "DEF"."SHORT_DEFINITIONS"("WORD_ID", "MEANING", "LAST_UPDATE_DATE", "RECORD_STATUS")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), $2, CURRENT_TIMESTAMP, B'1')
							RETURNING "WORD_ID"`
	ID := 0
	err := db.QueryRow(sqlStatement, word.UUID, wordDefinition).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New short defifinition was added to:", ID)
	return err
}
