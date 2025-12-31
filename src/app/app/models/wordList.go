package models

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/go-gorp/gorp"
	"github.com/google/uuid"
	"github.com/str20tbl/revel"
)

type Word struct {
	ID            int64  `db:"id"`
	Text          string `db:"Text"`
	English       string `db:"English"`
	AudioFilename string `db:"AudioFilename"`
}

// GetWordByID retrieves a word by its ID
func GetWordByID(txn *gorp.Transaction, wordID int64) (*Word, error) {
	var word Word
	err := txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", wordID)
	if err != nil {
		revel.AppLog.Errorf("Could not find word with id %d: %v", wordID, err)
		return nil, err
	}
	return &word, nil
}

// GetMaxWordID returns the maximum Word ID in the database
func GetMaxWordID(txn *gorp.Transaction) (int64, error) {
	maxID, err := txn.SelectInt("SELECT MAX(id) FROM Word")
	if err != nil {
		revel.AppLog.Errorf("Could not get max word ID: %v", err)
		return 0, err
	}
	return maxID, nil
}

func LoadWords(dbm *gorp.DbMap) (err error) {
	count := int64(0)
	count, err = dbm.SelectInt("SELECT COUNT(*) FROM Word")
	if err != nil {
		return
	}
	revel.AppLog.Infof("Found %d words", count)
	if count != 880 {
		queries := []string{
			"DELETE FROM Word",
			"ALTER TABLE Word AUTO_INCREMENT = 1",
		}
		for _, query := range queries {
			_, err = dbm.Exec(query)
			if err != nil {
				return
			}
		}
		filename := "/data/private/words_expanded.tsv"
		words := make([]Word, 0)
		words, err = readTSVFileMultiColumn(filename)
		if err != nil {
			return
		}
		fmt.Printf("Loaded %d words:\n", len(words))
		for _, word := range words {
			if len(word.Text) > 3 {
				err = dbm.Insert(&word)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

// readTSVFileMultiColumn reads TSV with multiple columns
func readTSVFileMultiColumn(filename string) ([]Word, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			revel.AppLog.Errorf("Failed to close file %s: %v", filename, err)
		}
	}()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading TSV file: %w", err)
	}

	var words []Word
	for i, record := range records {
		if len(record) == 0 {
			continue
		}
		// Skip header row
		if i == 0 && (record[0] == "Welsh" || record[0] == "Cymraeg") {
			continue
		}

		word := Word{
			Text:          strings.TrimSpace(record[0]),
			AudioFilename: uuid.New().String(),
		}

		// Read English translation from column 2 if it exists
		if len(record) > 1 {
			word.English = strings.TrimSpace(record[1])
		}
		words = append(words, word)
	}

	return words, nil
}
