package domain

type CurrentOnCallSchedule struct {
	Schedules []Schedule
}

type Schedule struct {
	Name        string
	OnCallUsers string
}
