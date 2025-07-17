package domain

type Actor struct {
	ID       int     `json:"id"`
	Name     string  `json:"name,omitempty"`
	EngName  string  `json:"enName"`
	Birthday string  `json:"birthday"`
	Photo    string  `json:"photo,omitempty"`
	Movies   []Movie `json:"movies"`
}

type Movie struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	EngName string  `json:"enName"`
	Poster  string  `json:"photo"`
	Rating  float32 `json:"rating"`
	Year    int     `json:"year"`
}

type PhotoData struct {
	ID      int
	URL     string
	Caption string
}
