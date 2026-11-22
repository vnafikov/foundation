package uploadingfileusecases

type DeleteExpiredParameters struct {
	TTLInterval   string
	LeaseInterval string
	BatchSize     int
}
