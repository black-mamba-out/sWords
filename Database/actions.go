package database

import (
	"fmt"
	"strings"

	t "github.com/black-mamba-out/go-rest-api/Types"
	_ "github.com/lib/pq"
)

// InsertWordToDatabase insert word to table of DEF.WORDS
func InsertWordToDatabase(word t.WordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."WORDS"("RECORD_STATUS", "NAME", "LANGUAGE", "IS_OFFENSIVE", "REGISTER_DATE", "LAST_UPDATE_DATE", "UUID", "TYPE")
						VALUES (B'1', $1, $2, B'1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $3, $4)
							RETURNING "ID"`
	ID := 0
	err := Db.QueryRow(sqlStatement, word.Name, word.Language, word.UUID, word.Type).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New record ID is:", ID)
	return err
}

// InsertSynonymRealtionToDatabase insert synonym relation to table of DEF.SYNONYMS
func InsertSynonymRealtionToDatabase(word t.WordTemplate, wordSyn t.WordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."SYNONYMS"( "WORD_ID", "SYNONYM_WORD_ID", "RECORD_STATUS", "LAST_UPDATE_DATE")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), (SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $2), B'1', CURRENT_TIMESTAMP)
							RETURNING "WORD_ID"`
	ID := 0
	err := Db.QueryRow(sqlStatement, word.UUID, wordSyn.UUID).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New synonym relation was added to:", ID)
	return err
}

// InsertAntonymRealtionToDatabase insert antonym relation to table of DEF.ANTONYMS
func InsertAntonymRealtionToDatabase(word t.WordTemplate, wordAnt t.WordTemplate) error {
	sqlStatement := `INSERT INTO "DEF"."ANTONYMS"( "WORD_ID", "ANTONYM_WORD_ID", "RECORD_STATUS", "LAST_UPDATE_DATE")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), (SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $2), B'1', CURRENT_TIMESTAMP)
							RETURNING "WORD_ID"`
	ID := 0
	err := Db.QueryRow(sqlStatement, word.UUID, wordAnt.UUID).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New antonym relation was added to:", ID)
	return err
}

// InsertShortDefinitionToDatabase insert short definition to table of DEF.SHORT_DEFINITIONS
func InsertShortDefinitionToDatabase(word t.WordTemplate, wordDefinition string) error {
	sqlStatement := `INSERT INTO "DEF"."SHORT_DEFINITIONS"("WORD_ID", "MEANING", "LAST_UPDATE_DATE", "RECORD_STATUS")
						VALUES ((SELECT "ID" FROM "DEF"."WORDS" WHERE "UUID" = $1), $2, CURRENT_TIMESTAMP, B'1')
							RETURNING "WORD_ID"`
	ID := 0
	err := Db.QueryRow(sqlStatement, word.UUID, wordDefinition).Scan(&ID)
	if err != nil {
		return err
	}
	fmt.Println("New short defifinition was added to:", ID)
	return err
}

// WordExistenceControlByName controls word existence by name
func WordExistenceControlByName(wordName string) (bool, error) {
	query := fmt.Sprintf(`select 1 as EXISTENCE FROM "DEF"."WORDS" WHERE "NAME" = '%v' limit 1;`, wordName)
	row, err := Db.Query(query)
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

// WordDefinitionExistenceControlByName controls word definition existence by name
func WordDefinitionExistenceControlByName(wordName string, wordDefinition string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 as EXISTENCE FROM "DEF"."WORDS" AS "A" INNER JOIN "DEF"."SHORT_DEFINITIONS" AS "B" ON "A"."ID" = "B"."WORD_ID" 
	                            WHERE "A"."NAME" = '%s' AND "B"."MEANING" = '%s' limit 1`, strings.Replace(wordName, "'", "''", -1), strings.Replace(wordDefinition, "'", "''", -1))
	row, err := Db.Query(query)
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
