package domain

type Puppy struct {
	ID        int
	Name      string
	Title     string
	Sex       string
	Price     string
	ReadyOut  bool
	Archived  bool
	City      string
	MotherID  int
	FatherID  int
	DateBirth string
	Color     string
	Urls      []string
}

type Dog struct {
	ID       int
	Name     string
	Title    string
	Gender   string
	Color    string
	Archived bool
	Urls     []string
}

type Feedback struct {
	ID       int
	PuppyID  int
	Name     string
	Number   string
	Title    string
	Verified bool
	Date     string
	Urls     []string
}
