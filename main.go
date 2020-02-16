package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	d "github.com/black-mamba-out/go-rest-api/Database"
	t "github.com/black-mamba-out/go-rest-api/Types"
	_ "github.com/lib/pq"
)

func main() {
	d.InitDb()
	//d.CloseDb()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/getBasicInfoOfTheWord", getBasicInfoOfTheWord).Methods("GET")
	router.HandleFunc("/indexHandler", indexHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
	fmt.Println("LFC Forever!")
}

func getBasicInfoOfTheWord(w http.ResponseWriter, r *http.Request) {
	var newRequest t.RequestTemplate
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(requestBody, &newRequest)

	existence, err := d.WordExistenceControlByName(newRequest.Word)
	if err != nil {
		fmt.Fprintf(w, "Come'n bro, couldn't check basic existence :( ")
	}
	if !existence {
		newWordResponse, err := goToMerriamWebsterServer(newRequest.Word, newRequest.Key)
		if err != nil {
			fmt.Fprintf(w, "Come'n bro, couldn't get anything :( ")
		}
		err = responseHandler(newWordResponse)
		if err != nil {
			fmt.Fprintf(w, "Come'n bro, couldn't handle response :( ")
		}
		result := populateWordTemplate(newWordResponse.Response[0])
		json.NewEncoder(w).Encode(result)

	} else {
		// result.Name = newRequest.Word
		// err = selectWordDefinitionByName(&result)
		// if err != nil {
		// 	fmt.Fprintf(w, "Come'n bro, couldn't get it from database :( ")
		// }
		json.NewEncoder(w).Encode("You'll never walk alone!")
	}

}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	var word t.WordTemplate
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
func selectWordDefinitionByName(word *t.WordTemplate) error {
	query := fmt.Sprintf(`SELECT * FROM "DEF"."WORDS" AS "A" INNER JOIN "DEF"."SHORT_DEFINITIONS" AS "B" ON "A"."ID" = "B"."WORD_ID" WHERE "A"."NAME" = '%v' limit 1`, word.Name)
	row, err := d.Db.Query(query)
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
func goToMerriamWebsterServer(word string, key string) (t.ResponseTemplate, error) {
	var comingWord t.ResponseTemplate
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
	json.Unmarshal(body, &comingWord.Response)
	return comingWord, err
}
func populateWordTemplate(item t.Response) t.WordTemplate {
	var word t.WordTemplate
	word.UUID = item.Meta.UUID
	word.Type = item.Fl
	word.Name = item.Meta.ID
	word.Language = "ENGLISH"
	word.IsOffensive = item.Meta.Offensive
	if len(item.Meta.Syns) > 0 {
		word.Synonyms = item.Meta.Syns[0]
	}
	if len(item.Meta.Ants) > 0 {
		word.Antonyms = item.Meta.Ants[0]
	}
	word.ShortDefinitions = item.Shortdef
	return word
}
func responseHandler(response t.ResponseTemplate) error {
	for _, item := range response.Response {
		word := populateWordTemplate(item)
		err := d.InsertWordToDatabase(word)
		if err != nil {
			return err
		}
		err = synonymHandler(word)
		if err != nil {
			return err
		}
		err = antonymHandler(word)
		if err != nil {
			return err
		}
		err = shortDefinitionHandler(word)
		if err != nil {
			return err
		}
	}
	return nil
}
func shortDefinitionHandler(word t.WordTemplate) error {
	for _, def := range word.ShortDefinitions {
		existence, err := d.WordDefinitionExistenceControlByName(word.Name, def)
		if err != nil {
			return err
		}
		if !existence {
			err = d.InsertShortDefinitionToDatabase(word, def)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func synonymHandler(word t.WordTemplate) error {
	for _, syn := range word.Synonyms {
		existence, err := d.WordExistenceControlByName(syn)
		if err != nil {
			return err
		}
		if !existence {
			newWord, err := goToMerriamWebsterServer(syn, "dd2d81f6-6955-48a3-a50d-64675396750f")
			if err != nil {
				return err
			}
			wordSyn := populateWordTemplate(newWord.Response[0])
			err = d.InsertWordToDatabase(wordSyn)
			if err != nil {
				return err
			}
			err = shortDefinitionHandler(wordSyn)
			if err != nil {
				return err
			}
			err = d.InsertSynonymRealtionToDatabase(word, wordSyn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func antonymHandler(word t.WordTemplate) error {
	for _, ant := range word.Antonyms {
		existence, err := d.WordExistenceControlByName(ant)
		if err != nil {
			return err
		}
		if !existence {
			newWord, err := goToMerriamWebsterServer(ant, "dd2d81f6-6955-48a3-a50d-64675396750f")
			if err != nil {
				return err
			}
			wordAnt := populateWordTemplate(newWord.Response[0])
			err = d.InsertWordToDatabase(wordAnt)
			if err != nil {
				return err
			}
			err = shortDefinitionHandler(wordAnt)
			if err != nil {
				return err
			}
			err = d.InsertAntonymRealtionToDatabase(word, wordAnt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
