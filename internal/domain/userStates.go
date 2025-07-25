package domain

type SessionState struct {
	CorrelationID     string
	Step              string
	FirstActorID      int
	SecondActorID     int
	SentMediaMessages []int
	TempActors        []PhotoData
}

type PhotoData struct {
	ID       int
	PhotoURL string
	ActorURL string
	Caption  string
}
