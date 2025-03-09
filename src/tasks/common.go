package tasks

type BackgroundTask interface {
	Name() string
	OnSchedule() error
}
