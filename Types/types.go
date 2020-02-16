package types

// Response is response
type Response struct {
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

// ResponseTemplate is response template
type ResponseTemplate struct {
	Response []Response
}

// RequestTemplate is request template
type RequestTemplate struct {
	Key  string `json:"Key"`
	Word string `json:"Word"`
}

// BasicInfoOfTheWord is basic info of the word
type BasicInfoOfTheWord struct {
	Word     string `json:"Word"`
	ShortDef string `json:"ShortDef"`
	WordType string `json:"WordType"`
}

// WordTemplate is a word's defult template
type WordTemplate struct {
	UUID             string   `json:"UUID"`
	Name             string   `json:"Name"`
	Language         string   `json:"Language"`
	IsOffensive      bool     `json:"IsOffensive"`
	Type             string   `json:"Type"`
	Synonyms         []string `json:"Synonyms"`
	Antonyms         []string `json:"Antonyms"`
	ShortDefinitions []string `json:"ShortDefinition"`
}
