package domain

type Actor struct {
	ID       int    `json:"id"`
	Name     string `json:"name,omitempty"`
	EngName  string `json:"enName"`
	Birthday string `json:"birthday"`
	PhotoURL string `json:"photo,omitempty"`
	ActorURL string
	Movies   []Movie `json:"movies"`
}

type Movie struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	EngName   string `json:"enName"`
	PosterURL string `json:"photo"`
	MovieURL  string
	Rating    float32 `json:"rating"`
	Year      int     `json:"year"`
}

//type SessionState struct {
//	CorrelationID string
//	Step          string
//	FirstActorID  int
//	SecondActorID int
//}
